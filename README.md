# K8s controller pattern

The project includes Go code to interact with the Kubernetes API, aiming to create a complete Kubernetes controller.

Includes:

1. [client-go](./client-go/)
2. [api-machinery](./api-machinery/README.md)
3. [example-controller](./example-controller/README.md)
4. [gitops-controller](./gitops/README.md)

## Prerequisites

- A Kubernetes cluster with kubectl configured
- Go 1.22 or later
- [controller-gen](https://github.com/kubernetes-sigs/controller-tools/tree/master/cmd/controller-gen)
- [code-generator](https://github.com/kubernetes/code-generator)
