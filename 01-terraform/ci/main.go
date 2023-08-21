package main

import (
	"context"
	"fmt"
	"os"
	"log"
	"strings"

	"dagger.io/dagger"
)

func main() {
	if err := build(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func build(ctx context.Context) error {
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

	// get reference to the local project
    src := client.Host().Directory("/home/decline/dagger-demo/01-terraform")

    // get `terraform` image
    terraform := client.Container().From("hashicorp/terraform:latest")

    // mount cloned repository into `terraform` image
    terraform = terraform.WithDirectory("/src", src).WithWorkdir("/src")

	// set Environment Variables in Container
	for _,v := range vars {
		terraform = terraform.WithEnvVariable(strings.Replace(v,"AZURE","ARM",1), os.Getenv(v))
	}

	// define the Terraform Init Command
    terraform = terraform.WithExec([]string{"init"})

	// define the Terraform Plan Command
	path := "/src"
    terraform = terraform.WithExec([]string{"plan","-out","tfplan"})

    // get reference to build output directory in container
    output := terraform.Directory(path)

    // write contents of container build/ directory to the host
    _, err = output.Export(ctx, path)
    if err != nil {
        return err
    }

	return nil
}
