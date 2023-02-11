/*
Copyright DNITSCH WTFPL
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	scheme "github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned/scheme"
	v1alpha1 "github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// RestStrategiesGetter has a method to return a RestStrategyInterface.
// A group's client should implement this interface.
type RestStrategiesGetter interface {
	RestStrategies(namespace string) RestStrategyInterface
}

// RestStrategyInterface has methods to work with RestStrategy resources.
type RestStrategyInterface interface {
	Create(ctx context.Context, restStrategy *v1alpha1.RestStrategy, opts v1.CreateOptions) (*v1alpha1.RestStrategy, error)
	Update(ctx context.Context, restStrategy *v1alpha1.RestStrategy, opts v1.UpdateOptions) (*v1alpha1.RestStrategy, error)
	UpdateStatus(ctx context.Context, restStrategy *v1alpha1.RestStrategy, opts v1.UpdateOptions) (*v1alpha1.RestStrategy, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.RestStrategy, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.RestStrategyList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.RestStrategy, err error)
	RestStrategyExpansion
}

// restStrategies implements RestStrategyInterface
type restStrategies struct {
	client rest.Interface
	ns     string
}

// newRestStrategies returns a RestStrategies
func newRestStrategies(c *ReststrategyV1alpha1Client, namespace string) *restStrategies {
	return &restStrategies{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the restStrategy, and returns the corresponding restStrategy object, and an error if there is any.
func (c *restStrategies) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.RestStrategy, err error) {
	result = &v1alpha1.RestStrategy{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("reststrategies").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of RestStrategies that match those selectors.
func (c *restStrategies) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.RestStrategyList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.RestStrategyList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("reststrategies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested restStrategies.
func (c *restStrategies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("reststrategies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a restStrategy and creates it.  Returns the server's representation of the restStrategy, and an error, if there is any.
func (c *restStrategies) Create(ctx context.Context, restStrategy *v1alpha1.RestStrategy, opts v1.CreateOptions) (result *v1alpha1.RestStrategy, err error) {
	result = &v1alpha1.RestStrategy{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("reststrategies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(restStrategy).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a restStrategy and updates it. Returns the server's representation of the restStrategy, and an error, if there is any.
func (c *restStrategies) Update(ctx context.Context, restStrategy *v1alpha1.RestStrategy, opts v1.UpdateOptions) (result *v1alpha1.RestStrategy, err error) {
	result = &v1alpha1.RestStrategy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("reststrategies").
		Name(restStrategy.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(restStrategy).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *restStrategies) UpdateStatus(ctx context.Context, restStrategy *v1alpha1.RestStrategy, opts v1.UpdateOptions) (result *v1alpha1.RestStrategy, err error) {
	result = &v1alpha1.RestStrategy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("reststrategies").
		Name(restStrategy.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(restStrategy).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the restStrategy and deletes it. Returns an error if one occurs.
func (c *restStrategies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("reststrategies").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *restStrategies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("reststrategies").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched restStrategy.
func (c *restStrategies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.RestStrategy, err error) {
	result = &v1alpha1.RestStrategy{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("reststrategies").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
