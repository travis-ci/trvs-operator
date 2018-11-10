/*
Copyright The Kubernetes Authors.

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
	travisciv1 "github.com/travis-ci/trvs-operator/pkg/apis/travisci/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTrvsSecrets implements TrvsSecretInterface
type FakeTrvsSecrets struct {
	Fake *FakeTravisciV1
	ns   string
}

var trvssecretsResource = schema.GroupVersionResource{Group: "travisci.com", Version: "v1", Resource: "trvssecrets"}

var trvssecretsKind = schema.GroupVersionKind{Group: "travisci.com", Version: "v1", Kind: "TrvsSecret"}

// Get takes name of the trvsSecret, and returns the corresponding trvsSecret object, and an error if there is any.
func (c *FakeTrvsSecrets) Get(name string, options v1.GetOptions) (result *travisciv1.TrvsSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(trvssecretsResource, c.ns, name), &travisciv1.TrvsSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*travisciv1.TrvsSecret), err
}

// List takes label and field selectors, and returns the list of TrvsSecrets that match those selectors.
func (c *FakeTrvsSecrets) List(opts v1.ListOptions) (result *travisciv1.TrvsSecretList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(trvssecretsResource, trvssecretsKind, c.ns, opts), &travisciv1.TrvsSecretList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &travisciv1.TrvsSecretList{ListMeta: obj.(*travisciv1.TrvsSecretList).ListMeta}
	for _, item := range obj.(*travisciv1.TrvsSecretList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested trvsSecrets.
func (c *FakeTrvsSecrets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(trvssecretsResource, c.ns, opts))

}

// Create takes the representation of a trvsSecret and creates it.  Returns the server's representation of the trvsSecret, and an error, if there is any.
func (c *FakeTrvsSecrets) Create(trvsSecret *travisciv1.TrvsSecret) (result *travisciv1.TrvsSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(trvssecretsResource, c.ns, trvsSecret), &travisciv1.TrvsSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*travisciv1.TrvsSecret), err
}

// Update takes the representation of a trvsSecret and updates it. Returns the server's representation of the trvsSecret, and an error, if there is any.
func (c *FakeTrvsSecrets) Update(trvsSecret *travisciv1.TrvsSecret) (result *travisciv1.TrvsSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(trvssecretsResource, c.ns, trvsSecret), &travisciv1.TrvsSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*travisciv1.TrvsSecret), err
}

// Delete takes name of the trvsSecret and deletes it. Returns an error if one occurs.
func (c *FakeTrvsSecrets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(trvssecretsResource, c.ns, name), &travisciv1.TrvsSecret{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTrvsSecrets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(trvssecretsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &travisciv1.TrvsSecretList{})
	return err
}

// Patch applies the patch and returns the patched trvsSecret.
func (c *FakeTrvsSecrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *travisciv1.TrvsSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(trvssecretsResource, c.ns, name, pt, data, subresources...), &travisciv1.TrvsSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*travisciv1.TrvsSecret), err
}