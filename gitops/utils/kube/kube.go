package k8s

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type K8s interface {
	ApplyResource(path string) error
	DeleteResource(path string) error
	GenerateManifests(path string) ([]*unstructured.Unstructured, error)
	GetResourceWithLabel(label map[string]string) ([]*unstructured.Unstructured, error)
	DiffResources(old []*unstructured.Unstructured, new []*unstructured.Unstructured) (bool, error)
	SetLabelsForResources(resources []*unstructured.Unstructured, labels map[string]string) error
}

type k8s struct {
	discoveryClient discovery.DiscoveryInterface
	dynClientSet    dynamic.Interface
}

func NewK8s(discoveryClient discovery.DiscoveryInterface, dynClientSet dynamic.Interface) *k8s {
	return &k8s{
		discoveryClient: discoveryClient,
		dynClientSet:    dynClientSet,
	}
}

func (k *k8s) ApplyResource(path string) error {
	return nil
}

func (k *k8s) DeleteResource(path string) error {
	return nil
}

func (k *k8s) GenerateManifests(path string) ([]*unstructured.Unstructured, error) {
	var objs []*unstructured.Unstructured

	// Get all file names in the directory
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read the content of the file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var unstructured unstructured.Unstructured
		dec := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(content)), 1000)
		if err := dec.Decode(&unstructured); err != nil {
			return nil
		}

		objs = append(objs, &unstructured)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return objs, nil
}

func (k *k8s) GetResourceWithLabel(label map[string]string) ([]*unstructured.Unstructured, error) {
	var objs []*unstructured.Unstructured
	var apiError error
	var wg sync.WaitGroup
	var lock sync.Mutex
	listOption := metav1.ListOptions{
		LabelSelector: labels.Set(label).String(),
	}

	if len(label) == 0 {
		return nil, fmt.Errorf("label is empty")
	}

	// Get the list of all API resources available
	serverResources, err := k.discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	wg.Add(len(serverResources))
	for _, group := range serverResources {
		go func(group *metav1.APIResourceList) {
			defer wg.Done()

			for _, resource := range group.APIResources {
				// Skip subresources like pod/logs, pod/status
				if containsSlash(resource.Name) {
					continue
				}
				gvr := schema.GroupVersionResource{
					Group:    group.GroupVersion,
					Version:  resource.Version,
					Resource: resource.Name,
				}
				if gvr.Group == "v1" {
					gvr.Version = gvr.Group
					gvr.Group = ""
				}

				var list *unstructured.UnstructuredList
				list, err = k.dynClientSet.Resource(gvr).List(context.TODO(), listOption)
				if err != nil {
					log.Warningf("Error listing resource %s: %s", gvr.String(), err)
					continue
				}

				// Append the resources to the list
				lock.Lock()
				for _, item := range list.Items {
					objs = append(objs, &item)
				}
				lock.Unlock()
			}
		}(group)
	}
	wg.Wait()

	return objs, apiError
}

func (k *k8s) DiffResources(current []*unstructured.Unstructured, new []*unstructured.Unstructured) (bool, error) {
	isChanged := false

	// Should use a cache to store current resources
	hashTable := make(map[string]*unstructured.Unstructured)
	for _, c := range current {
		hashTable[c.GetKind()+c.GetName()] = c
	}

	for _, n := range new {
		key := n.GetKind() + n.GetName()
		c, ok := hashTable[key]
		if !ok {
			log.Debugf("Found new resource %s with name %s", n.GetKind(), n.GetName())
			isChanged = true
			continue
		}

		if !reflect.DeepEqual(c.Object["spec"], n.Object["spec"]) {
			log.Debugf("Resource %s with name %s has changed", n.GetKind(), n.GetName())
			isChanged = true
		}
		delete(hashTable, key)
	}

	// Check if there are resources that need to be deleted
	if len(hashTable) > 0 {
		for _, c := range hashTable {
			log.Debugf("Resource %s with name %s should be deleted", c.GetKind(), c.GetName())
		}
		isChanged = true
	}

	return isChanged, nil
}

func (k *k8s) SetLabelsForResources(resources []*unstructured.Unstructured, labels map[string]string) error {
	for _, r := range resources {
		r.SetLabels(labels)
	}

	return nil
}

func containsSlash(s string) bool {
	return len(s) > 0 && s[0] == '/'
}
