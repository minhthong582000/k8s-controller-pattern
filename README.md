# K8s controller pattern

The project includes Go code to interact with the Kubernetes API, aiming to create a complete Kubernetes controller.

The code is organized into the following directories, click to navigate:

1. [client-go](./client-go/): A very basic example of using client-go to interact with the Kubernetes API, including listing and watching resources.
2. [api-machinery](./api-machinery/README.md): A more complete example of a Kubernetes controller using the controller-runtime library, including custom resource definitions (CRDs), reconciliation logic, and event handling.
3. [example-controller](./example-controller/README.md): A simple example of a Kubernetes controller implementation.
4. [gitops-controller](./gitops/README.md): An example of a GitOps-style controller for managing Kubernetes resources.

## Prerequisites

- A Kubernetes cluster with kubectl configured
- Go 1.22 or later
- [controller-gen](https://github.com/kubernetes-sigs/controller-tools/tree/master/cmd/controller-gen)
- [code-generator](https://github.com/kubernetes/code-generator)
