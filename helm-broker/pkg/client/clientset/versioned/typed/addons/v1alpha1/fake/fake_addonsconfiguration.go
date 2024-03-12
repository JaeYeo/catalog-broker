/*
Copyright 2020 The Helm Broker Authors.

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

package fake

import (
	"context"

	v1alpha1 "github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAddonsConfigurations implements AddonsConfigurationInterface
type FakeAddonsConfigurations struct {
	Fake *FakeAddonsV1alpha1
	ns   string
}

var addonsconfigurationsResource = schema.GroupVersionResource{Group: "addons.kyma-project.io", Version: "v1alpha1", Resource: "addonsconfigurations"}

var addonsconfigurationsKind = schema.GroupVersionKind{Group: "addons.kyma-project.io", Version: "v1alpha1", Kind: "AddonsConfiguration"}

// Get takes name of the addonsConfiguration, and returns the corresponding addonsConfiguration object, and an error if there is any.
func (c *FakeAddonsConfigurations) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.AddonsConfiguration, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(addonsconfigurationsResource, c.ns, name), &v1alpha1.AddonsConfiguration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AddonsConfiguration), err
}

// List takes label and field selectors, and returns the list of AddonsConfigurations that match those selectors.
func (c *FakeAddonsConfigurations) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.AddonsConfigurationList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(addonsconfigurationsResource, addonsconfigurationsKind, c.ns, opts), &v1alpha1.AddonsConfigurationList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AddonsConfigurationList{ListMeta: obj.(*v1alpha1.AddonsConfigurationList).ListMeta}
	for _, item := range obj.(*v1alpha1.AddonsConfigurationList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested addonsConfigurations.
func (c *FakeAddonsConfigurations) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(addonsconfigurationsResource, c.ns, opts))

}

// Create takes the representation of a addonsConfiguration and creates it.  Returns the server's representation of the addonsConfiguration, and an error, if there is any.
func (c *FakeAddonsConfigurations) Create(ctx context.Context, addonsConfiguration *v1alpha1.AddonsConfiguration, opts v1.CreateOptions) (result *v1alpha1.AddonsConfiguration, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(addonsconfigurationsResource, c.ns, addonsConfiguration), &v1alpha1.AddonsConfiguration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AddonsConfiguration), err
}

// Update takes the representation of a addonsConfiguration and updates it. Returns the server's representation of the addonsConfiguration, and an error, if there is any.
func (c *FakeAddonsConfigurations) Update(ctx context.Context, addonsConfiguration *v1alpha1.AddonsConfiguration, opts v1.UpdateOptions) (result *v1alpha1.AddonsConfiguration, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(addonsconfigurationsResource, c.ns, addonsConfiguration), &v1alpha1.AddonsConfiguration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AddonsConfiguration), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeAddonsConfigurations) UpdateStatus(ctx context.Context, addonsConfiguration *v1alpha1.AddonsConfiguration, opts v1.UpdateOptions) (*v1alpha1.AddonsConfiguration, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(addonsconfigurationsResource, "status", c.ns, addonsConfiguration), &v1alpha1.AddonsConfiguration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AddonsConfiguration), err
}

// Delete takes name of the addonsConfiguration and deletes it. Returns an error if one occurs.
func (c *FakeAddonsConfigurations) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(addonsconfigurationsResource, c.ns, name), &v1alpha1.AddonsConfiguration{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAddonsConfigurations) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(addonsconfigurationsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.AddonsConfigurationList{})
	return err
}

// Patch applies the patch and returns the patched addonsConfiguration.
func (c *FakeAddonsConfigurations) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.AddonsConfiguration, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(addonsconfigurationsResource, c.ns, name, pt, data, subresources...), &v1alpha1.AddonsConfiguration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AddonsConfiguration), err
}
