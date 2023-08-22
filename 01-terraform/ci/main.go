package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

func main() {
	var mode string

	flag.StringVar(&mode, "m", "plan", "Specify Terraform Operation Mode. Default is plan")
	flag.Parse()

	switch mode {
	case "plan":
		fmt.Printf("Selected Mode is: %s\n", mode)
		if err := plan(context.Background()); err != nil {
			fmt.Println(err)
		}
	case "apply":
		fmt.Printf("Selected Mode is: %s\n", mode)
		if err := apply(context.Background()); err != nil {
			fmt.Println(err)
		}
	default:
		log.Fatalf("Selected Mode %s is not implemented", mode)
	}

}

func plan(ctx context.Context) error {
	fmt.Println("Terraform Plan Pipeline")

	vars := []string{"AZURE_SUBSCRIPTION_ID", "AZURE_TENANT_ID", "AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET"}
	for _, v := range vars {
		if os.Getenv(v) == "" {
			log.Fatalf("Environment variable %s is not set", v)
		}
	}

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// Get Current Working Directory
	cwd, err := os.Getwd()
	fmt.Printf("Current working dir: %s\n", cwd)
	if err != nil {
		log.Fatalf("Working Directory could not be found")
	}
	// get reference to the local project
	src := client.Host().Directory(filepath.Dir(cwd))

	// get `terraform` image
	terraform := client.Container().From("hashicorp/terraform:latest")

	// mount cloned repository into `terraform` image
	terraform = terraform.WithDirectory("/src", src).WithWorkdir("/src")

	// create output folder on runner
	planPath := cwd + "/plan"
	err = os.Mkdir(planPath, os.ModePerm)
	if err != nil {
		log.Fatalf("Output folder %s could not be create", planPath)
	}

	// Mount Plan folder to container
	planDirectory := client.Host().Directory(planPath)
	terraform = terraform.WithDirectory("/plan", planDirectory)

	// set Environment Variables in Container
	for _, v := range vars {
		terraform = terraform.WithEnvVariable(strings.Replace(v, "AZURE", "ARM", 1), os.Getenv(v))
	}

	// define the Terraform Init Command
	terraform = terraform.WithExec([]string{"init"})

	// define the Terraform Plan Command
	path := "/plan/tfplan"
	terraform = terraform.WithExec([]string{"plan", "-out", path})

	// get reference to build output directory in container
	output := terraform.Directory(filepath.Dir(path))

	// write contents of container build/ directory to the host
	_, err = output.Export(ctx, planPath)
	if err != nil {
		return err
	}
	fmt.Printf("Terraform Plan Dirctory: %s\n", path)

	return nil
}

func apply(ctx context.Context) error {
	fmt.Println("Terraform Apply Pipeline")

	vars := []string{"AZURE_SUBSCRIPTION_ID", "AZURE_TENANT_ID", "AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET"}
	for _, v := range vars {
		if os.Getenv(v) == "" {
			log.Fatalf("Environment variable %s is not set", v)
		}
	}

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// Get Current Working Directory
	cwd, err := os.Getwd()
	fmt.Printf("Current working dir: %s\n", cwd)
	if err != nil {
		log.Fatalf("Working Directory could not be found")
	}
	// get reference to the local project
	src := client.Host().Directory(filepath.Dir(cwd))

	// get `terraform` image
	terraform := client.Container().From("hashicorp/terraform:latest")

	// mount cloned repository into `terraform` image
	terraform = terraform.WithDirectory("/src", src).WithWorkdir("/src")

	// create output folder on runner
	planPath := cwd + "/plan/tfplan"
	exists, err := exists(planPath)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Fatal("Plan File Path does not exist")
	}

	// Mount Plan folder to container
	planDirectory := client.Host().Directory(filepath.Dir(planPath))
	terraform = terraform.WithDirectory("/plan", planDirectory)

	// set Environment Variables in Container
	for _, v := range vars {
		terraform = terraform.WithEnvVariable(strings.Replace(v, "AZURE", "ARM", 1), os.Getenv(v))
	}

	// define the Terraform Init Command
	terraform = terraform.WithExec([]string{"init"})

	// define the Terraform Apply Command
	path := "/plan/tfplan"
	terraform = terraform.WithExec([]string{"apply", "-auto-approve", path})

	// get reference to build output directory in container
	output := terraform.Directory(filepath.Dir(path))

	// write contents of container build/ directory to the host
	_, err = output.Export(ctx, filepath.Dir(planPath))
	if err != nil {
		return err
	}

	return nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
