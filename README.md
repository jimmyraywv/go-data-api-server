# go-data-api-server
Simple Golang data api using:
- API routing - [gorilla/mux](https://github.com/gorilla/mux)
- Structured logging - [logrus](https://github.com/sirupsen/logrus)
- JSON validation - [encoding/json](https://pkg.go.dev/encoding/json) & [go-playground/validator](https://github.com/go-playground/validator)

Most tasks are automated with a `Makefile`. Run `make help` for Makefile help.
```
Usage:
  make <target>

General
  help             Display this help.
  meta             Provides metadata for other commands; good for DevOps logging. Can be called as a target, but is mostly used by other targets as a dependency.

Build and deploy
  build            Build container with Docker buildx, based on PLATFORM argument (default linux/amd64)
  login            Login to remote image registry
  push             Push to remote image registry
  pull             Pull from remote image registry

Local Development
  compile          Compile for local MacOS
  clean            Remove compile binary
  init             Initialize Go project
  check            Vet and Lint Go codebase
  test             Run tests
  run              Run local binary
```

## Purpose
Created to test API use cases in Kubernetes, with associated network plumbing in AWS.

## WARNING: Not production ready!
