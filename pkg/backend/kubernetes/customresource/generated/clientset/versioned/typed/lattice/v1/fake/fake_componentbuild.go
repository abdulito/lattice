package fake

import (
	lattice_v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeComponentBuilds implements ComponentBuildInterface
type FakeComponentBuilds struct {
	Fake *FakeLatticeV1
	ns   string
}

var componentbuildsResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "componentbuilds"}

var componentbuildsKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "ComponentBuild"}

// Get takes name of the componentBuild, and returns the corresponding componentBuild object, and an error if there is any.
func (c *FakeComponentBuilds) Get(name string, options v1.GetOptions) (result *lattice_v1.ComponentBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(componentbuildsResource, c.ns, name), &lattice_v1.ComponentBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ComponentBuild), err
}

// List takes label and field selectors, and returns the list of ComponentBuilds that match those selectors.
func (c *FakeComponentBuilds) List(opts v1.ListOptions) (result *lattice_v1.ComponentBuildList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(componentbuildsResource, componentbuildsKind, c.ns, opts), &lattice_v1.ComponentBuildList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.ComponentBuildList{}
	for _, item := range obj.(*lattice_v1.ComponentBuildList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested componentBuilds.
func (c *FakeComponentBuilds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(componentbuildsResource, c.ns, opts))

}

// Create takes the representation of a componentBuild and creates it.  Returns the server's representation of the componentBuild, and an error, if there is any.
func (c *FakeComponentBuilds) Create(componentBuild *lattice_v1.ComponentBuild) (result *lattice_v1.ComponentBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(componentbuildsResource, c.ns, componentBuild), &lattice_v1.ComponentBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ComponentBuild), err
}

// Update takes the representation of a componentBuild and updates it. Returns the server's representation of the componentBuild, and an error, if there is any.
func (c *FakeComponentBuilds) Update(componentBuild *lattice_v1.ComponentBuild) (result *lattice_v1.ComponentBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(componentbuildsResource, c.ns, componentBuild), &lattice_v1.ComponentBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ComponentBuild), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeComponentBuilds) UpdateStatus(componentBuild *lattice_v1.ComponentBuild) (*lattice_v1.ComponentBuild, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(componentbuildsResource, "status", c.ns, componentBuild), &lattice_v1.ComponentBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ComponentBuild), err
}

// Delete takes name of the componentBuild and deletes it. Returns an error if one occurs.
func (c *FakeComponentBuilds) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(componentbuildsResource, c.ns, name), &lattice_v1.ComponentBuild{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeComponentBuilds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(componentbuildsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.ComponentBuildList{})
	return err
}

// Patch applies the patch and returns the patched componentBuild.
func (c *FakeComponentBuilds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.ComponentBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(componentbuildsResource, c.ns, name, data, subresources...), &lattice_v1.ComponentBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ComponentBuild), err
}
