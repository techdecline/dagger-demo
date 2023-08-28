# Dagger Demonstration for Infrastructure as Code Pipelines

## Preparation Golang

see [Dagger.io Documentation for Golang](https://docs.dagger.io/sdk/go/371491/install)

* install Dagger SDK
    `go get dagger.io/dagger@latest` or call `go mod tidy` from your ci module

## General Notes

* When developing in a Dev Container, Docker in Docker is required. Same goes for Container-based CI/CD-Workers
* In Example 01, Backend Configuration needs to be updated
* Examples are implemented with GitHub Actions as CI/CD system. Variables for local Execution will need to be set up manually
