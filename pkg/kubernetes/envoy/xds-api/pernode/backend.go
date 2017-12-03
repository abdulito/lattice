package pernode

import (
	"fmt"
	"time"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/envoy"
	latticeresource "github.com/mlab-lattice/system/pkg/kubernetes/customresource"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"

	corev1informers "k8s.io/client-go/informers/core/v1"
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
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}

	latticeResourceClient, _, err := latticeresource.NewClient(config)
	if err != nil {
		return nil, err
	}

	rest.AddUserAgent(config, "envoy-api-backend")
	kClient := clientset.NewForConfigOrDie(config)

	listerWatcher := cache.NewListWatchFromClient(
		latticeResourceClient,
		crv1.ResourcePluralService,
		string(constants.UserSystemNamespace),
		fields.Everything(),
	)
	lSvcInformer := cache.NewSharedInformer(
		listerWatcher,
		&crv1.Service{},
		time.Duration(12*time.Hour),
	)

	kEndpointInformer := corev1informers.NewEndpointsInformer(
		kClient,
		string(constants.UserSystemNamespace),
		time.Duration(12*time.Hour),
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	// FIXME: should we add a stopCh?
	go lSvcInformer.Run(nil)
	go kEndpointInformer.Run(nil)

	kpnb := &KubernetesPerNodeBackend{
		kEndpointLister:           corelisters.NewEndpointsLister(kEndpointInformer.GetIndexer()),
		kEndpointListerSynced:     kEndpointInformer.HasSynced,
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
func (kpnb *KubernetesPerNodeBackend) Services() (map[tree.NodePath]*envoy.Service, error) {
	result := map[tree.NodePath]*envoy.Service{}

	for _, svcObj := range kpnb.latticeServiceStore.List() {
		svc := svcObj.(*crv1.Service)

		kSvcName := fmt.Sprintf("svc-%v-lattice", svc.Name)
		ep, err := kpnb.kEndpointLister.Endpoints(svc.Namespace).Get(kSvcName)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				continue
			}
			return nil, err
		}

		bsvc := &envoy.Service{
			EgressPort:  svc.Spec.EnvoyEgressPort,
			Components:  map[string]envoy.Component{},
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
			bc := envoy.Component{
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
