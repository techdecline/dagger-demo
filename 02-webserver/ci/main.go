package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance/v2"
)

func main() {
	vars := []string{"AZURE_SUBSCRIPTION_ID", "AZURE_TENANT_ID", "AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET", "PULUMI_ACCESS_TOKEN"}
	for _, v := range vars {
		if os.Getenv(v) == "" {
			log.Fatalf("Environment variable %s is not set", v)
		}
	}
	releaseTag := os.Getenv("GITHUB_RUN_ID")

	//TODO: implement dynamic stack selection
	// Run Infra Pipeline and fetch stack outputs
	err, stackOutput := infra(context.Background(), vars, "dev")
	if err != nil {
		fmt.Println(err)
	}
	loginServer := get(stackOutput, "loginServer")
	// containerRegistryName := get(stackOutput, "containerRegistryName")
	registryPassword := get(stackOutput, "registryPassword")
	registryUsername := get(stackOutput, "registryUsername")
	resourceGroupName := get(stackOutput, "resourceGroupName")
	containerGroupName := get(stackOutput, "containerGroupName")
	azureLocation := get(stackOutput, "azureLocation")

	// Run Build Pipeline
	err, imageUrl := build(context.Background(), loginServer, registryUsername, registryPassword, releaseTag)
	if err != nil {
		fmt.Println(err)
	}

	// Run Deployment Pipeline
	if err := deploy(context.Background(), containerGroupName, resourceGroupName, azureLocation, imageUrl, "webserver"); err != nil {
		fmt.Println(err)
	}
}

func deploy(ctx context.Context, containerGroupName string, resourceGroupName string, location string, imageUrl string, containerName string) error {
	fmt.Println("App Deployment Pipeline")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// initialize Azure credentials
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	// initialize Azure client
	azureClient, err := armcontainerinstance.NewClientFactory(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return err
	}

	// define deployment request
	containerGroup := armcontainerinstance.ContainerGroup{
		Properties: &armcontainerinstance.ContainerGroupPropertiesProperties{
			Containers: []*armcontainerinstance.Container{
				{
					Name: to.Ptr(containerName),
					Properties: &armcontainerinstance.ContainerProperties{
						Command:              []*string{},
						EnvironmentVariables: []*armcontainerinstance.EnvironmentVariable{},
						Image:                to.Ptr(imageUrl),
						Ports: []*armcontainerinstance.ContainerPort{
							{
								Port: to.Ptr[int32](8080),
							}},
						Resources: &armcontainerinstance.ResourceRequirements{
							Requests: &armcontainerinstance.ResourceRequests{
								CPU:        to.Ptr[float64](1),
								MemoryInGB: to.Ptr[float64](1.5),
							},
						},
					},
				}},
			IPAddress: &armcontainerinstance.IPAddress{
				Type: to.Ptr(armcontainerinstance.ContainerGroupIPAddressTypePublic),
				Ports: []*armcontainerinstance.Port{
					{
						Port:     to.Ptr[int32](8080),
						Protocol: to.Ptr(armcontainerinstance.ContainerGroupNetworkProtocolTCP),
					}},
			},
			OSType:        to.Ptr(armcontainerinstance.OperatingSystemTypesLinux),
			RestartPolicy: to.Ptr(armcontainerinstance.ContainerGroupRestartPolicyAlways),
		},
		Location: to.Ptr(location),
	}

	poller, err := azureClient.NewContainerGroupsClient().BeginCreateOrUpdate(ctx, resourceGroupName, containerGroupName, containerGroup, nil)
	if err != nil {
		return err
	}

	// send request and wait until done
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}

func infra(ctx context.Context, environmentVariables []string, stackName string) (error, map[string]interface{}) {
	fmt.Println("Pulumi Deployment Pipeline")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err, nil
	}
	defer client.Close()

	// Get Current Working Directory
	cwd, err := os.Getwd()
	fmt.Printf("Current working dir: %s\n", cwd)
	if err != nil {
		log.Fatalf("Working Directory could not be found")
	}
	// get reference to the local project
	src := client.Host().Directory(filepath.Dir(cwd) + "/infra")

	// get `terraform` image
	pulumi := client.Container().From("pulumi/pulumi-go:latest")

	// mount cloned repository into `pulumi` image
	pulumi = pulumi.WithDirectory("/src", src).WithWorkdir("/src")

	// set Environment Variables in Container
	for _, v := range environmentVariables {
		pulumi = pulumi.WithEnvVariable(strings.Replace(v, "AZURE", "ARM", 1), os.Getenv(v))
	}

	// Select Stack
	pulumi = pulumi.WithExec([]string{"pulumi", "stack", "select", stackName})

	// Apply pulumi stack
	pulumi = pulumi.WithExec([]string{"pulumi", "up", "-y", "--skip-preview"})

	// Read Stack Outputs
	var x map[string]interface{}
	stdout, err := pulumi.WithExec([]string{"pulumi", "stack", "output", "-j"}).Stdout(ctx)
	fmt.Println(stdout)
	if err != nil {
		return err, nil
	}
	json.Unmarshal([]byte(stdout), &x)

	return nil, x
}

func build(ctx context.Context, registryLoginServer string, registryUsername string, registryPassword string, tag string) (error, string) {
	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err, ""
	}
	defer client.Close()

	// Get Current Working Directory
	cwd, err := os.Getwd()
	fmt.Printf("Current working dir: %s\n", cwd)
	if err != nil {
		log.Fatalf("Working Directory could not be found")
	}
	// get reference to the local project
	src := client.Host().Directory(filepath.Dir(cwd) + "/app")

	// build app
	builder := client.Container().
		From("golang:latest").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", "webserver"})

		// publish binary on alpine base
	prodImage := client.Container().
		From("alpine").
		WithFile("/bin/webserver", builder.File("/src/webserver")).
		WithEntrypoint([]string{"/bin/webserver"})

		// publish image
	secretPassword := client.SetSecret("registryCred", registryPassword)
	imageRef := fmt.Sprintf("%s/webserver", registryLoginServer)
	if tag != "" {
		imageRef = imageRef + fmt.Sprintf(":%s", tag)
	}
	fmt.Println(imageRef)
	imageUrl, err := prodImage.WithRegistryAuth(registryLoginServer, registryUsername, secretPassword).
		Publish(ctx, imageRef)
	if err != nil {
		log.Fatalf("Image Upload failed to %s: %s", registryLoginServer, err)
		return err, ""
	}

	return nil, imageUrl
}

// get is a function that takes a map[string]interface{} and a key string and returns the value associated with the key or nil if not found
func get(m map[string]interface{}, key string) string {
	// check if the map contains the key
	if value, ok := m[key]; ok {
		// check if the value is a string
		if str, ok := value.(string); ok {
			// return the string
			return str
		}
	}
	// return an empty string otherwise
	return ""
}
