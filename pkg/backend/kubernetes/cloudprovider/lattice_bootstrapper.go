package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/util/cli"
)

type ClusterBootstrapperOptions struct {
	AWS   *aws.LatticeBootstrapperOptions
	Local *local.LatticeBootstrapperOptions
}

func NewLatticeBootstrapper(latticeID v1.LatticeID, options *ClusterBootstrapperOptions) (clusterbootstrapper.Interface, error) {
	if options.AWS != nil {
		return aws.NewLatticeBootstrapper(options.AWS), nil
	}

	if options.Local != nil {
		return local.NewLatticeBootstrapper(latticeID, options.Local), nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

func LatticeBoostrapperFlag(cloudProvider *string) (cli.Flag, *ClusterBootstrapperOptions) {
	awsFlags, awsOptions := aws.LatticeBootstrapperFlags()
	localFlags, localOptions := local.LatticeBootstrapperFlags()
	options := &ClusterBootstrapperOptions{}

	flag := &cli.DelayedEmbeddedFlag{
		Name:     "cloud-provider-var",
		Required: true,
		Usage:    "configuration for the cloud provider lattice bootstrapper",
		Flags: map[string]cli.Flags{
			AWS:   awsFlags,
			Local: localFlags,
		},
		FlagChooser: func() (string, error) {
			if cloudProvider == nil {
				return "", fmt.Errorf("cloud provider cannot be nil")
			}

			switch *cloudProvider {
			case Local:
				options.Local = localOptions
			case AWS:
				options.AWS = awsOptions
			default:
				return "", fmt.Errorf("unsupported cloud provider %v", *cloudProvider)
			}

			return *cloudProvider, nil
		},
	}

	return flag, options
}
