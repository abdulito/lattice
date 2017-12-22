// This file was automatically generated by informer-gen

package v1

import (
	lattice_v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	versioned "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	internalinterfaces "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/internalinterfaces"
	v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// ConfigInformer provides access to a shared informer and lister for
// Configs.
type ConfigInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.ConfigLister
}

type configInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewConfigInformer constructs a new informer for Config type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewConfigInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredConfigInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredConfigInformer constructs a new informer for Config type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredConfigInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.LatticeV1().Configs(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.LatticeV1().Configs(namespace).Watch(options)
			},
		},
		&lattice_v1.Config{},
		resyncPeriod,
		indexers,
	)
}

func (f *configInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredConfigInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *configInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&lattice_v1.Config{}, f.defaultInformer)
}

func (f *configInformer) Lister() v1.ConfigLister {
	return v1.NewConfigLister(f.Informer().GetIndexer())
}
