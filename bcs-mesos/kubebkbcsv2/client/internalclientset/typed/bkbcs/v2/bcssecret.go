/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Code generated by client-gen. DO NOT EDIT.

package v2

import (
	v2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	scheme "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// BcsSecretsGetter has a method to return a BcsSecretInterface.
// A group's client should implement this interface.
type BcsSecretsGetter interface {
	BcsSecrets(namespace string) BcsSecretInterface
}

// BcsSecretInterface has methods to work with BcsSecret resources.
type BcsSecretInterface interface {
	Create(*v2.BcsSecret) (*v2.BcsSecret, error)
	Update(*v2.BcsSecret) (*v2.BcsSecret, error)
	UpdateStatus(*v2.BcsSecret) (*v2.BcsSecret, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v2.BcsSecret, error)
	List(opts v1.ListOptions) (*v2.BcsSecretList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v2.BcsSecret, err error)
	BcsSecretExpansion
}

// bcsSecrets implements BcsSecretInterface
type bcsSecrets struct {
	client rest.Interface
	ns     string
}

// newBcsSecrets returns a BcsSecrets
func newBcsSecrets(c *BkbcsV2Client, namespace string) *bcsSecrets {
	return &bcsSecrets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the bcsSecret, and returns the corresponding bcsSecret object, and an error if there is any.
func (c *bcsSecrets) Get(name string, options v1.GetOptions) (result *v2.BcsSecret, err error) {
	result = &v2.BcsSecret{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bcssecrets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of BcsSecrets that match those selectors.
func (c *bcsSecrets) List(opts v1.ListOptions) (result *v2.BcsSecretList, err error) {
	result = &v2.BcsSecretList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bcssecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested bcsSecrets.
func (c *bcsSecrets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("bcssecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a bcsSecret and creates it.  Returns the server's representation of the bcsSecret, and an error, if there is any.
func (c *bcsSecrets) Create(bcsSecret *v2.BcsSecret) (result *v2.BcsSecret, err error) {
	result = &v2.BcsSecret{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("bcssecrets").
		Body(bcsSecret).
		Do().
		Into(result)
	return
}

// Update takes the representation of a bcsSecret and updates it. Returns the server's representation of the bcsSecret, and an error, if there is any.
func (c *bcsSecrets) Update(bcsSecret *v2.BcsSecret) (result *v2.BcsSecret, err error) {
	result = &v2.BcsSecret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bcssecrets").
		Name(bcsSecret.Name).
		Body(bcsSecret).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *bcsSecrets) UpdateStatus(bcsSecret *v2.BcsSecret) (result *v2.BcsSecret, err error) {
	result = &v2.BcsSecret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bcssecrets").
		Name(bcsSecret.Name).
		SubResource("status").
		Body(bcsSecret).
		Do().
		Into(result)
	return
}

// Delete takes name of the bcsSecret and deletes it. Returns an error if one occurs.
func (c *bcsSecrets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bcssecrets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *bcsSecrets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bcssecrets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched bcsSecret.
func (c *bcsSecrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v2.BcsSecret, err error) {
	result = &v2.BcsSecret{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("bcssecrets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
