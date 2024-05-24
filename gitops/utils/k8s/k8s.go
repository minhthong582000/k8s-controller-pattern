package k8s

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type K8s interface {
	ApplyResource(path string) error
	DeleteResource(path string) error
	GenerateManifests(path string) ([]unstructured.Unstructured, error)
	GetResourceWithLabel(label string) ([]unstructured.Unstructured, error)
	DiffResources(old []unstructured.Unstructured, new []unstructured.Unstructured) (bool, error)
}

type k8s struct {
}

func NewK8s() K8s {
	return &k8s{}
}

func (k *k8s) ApplyResource(path string) error {
	return nil
}

func (k *k8s) DeleteResource(path string) error {
	return nil
}

func (k *k8s) GenerateManifests(path string) ([]unstructured.Unstructured, error) {
	return nil, nil
}

func (k *k8s) GetResourceWithLabel(label string) ([]unstructured.Unstructured, error) {
	return nil, nil
}

func (k *k8s) DiffResources(old []unstructured.Unstructured, new []unstructured.Unstructured) (bool, error) {
	return true, nil
}
