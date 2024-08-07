# Example Controller

A simple example of a Kubernetes controller written in Golang.

What it does is whenever a new Deployment is created in "default" namespace, it will expose the Deployment by creating a Service and an Ingress.

![Kubernetes Controller Diagram](./docs/k8s-controller.drawio.svg)

See also: [client-go under the hood](https://github.com/kubernetes/sample-controller/blob/master/docs/controller-client-go.md)

## How to run?

```bash
go build -o bin/example-controller

bin/example-controller \
    --kubeconfig=$HOME/.kube/config \
    --workers 2 \
    --namespace default
```

Open another terminal and create a Deployment:

```bash
kubectl create -f ./example/deployment.yaml
```

Watch the logs of the controller.
