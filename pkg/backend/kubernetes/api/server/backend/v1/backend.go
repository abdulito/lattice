package backend

import (
	serverv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/api/server/backend/v1/system"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"

	kubeclientset "k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset_generated/clientset"
)

type Backend struct {
	systems *system.Backend
}

func NewBackend(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
	metricsClient metricsclientset.Interface,
) *Backend {
	return &Backend{
		systems: system.NewBackend(namespacePrefix, kubeClient, latticeClient, metricsClient),
	}
}

func (kb *Backend) Systems() serverv1.SystemBackend {
	return kb.systems
}
