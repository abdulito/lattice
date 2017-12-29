// This file was automatically generated by lister-gen

package v1

import (
	v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SystemTeardownLister helps list SystemTeardowns.
type SystemTeardownLister interface {
	// List lists all SystemTeardowns in the indexer.
	List(selector labels.Selector) (ret []*v1.SystemTeardown, err error)
	// SystemTeardowns returns an object that can list and get SystemTeardowns.
	SystemTeardowns(namespace string) SystemTeardownNamespaceLister
	SystemTeardownListerExpansion
}

// systemTeardownLister implements the SystemTeardownLister interface.
type systemTeardownLister struct {
	indexer cache.Indexer
}

// NewSystemTeardownLister returns a new SystemTeardownLister.
func NewSystemTeardownLister(indexer cache.Indexer) SystemTeardownLister {
	return &systemTeardownLister{indexer: indexer}
}

// List lists all SystemTeardowns in the indexer.
func (s *systemTeardownLister) List(selector labels.Selector) (ret []*v1.SystemTeardown, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.SystemTeardown))
	})
	return ret, err
}

// SystemTeardowns returns an object that can list and get SystemTeardowns.
func (s *systemTeardownLister) SystemTeardowns(namespace string) SystemTeardownNamespaceLister {
	return systemTeardownNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SystemTeardownNamespaceLister helps list and get SystemTeardowns.
type SystemTeardownNamespaceLister interface {
	// List lists all SystemTeardowns in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.SystemTeardown, err error)
	// Get retrieves the SystemTeardown from the indexer for a given namespace and name.
	Get(name string) (*v1.SystemTeardown, error)
	SystemTeardownNamespaceListerExpansion
}

// systemTeardownNamespaceLister implements the SystemTeardownNamespaceLister
// interface.
type systemTeardownNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SystemTeardowns in the indexer for a given namespace.
func (s systemTeardownNamespaceLister) List(selector labels.Selector) (ret []*v1.SystemTeardown, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.SystemTeardown))
	})
	return ret, err
}

// Get retrieves the SystemTeardown from the indexer for a given namespace and name.
func (s systemTeardownNamespaceLister) Get(name string) (*v1.SystemTeardown, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("systemteardown"), name)
	}
	return obj.(*v1.SystemTeardown), nil
}
