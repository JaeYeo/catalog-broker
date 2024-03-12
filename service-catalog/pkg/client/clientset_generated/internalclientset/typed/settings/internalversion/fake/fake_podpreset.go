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

package fake

import (
	"context"

	settings "github.com/kubernetes-sigs/service-catalog/pkg/apis/settings"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePodPresets implements PodPresetInterface
type FakePodPresets struct {
	Fake *FakeSettings
	ns   string
}

var podpresetsResource = settings.SchemeGroupVersion.WithResource("podpresets")

var podpresetsKind = settings.SchemeGroupVersion.WithKind("PodPreset")

// Get takes name of the podPreset, and returns the corresponding podPreset object, and an error if there is any.
func (c *FakePodPresets) Get(ctx context.Context, name string, options v1.GetOptions) (result *settings.PodPreset, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podpresetsResource, c.ns, name), &settings.PodPreset{})

	if obj == nil {
		return nil, err
	}
	return obj.(*settings.PodPreset), err
}

// List takes label and field selectors, and returns the list of PodPresets that match those selectors.
func (c *FakePodPresets) List(ctx context.Context, opts v1.ListOptions) (result *settings.PodPresetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podpresetsResource, podpresetsKind, c.ns, opts), &settings.PodPresetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &settings.PodPresetList{ListMeta: obj.(*settings.PodPresetList).ListMeta}
	for _, item := range obj.(*settings.PodPresetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested podPresets.
func (c *FakePodPresets) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podpresetsResource, c.ns, opts))

}

// Create takes the representation of a podPreset and creates it.  Returns the server's representation of the podPreset, and an error, if there is any.
func (c *FakePodPresets) Create(ctx context.Context, podPreset *settings.PodPreset, opts v1.CreateOptions) (result *settings.PodPreset, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podpresetsResource, c.ns, podPreset), &settings.PodPreset{})

	if obj == nil {
		return nil, err
	}
	return obj.(*settings.PodPreset), err
}

// Update takes the representation of a podPreset and updates it. Returns the server's representation of the podPreset, and an error, if there is any.
func (c *FakePodPresets) Update(ctx context.Context, podPreset *settings.PodPreset, opts v1.UpdateOptions) (result *settings.PodPreset, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podpresetsResource, c.ns, podPreset), &settings.PodPreset{})

	if obj == nil {
		return nil, err
	}
	return obj.(*settings.PodPreset), err
}

// Delete takes name of the podPreset and deletes it. Returns an error if one occurs.
func (c *FakePodPresets) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(podpresetsResource, c.ns, name, opts), &settings.PodPreset{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePodPresets) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podpresetsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &settings.PodPresetList{})
	return err
}

// Patch applies the patch and returns the patched podPreset.
func (c *FakePodPresets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *settings.PodPreset, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podpresetsResource, c.ns, name, pt, data, subresources...), &settings.PodPreset{})

	if obj == nil {
		return nil, err
	}
	return obj.(*settings.PodPreset), err
}
