# Gitops

This controller does the same job as ArgoCD, which is a k8s Continuous Deployment controller following the GitOps pattern.

## How to build ?

### Generate code using code-generator

```bash
export CODEGEN_PKG=../../code-generator
./hack/update-codegen.sh
```

### Generate CRD manifests

```bash
./hack/controller-gen.sh
```

## How to run ?

Apply the CRD manifests:

```bash
kubectl apply -Rf deploy/crds
```

Apply example CR:

```bash
kubectl apply -Rf deploy/example
```

Run the controller:

```bash
go run main.go run \
    -k ~/.kube/config \
    -l info
```

Or you can run the controller in Kubernetes:

```bash
kubectl apply -Rf deploy
```
