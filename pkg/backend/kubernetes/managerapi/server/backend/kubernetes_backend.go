package backend

import (
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"

	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/types"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesBackend struct {
	clusterID     types.LatticeID
	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	systemBootstrappers []systembootstrapper.Interface
}

func NewKubernetesBackend(
	clusterID types.LatticeID,
	kubeconfig string,
	systemBootstrappers []systembootstrapper.Interface,
) (*KubernetesBackend, error) {
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

	kubeClient, err := kubeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	latticeClient, err := latticeclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kb := &KubernetesBackend{
		clusterID:     clusterID,
		kubeClient:    kubeClient,
		latticeClient: latticeClient,

		systemBootstrappers: systemBootstrappers,
	}
	return kb, nil
}
