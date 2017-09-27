package backend

import (
	"time"

	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	latticeresource "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	"github.com/mlab-lattice/envoy-xds-api/pkg/backend"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"

	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesPerNodeBackend struct {
	kEndpointLister       corelisters.EndpointsLister
	kEndpointListerSynced cache.InformerSynced

	latticeServiceStore       cache.Store
	latticeServiceStoreSynced cache.InformerSynced
}

func NewKubernetesPerNodeBackend(kubeconfig string) (*KubernetesPerNodeBackend, error) {
	// TODO: create in-cluster config if in cluster
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	latticeResourceClient, _, err := latticeresource.NewClient(config)
	if err != nil {
		return nil, err
	}

	rest.AddUserAgent(config, "envoy-api-backend")
	kClient := clientset.NewForConfigOrDie(config)
	kInformers := informers.NewSharedInformerFactory(kClient, time.Duration(12*time.Hour))

	listerWatcher := cache.NewListWatchFromClient(
		latticeResourceClient,
		crv1.ServiceResourcePlural,
		apiv1.NamespaceAll,
		fields.Everything(),
	)
	lSvcInformer := cache.NewSharedInformer(
		listerWatcher,
		&crv1.Service{},
		time.Duration(12*time.Hour),
	)

	// FIXME: should we add a stopCh?
	lSvcInformer.Run(nil)
	kInformers.Start(nil)

	kEndpointInformer := kInformers.Core().V1().Endpoints()
	kpnb := &KubernetesPerNodeBackend{
		kEndpointLister:           kEndpointInformer.Lister(),
		kEndpointListerSynced:     kEndpointInformer.Informer().HasSynced,
		latticeServiceStore:       lSvcInformer.GetStore(),
		latticeServiceStoreSynced: lSvcInformer.HasSynced,
	}
	return kpnb, nil
}

func (kpnb *KubernetesPerNodeBackend) Ready() bool {
	return cache.WaitForCacheSync(nil, kpnb.kEndpointListerSynced, kpnb.latticeServiceStoreSynced)
}

// TODO: probably want to have Services return a cached snapshot of the service state so we don't have to recompute this every time
// 	     For example, could add hooks to the informers which creates a new Services map and changes the pointer to point to the new one
//       so future Services() calls will return the new map.
// 		 Could also have the backend have a channel passed into it and it could notify the API when an update has occurred.
//       This could be useful for the GRPC streaming version of the API.
func (kpnb *KubernetesPerNodeBackend) Services() (map[systemtree.NodePath]*backend.Service, error) {
	result := map[systemtree.NodePath]*backend.Service{}

	for _, svcObj := range kpnb.latticeServiceStore.List() {
		svc := svcObj.(*crv1.Service)

		ep, err := kpnb.kEndpointLister.Endpoints(svc.Namespace).Get(svc.Name)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				continue
			}
			return nil, err
		}

		bsvc := &backend.Service{
			Components:  map[string]backend.Component{},
			IPAddresses: []string{},
		}

		addressSet := map[string]bool{}
		for _, subset := range ep.Subsets {
			for _, address := range subset.Addresses {
				// FIXME: check if this is necessary (i.e. does Endpoint ever repeat IPAddresses)
				if _, ok := addressSet[address.IP]; !ok {
					addressSet[address.IP] = true
					bsvc.IPAddresses = append(bsvc.IPAddresses, address.IP)
				}
			}
		}

		for component, ports := range svc.Spec.Ports {
			bc := backend.Component{
				Ports: map[int32]int32{},
			}

			for _, port := range ports {
				bc.Ports[port.Port] = port.EnvoyPort
			}

			bsvc.Components[component] = bc
		}

		result[svc.Spec.Path] = bsvc
	}

	return result, nil
}
