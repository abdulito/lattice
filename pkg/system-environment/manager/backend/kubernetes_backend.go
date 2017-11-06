package backend

import (
	latticeresource "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesBackend struct {
	LatticeResourceClient rest.Interface
	KubeClientset         clientset.Interface
}

func NewKubernetesBackend(kubeconfig string) (*KubernetesBackend, error) {
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

	kubeClientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	latticeResourceClient, _, err := latticeresource.NewClient(config)
	if err != nil {
		return nil, err
	}

	kb := &KubernetesBackend{
		LatticeResourceClient: latticeResourceClient,
		KubeClientset:         kubeClientset,
	}
	return kb, nil
}