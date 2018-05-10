package cloudprovider

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/local"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/dnsprovider"
	systembootstrapper "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	kubeclientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type Interface interface {
	systembootstrapper.Interface
	dnsprovider.Interface
	AddressLoadBalancer
	NodePool

	// TransformComponentBuildJobSpec takes in the JobSpec generated for a ComponentBuild, and applies any cloud provider
	// related transforms necessary to a copy of the JobSpec, and returns it.
	TransformComponentBuildJobSpec(*batchv1.JobSpec) *batchv1.JobSpec

	ComponentBuildWorkDirectoryVolumeSource(jobName string) corev1.VolumeSource

	// TransformServiceDeploymentSpec takes in the DeploymentSpec generated for a Service, and applies any cloud provider
	// related transforms necessary to a copy of the DeploymentSpec, and returns it.
	TransformServiceDeploymentSpec(*latticev1.Service, *appsv1.DeploymentSpec) *appsv1.DeploymentSpec

	// IsDeploymentSpecCurrent checks to see if any part of the current DeploymentSpec that the service mesh is responsible
	// for is out of date compared to the desired deployment spec. If the current DeploymentSpec is current, it also returns
	// a copy of the desired DeploymentSpec with the negation of TransformServiceDeploymentSpec applied.
	// That is, if the aspects of the DeploymentSpec that were transformed by TransformServiceDeploymentSpec are all still
	// current, this method should return true, along with a copy of the DeploymentSpec that should be identical to the
	// DeploymentSpec that was passed in to TransformServiceDeploymentSpec.
	IsDeploymentSpecUpdated(service *latticev1.Service, current, desired, untransformed *appsv1.DeploymentSpec) (bool, string, *appsv1.DeploymentSpec)
}

type AddressLoadBalancer interface {
	ServiceAddressLoadBalancerNeedsUpdate(
		latticeID v1.LatticeID,
		address *latticev1.Address,
		service *latticev1.Service,
		serviceMeshPorts map[int32]int32,
	) (bool, error)
	EnsureServiceAddressLoadBalancer(
		latticeID v1.LatticeID,
		address *latticev1.Address,
		service *latticev1.Service,
		serviceMeshPorts map[int32]int32,
	) error
	DestroyServiceAddressLoadBalancer(v1.LatticeID, *latticev1.Address) error
	ServiceAddressLoadBalancerAddAnnotations(
		latticeID v1.LatticeID,
		address *latticev1.Address,
		service *latticev1.Service,
		serviceMeshPorts map[int32]int32,
		annotations map[string]string,
	) error
	ServiceAddressLoadBalancerPorts(v1.LatticeID, *latticev1.Address, *latticev1.Service) (map[int32]string, error)
}

type NodePool interface {
	NodePoolNeedsNewEpoch(*latticev1.NodePool) (bool, error)
	EnsureNodePoolEpoch(v1.LatticeID, *latticev1.NodePool, latticev1.NodePoolEpoch) error
	DestroyNodePoolEpoch(v1.LatticeID, *latticev1.NodePool, latticev1.NodePoolEpoch) error
	NodePoolAddAnnotations(v1.LatticeID, *latticev1.NodePool, map[string]string, latticev1.NodePoolEpoch) error
}

type Options struct {
	AWS   *aws.Options
	Local *local.Options
}

func NewCloudProvider(
	namespacePrefix string,
	kubeClient kubeclientset.Interface,
	kubeServiceLister corelisters.ServiceLister,
	nodePoolLister latticelisters.NodePoolLister,
	options *Options,
) (Interface, error) {
	if options.AWS != nil {
		return aws.NewCloudProvider(namespacePrefix, kubeClient, kubeServiceLister, nodePoolLister, options.AWS), nil
	}

	if options.Local != nil {
		return local.NewCloudProvider(namespacePrefix, kubeClient, kubeServiceLister, options.Local), nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

func OverlayConfigOptions(staticOptions *Options, dynamicConfig *latticev1.ConfigCloudProvider) (*Options, error) {
	if staticOptions.AWS != nil {
		if dynamicConfig.AWS == nil {
			return nil, fmt.Errorf("static options were for AWS but dynamic config did not have AWS options set")
		}

		awsOptions, err := aws.NewOptions(staticOptions.AWS, dynamicConfig.AWS)
		if err != nil {
			return nil, err
		}

		options := &Options{
			AWS: awsOptions,
		}
		return options, nil
	}

	if staticOptions.Local != nil {
		if dynamicConfig.Local == nil {
			return nil, fmt.Errorf("static options were for local but dynamic config did not have local options set")
		}

		localOptions, err := local.NewOptions(staticOptions.Local, dynamicConfig.Local)
		if err != nil {
			return nil, err
		}

		options := &Options{
			Local: localOptions,
		}
		return options, nil
	}

	return nil, fmt.Errorf("must provide cloud provider options")
}

func Flag(cloudProvider *string) (cli.Flag, *Options) {
	awsFlags, awsOptions := aws.Flags()
	localFlags, localOptions := local.Flags()
	options := &Options{}

	flag := &cli.DelayedEmbeddedFlag{
		Name:     "cloud-provider-var",
		Required: true,
		Usage:    "configuration for the cloud provider",
		Flags: map[string]cli.Flags{
			AWS:   awsFlags,
			Local: localFlags,
		},
		FlagChooser: func() (*string, error) {
			if cloudProvider == nil {
				return nil, fmt.Errorf("cloud provider cannot be nil")
			}

			switch *cloudProvider {
			case Local:
				options.Local = localOptions
			case AWS:
				options.AWS = awsOptions
			default:
				return nil, fmt.Errorf("unsupported cloud provider %v", *cloudProvider)
			}

			return cloudProvider, nil
		},
	}

	return flag, options
}
