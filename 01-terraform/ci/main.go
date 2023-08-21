package main

import (
	"context"
	"fmt"
	"os"
	"log"

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
		terraform = terraform.WithEnvVariable("v", os.Getenv(v))
	}

	// run init
    terraform = terraform.WithExec([]string{"init"})

	// define the application build command
    path := "build/tfplan"
    terraform = terraform.WithExec([]string{"plan", "-out", path})

    // get reference to build output directory in container
    output := terraform.Directory(path)

    // write contents of container build/ directory to the host
    _, err = output.Export(ctx, path)
    if err != nil {
        return err
    }

	return nil
}
