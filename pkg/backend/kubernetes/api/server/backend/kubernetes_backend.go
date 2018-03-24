package backend

import (
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"

	"github.com/mlab-lattice/system/pkg/api/v1"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesBackend struct {
	latticeID     v1.LatticeID
	kubeClient    kubeclientset.Interface
	latticeClient latticeclientset.Interface

	systemBootstrappers []systembootstrapper.Interface
}

func NewKubernetesBackend(
	latticeID v1.LatticeID,
	kubeconfig string,
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
		latticeID:     latticeID,
		kubeClient:    kubeClient,
		latticeClient: latticeClient,
	}
	return kb, nil
}
