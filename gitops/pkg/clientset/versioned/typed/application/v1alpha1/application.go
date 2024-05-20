/*
Copyright thongdepzai-cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/apis/application/v1alpha1"
	scheme "github.com/minhthong582000/k8s-controller-pattern/gitops/pkg/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ApplicationsGetter has a method to return a ApplicationInterface.
// A group's client should implement this interface.
type ApplicationsGetter interface {
	Applications(namespace string) ApplicationInterface
}

// ApplicationInterface has methods to work with Application resources.
type ApplicationInterface interface {
	Create(ctx context.Context, application *v1alpha1.Application, opts v1.CreateOptions) (*v1alpha1.Application, error)
	Update(ctx context.Context, application *v1alpha1.Application, opts v1.UpdateOptions) (*v1alpha1.Application, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ApplicationList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Application, err error)
	ApplicationExpansion
}

// applications implements ApplicationInterface
type applications struct {
	client rest.Interface
	ns     string
}

// newApplications returns a Applications
func newApplications(c *ThongdepzaiV1alpha1Client, namespace string) *applications {
	return &applications{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the application, and returns the corresponding application object, and an error if there is any.
func (c *applications) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Application, err error) {
	result = &v1alpha1.Application{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("applications").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Applications that match those selectors.
func (c *applications) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ApplicationList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ApplicationList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("applications").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested applications.
func (c *applications) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("applications").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a application and creates it.  Returns the server's representation of the application, and an error, if there is any.
func (c *applications) Create(ctx context.Context, application *v1alpha1.Application, opts v1.CreateOptions) (result *v1alpha1.Application, err error) {
	result = &v1alpha1.Application{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("applications").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(application).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a application and updates it. Returns the server's representation of the application, and an error, if there is any.
func (c *applications) Update(ctx context.Context, application *v1alpha1.Application, opts v1.UpdateOptions) (result *v1alpha1.Application, err error) {
	result = &v1alpha1.Application{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("applications").
		Name(application.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(application).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the application and deletes it. Returns an error if one occurs.
func (c *applications) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("applications").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *applications) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("applications").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched application.
func (c *applications) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Application, err error) {
	result = &v1alpha1.Application{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("applications").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
