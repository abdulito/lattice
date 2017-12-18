package base

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/types"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Options struct {
	DryRun           bool
	Config           crv1.ConfigSpec
	MasterComponents MasterComponentOptions
}

type MasterComponentOptions struct {
	LatticeControllerManager LatticeControllerManagerOptions
	ManagerAPI               ManagerAPIOptions
}

type LatticeControllerManagerOptions struct {
	Image string
	Args  []string
}

type ManagerAPIOptions struct {
	Image       string
	Port        int32
	HostNetwork bool
	Args        []string
}

func NewBootstrapper(
	clusterID types.ClusterID,
	options *Options,
	kubeConfig *rest.Config,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) (*DefaultBootstrapper, error) {
	if options == nil {
		return nil, fmt.Errorf("options required")
	}

	provider, err := crv1.GetProviderFromConfigSpec(&options.Config)
	if err != nil {
		return nil, err
	}

	b := &DefaultBootstrapper{
		Options:       options,
		ClusterID:     clusterID,
		KubeConfig:    kubeConfig,
		KubeClient:    kubeClient,
		Provider:      provider,
		LatticeClient: latticeClient,
	}
	return b, nil
}

type DefaultBootstrapper struct {
	Options    *Options
	ClusterID  types.ClusterID
	KubeConfig *rest.Config
	KubeClient kubeclientset.Interface

	Provider string

	LatticeClient latticeclientset.Interface
}

func (b *DefaultBootstrapper) BaseBootstrap() ([]interface{}, error) {
	bootstrapFuncs := []func() ([]interface{}, error){
		b.seedNamespaces,
		b.seedCRD,
		b.seedRBAC,
		b.seedConfig,
		b.seedMasterComponents,
	}

	objects := []interface{}{}
	for _, bootstrapFunc := range bootstrapFuncs {
		additionalObjects, err := bootstrapFunc()
		if err != nil {
			return nil, err
		}
		objects = append(objects, additionalObjects...)
	}
	return objects, nil
}
