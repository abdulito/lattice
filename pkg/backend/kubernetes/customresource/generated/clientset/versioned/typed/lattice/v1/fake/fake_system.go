package fake

import (
	lattice_v1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSystems implements SystemInterface
type FakeSystems struct {
	Fake *FakeLatticeV1
	ns   string
}

var systemsResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "systems"}

var systemsKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "System"}

// Get takes name of the system, and returns the corresponding system object, and an error if there is any.
func (c *FakeSystems) Get(name string, options v1.GetOptions) (result *lattice_v1.System, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(systemsResource, c.ns, name), &lattice_v1.System{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.System), err
}

// List takes label and field selectors, and returns the list of Systems that match those selectors.
func (c *FakeSystems) List(opts v1.ListOptions) (result *lattice_v1.SystemList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(systemsResource, systemsKind, c.ns, opts), &lattice_v1.SystemList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.SystemList{}
	for _, item := range obj.(*lattice_v1.SystemList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested systems.
func (c *FakeSystems) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(systemsResource, c.ns, opts))

}

// Create takes the representation of a system and creates it.  Returns the server's representation of the system, and an error, if there is any.
func (c *FakeSystems) Create(system *lattice_v1.System) (result *lattice_v1.System, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(systemsResource, c.ns, system), &lattice_v1.System{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.System), err
}

// Update takes the representation of a system and updates it. Returns the server's representation of the system, and an error, if there is any.
func (c *FakeSystems) Update(system *lattice_v1.System) (result *lattice_v1.System, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(systemsResource, c.ns, system), &lattice_v1.System{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.System), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSystems) UpdateStatus(system *lattice_v1.System) (*lattice_v1.System, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(systemsResource, "status", c.ns, system), &lattice_v1.System{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.System), err
}

// Delete takes name of the system and deletes it. Returns an error if one occurs.
func (c *FakeSystems) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(systemsResource, c.ns, name), &lattice_v1.System{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSystems) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(systemsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.SystemList{})
	return err
}

// Patch applies the patch and returns the patched system.
func (c *FakeSystems) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.System, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(systemsResource, c.ns, name, data, subresources...), &lattice_v1.System{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.System), err
}
