# Dagger Demonstration for Infrastructure as Code Pipelines

## Preparation TypeScript

see [Dagger.io Documentation for Typescript](https://docs.dagger.io/sdk/nodejs/835948/install)

* install NPM
    `npm install '@dagger.io/dagger@latest' --save-dev`
* install TypeScript Engine
    `npm install ts-node typescript`
* create TypeScript configuration file
    `npx tsc --init --module esnext --moduleResolution nodenext`


## General Notes

* When developing in a Dev Container, Docker in Docker is required.
* Everything is cached by default