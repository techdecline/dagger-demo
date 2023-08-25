package main

import (
	"github.com/pulumi/pulumi-azure-native-sdk/containerinstance"
	"github.com/pulumi/pulumi-azure-native-sdk/containerregistry"
	"github.com/pulumi/pulumi-azure-native-sdk/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(ctx, "rg-dagger-go-webserver", nil)
		if err != nil {
			return err
		}

		// Create an Azure Container Registry
		registry, err := containerregistry.NewRegistry(ctx, "acr-dagger-go-webserver", &containerregistry.RegistryArgs{
			AdminUserEnabled:  pulumi.Bool(true),
			ResourceGroupName: resourceGroup.Name,
			RegistryName:      pulumi.String("acrdaggerdemo"),
			Sku: &containerregistry.SkuArgs{
				Name: pulumi.String("Premium"),
			},
		})
		if err != nil {
			return err
		}

		// // Get the registry credentials
		credentials := pulumi.All(resourceGroup.Name, registry.Name).ApplyT(
			func(args []interface{}) (*containerregistry.ListRegistryCredentialsResult, error) {
				resourceGroupName := args[0].(string)
				registryName := args[1].(string)
				return containerregistry.ListRegistryCredentials(ctx, &containerregistry.ListRegistryCredentialsArgs{
					ResourceGroupName: resourceGroupName,
					RegistryName:      registryName,
				})
			},
		)

		adminUsername := credentials.ApplyT(func(result interface{}) (string, error) {
			credentials := result.(*containerregistry.ListRegistryCredentialsResult)
			return *credentials.Username, nil
		}).(pulumi.StringOutput)

		adminPassword := credentials.ApplyT(func(result interface{}) (string, error) {
			credentials := result.(*containerregistry.ListRegistryCredentialsResult)
			return *credentials.Passwords[0].Value, nil
		}).(pulumi.StringOutput)

		// Create an Azure Container Group with a single container instance
		containerGroup, err := containerinstance.NewContainerGroup(ctx, "myContainerGroup", &containerinstance.ContainerGroupArgs{
			ResourceGroupName:  resourceGroup.Name,
			ContainerGroupName: pulumi.String("helloworld"),
			Containers: containerinstance.ContainerArray{
				&containerinstance.ContainerArgs{
					Name:  pulumi.String("hw"),
					Image: pulumi.String("mcr.microsoft.com/azuredocs/aci-helloworld"),
					Resources: &containerinstance.ResourceRequirementsArgs{
						Requests: &containerinstance.ResourceRequestsArgs{
							Cpu:        pulumi.Float64(1.0),
							MemoryInGB: pulumi.Float64(1.5),
						},
					},
				},
			},
			OsType: pulumi.String("Linux"),
			// IpAddressType: pulumi.String("public"),
			Location: resourceGroup.Location,
		}, pulumi.IgnoreChanges([]string{"containers"})) // Here we ignore changes to 'containers' property)
		if err != nil {
			return err
		}

		ctx.Export("resourceGroupName", resourceGroup.Name)
		ctx.Export("containerRegistryName", registry.Name)
		ctx.Export("registryUsername", adminUsername)
		ctx.Export("registryPassword", adminPassword)
		ctx.Export("loginServer", registry.LoginServer)
		ctx.Export("containerGroupName", containerGroup.Name)
		ctx.Export("azureLocation", resourceGroup.Location)

		return nil
	})
}
