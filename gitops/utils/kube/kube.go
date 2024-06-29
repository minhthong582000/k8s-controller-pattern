package k8s

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type K8s interface {
	CreateResource(ctx context.Context, obj *unstructured.Unstructured, namespace string) error
	PatchResource(ctx context.Context, currentObj *unstructured.Unstructured, namespace string) error
	DeleteResource(ctx context.Context, currentObj *unstructured.Unstructured, namespace string) error
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

func (k *k8s) CreateResource(ctx context.Context, obj *unstructured.Unstructured, namespace string) error {
	gvk := obj.GroupVersionKind()
	apiResource, err := ServerResourceForGroupVersionKind(
		k.discoveryClient,
		gvk,
		"create",
	)
	if err != nil {
		return err
	}

	resource := gvk.GroupVersion().WithResource(apiResource.Name)

	var dynInterface dynamic.ResourceInterface = k.dynClientSet.Resource(resource)
	if apiResource.Namespaced {
		dynInterface = k.dynClientSet.Resource(resource).Namespace(namespace)
	}
	_, err = dynInterface.Create(ctx, obj, metav1.CreateOptions{})
	return err
}

func (k *k8s) PatchResource(ctx context.Context, currentObj *unstructured.Unstructured, namespace string) error {
	gvk := currentObj.GroupVersionKind()
	apiResource, err := ServerResourceForGroupVersionKind(
		k.discoveryClient,
		gvk,
		"patch",
	)
	if err != nil {
		return err
	}

	resource := gvk.GroupVersion().WithResource(apiResource.Name)

	var dynInterface dynamic.ResourceInterface = k.dynClientSet.Resource(resource)
	if apiResource.Namespaced {
		dynInterface = k.dynClientSet.Resource(resource).Namespace(namespace)
	}
	outBytes, err := runtime.Encode(unstructured.UnstructuredJSONScheme, currentObj)
	if err != nil {
		return err
	}
	_, err = dynInterface.Patch(
		ctx,
		currentObj.GetName(),
		types.ApplyPatchType,
		outBytes,
		metav1.PatchOptions{},
	)
	return err
}

func (k *k8s) DeleteResource(ctx context.Context, currentObj *unstructured.Unstructured, namespace string) error {
	gvk := currentObj.GroupVersionKind()
	apiResource, err := ServerResourceForGroupVersionKind(
		k.discoveryClient,
		gvk,
		"delete",
	)
	if err != nil {
		return err
	}

	resource := gvk.GroupVersion().WithResource(apiResource.Name)

	var dynInterface dynamic.ResourceInterface = k.dynClientSet.Resource(resource)
	if apiResource.Namespaced {
		dynInterface = k.dynClientSet.Resource(resource).Namespace(namespace)
	}
	return dynInterface.Delete(ctx, currentObj.GetName(), metav1.DeleteOptions{})
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
				gv, err := schema.ParseGroupVersion(group.GroupVersion)
				if err != nil {
					log.Warningf("parsing GroupVersion %s failed: %s", group.GroupVersion, err)
				}
				gvr := schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource.Name,
				}

				var list *unstructured.UnstructuredList
				list, err = k.dynClientSet.Resource(gvr).List(context.TODO(), listOption)
				if err != nil {
					log.Warningf("Error listing resource %s, %s", gvr.String(), err)
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

func ServerResourceForGroupVersionKind(disco discovery.DiscoveryInterface, gvk schema.GroupVersionKind, verb string) (*metav1.APIResource, error) {
	// default is to return a not found for the requested resource
	retErr := apierr.NewNotFound(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, "")
	resources, err := disco.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, err
	}
	for _, r := range resources.APIResources {
		if r.Kind == gvk.Kind {
			if isSupportedVerb(&r, verb) {
				return &r, nil
			} else {
				// We have a match, but the API does not support the action
				// that was requested. Memorize this.
				retErr = apierr.NewMethodNotSupported(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, verb)
			}
		}
	}
	return nil, retErr
}

// isSupportedVerb returns whether or not a APIResource supports a specific verb.
// The verb will be matched case-insensitive.
func isSupportedVerb(apiResource *metav1.APIResource, verb string) bool {
	if verb == "" || verb == "*" {
		return true
	}
	for _, v := range apiResource.Verbs {
		if strings.EqualFold(v, verb) {
			return true
		}
	}
	return false
}

func ToResourceInterface(dynamicIf dynamic.Interface, apiResource *metav1.APIResource, resource schema.GroupVersionResource, namespace string) dynamic.ResourceInterface {
	if apiResource.Namespaced {
		return dynamicIf.Resource(resource).Namespace(namespace)
	}
	return dynamicIf.Resource(resource)
}
