# Dagger Demonstration for Infrastructure as Code Pipelines

## Preparation Golang

see [Dagger.io Documentation for Golang](https://docs.dagger.io/sdk/go/371491/install)

* install Dagger SDK
    `go get dagger.io/dagger@latest` or call `go mod tidy` from your ci module

## Demos

### 00-helloworld

This pipeline initializes the Dagger Engine and prints a string to the commandline.

### 01-terraform

This pipeline can plan and apply a Terraform configuration. A mode selector implemented as commandline parameter is used to select execution mode (Plan or Apply).

### 02-webserver

In this example, a Web Server implemented in Golang will be packaged as a container and updated to an Azure Container Registry and rolled out to an Azure Container Instance. Required infrastructure will be rollout by a Pulumi Program also part of the repo.

## General Notes

* When developing in a Dev Container, Docker in Docker is required. Same goes for Container-based CI/CD-Workers
* In Example 01, Backend Configuration needs to be updated
* Examples are implemented with GitHub Actions as CI/CD system. Variables for local Execution will need to be set up manually
