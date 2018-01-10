package provision

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/lifecycle/cluster/provisioner"
)

func getKubernetesProvisioner(providerName string) (provisioner.Interface, error) {
	var options *cloudprovider.ClusterProvisionerOptions
	switch providerName {
	case constants.ProviderLocal:
		options = &cloudprovider.ClusterProvisionerOptions{
			Local: &local.ClusterProvisionerOptions{},
		}

	case constants.ProviderAWS:
		options = &cloudprovider.ClusterProvisionerOptions{
			AWS: backendConfigKubernetes.ProviderConfig.AWS,
		}

	default:
		panic(fmt.Sprintf("unsupported provider: %v", providerName))
	}

	return cloudprovider.NewClusterProvisioner(
		backendConfigKubernetes.LatticeContainerRegistry,
		backendConfigKubernetes.LatticeContainerRepoPrefix,
		workingDir,
		options,
	)
}
