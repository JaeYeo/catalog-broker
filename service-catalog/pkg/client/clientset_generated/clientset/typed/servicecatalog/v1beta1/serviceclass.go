/*
Copyright 2023 The Kubernetes Authors.

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

package v1beta1

import (
	"context"
	"time"

	v1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scheme "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ServiceClassesGetter has a method to return a ServiceClassInterface.
// A group's client should implement this interface.
type ServiceClassesGetter interface {
	ServiceClasses(namespace string) ServiceClassInterface
}

// ServiceClassInterface has methods to work with ServiceClass resources.
type ServiceClassInterface interface {
	Create(ctx context.Context, serviceClass *v1beta1.ServiceClass, opts v1.CreateOptions) (*v1beta1.ServiceClass, error)
	Update(ctx context.Context, serviceClass *v1beta1.ServiceClass, opts v1.UpdateOptions) (*v1beta1.ServiceClass, error)
	UpdateStatus(ctx context.Context, serviceClass *v1beta1.ServiceClass, opts v1.UpdateOptions) (*v1beta1.ServiceClass, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.ServiceClass, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.ServiceClassList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.ServiceClass, err error)
	ServiceClassExpansion
}

// serviceClasses implements ServiceClassInterface
type serviceClasses struct {
	client rest.Interface
	ns     string
}

// newServiceClasses returns a ServiceClasses
func newServiceClasses(c *ServicecatalogV1beta1Client, namespace string) *serviceClasses {
	return &serviceClasses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the serviceClass, and returns the corresponding serviceClass object, and an error if there is any.
func (c *serviceClasses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.ServiceClass, err error) {
	result = &v1beta1.ServiceClass{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("serviceclasses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceClasses that match those selectors.
func (c *serviceClasses) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.ServiceClassList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.ServiceClassList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("serviceclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceClasses.
func (c *serviceClasses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("serviceclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a serviceClass and creates it.  Returns the server's representation of the serviceClass, and an error, if there is any.
func (c *serviceClasses) Create(ctx context.Context, serviceClass *v1beta1.ServiceClass, opts v1.CreateOptions) (result *v1beta1.ServiceClass, err error) {
	result = &v1beta1.ServiceClass{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("serviceclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(serviceClass).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a serviceClass and updates it. Returns the server's representation of the serviceClass, and an error, if there is any.
func (c *serviceClasses) Update(ctx context.Context, serviceClass *v1beta1.ServiceClass, opts v1.UpdateOptions) (result *v1beta1.ServiceClass, err error) {
	result = &v1beta1.ServiceClass{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("serviceclasses").
		Name(serviceClass.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(serviceClass).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *serviceClasses) UpdateStatus(ctx context.Context, serviceClass *v1beta1.ServiceClass, opts v1.UpdateOptions) (result *v1beta1.ServiceClass, err error) {
	result = &v1beta1.ServiceClass{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("serviceclasses").
		Name(serviceClass.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(serviceClass).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the serviceClass and deletes it. Returns an error if one occurs.
func (c *serviceClasses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("serviceclasses").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceClasses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("serviceclasses").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched serviceClass.
func (c *serviceClasses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.ServiceClass, err error) {
	result = &v1beta1.ServiceClass{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("serviceclasses").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
