package controllers

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/address"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/build"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/componentbuild"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/nodepool"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/service"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/servicebuild"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/system"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/systemlifecycle"
)

type Initializer func(context Context)

var Initializers = map[string]Initializer{
	AddressController:         initializeAddressController,
	BuildController:           initializeBuildController,
	ComponentBuildController:  initializeComponentBuildController,
	NodePoolController:        initializeNodePoolController,
	ServiceController:         initializeServiceController,
	ServiceBuildController:    initializeServiceBuildController,
	SystemController:          initializeSystemController,
	SystemLifecycleController: initializeSystemLifecycleController,
}

func controllerName(name string) string {
	return fmt.Sprintf("lattice-controller-%v", name)
}

func initializeAddressController(ctx Context) {
	go address.NewController(
		ctx.NamespacePrefix,
		ctx.LatticeID,
		ctx.CloudProviderOptions,
		ctx.ServiceMeshOptions,
		ctx.KubeClientBuilder.ClientOrDie(controllerName(AddressController)),
		ctx.LatticeClientBuilder.ClientOrDie(controllerName(AddressController)),
		ctx.KubeInformerFactory,
		ctx.LatticeInformerFactory,
	).Run(4, ctx.Stop)
}

func initializeBuildController(ctx Context) {
	go build.NewController(
		ctx.NamespacePrefix,
		ctx.LatticeClientBuilder.ClientOrDie(controllerName(BuildController)),
		ctx.LatticeInformerFactory.Lattice().V1().Builds(),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
	).Run(4, ctx.Stop)
}

func initializeComponentBuildController(ctx Context) {
	go componentbuild.NewController(
		ctx.NamespacePrefix,
		ctx.CloudProviderOptions,
		ctx.KubeClientBuilder.ClientOrDie(controllerName(ComponentBuildController)),
		ctx.LatticeClientBuilder.ClientOrDie(controllerName(ComponentBuildController)),
		ctx.KubeInformerFactory,
		ctx.LatticeInformerFactory,
	).Run(4, ctx.Stop)
}

func initializeNodePoolController(ctx Context) {
	go nodepool.NewController(
		ctx.NamespacePrefix,
		ctx.LatticeID,
		ctx.CloudProviderOptions,
		ctx.KubeClientBuilder.ClientOrDie(controllerName(NodePoolController)),
		ctx.LatticeClientBuilder.ClientOrDie(controllerName(NodePoolController)),
		ctx.KubeInformerFactory,
		ctx.LatticeInformerFactory,
	).Run(4, ctx.Stop)
}

func initializeServiceController(ctx Context) {
	go service.NewController(
		ctx.NamespacePrefix,
		ctx.LatticeID,
		ctx.InternalDNSDomain,
		ctx.CloudProviderOptions,
		ctx.ServiceMeshOptions,
		ctx.KubeClientBuilder.ClientOrDie(controllerName(ServiceController)),
		ctx.LatticeClientBuilder.ClientOrDie(controllerName(ServiceController)),
		ctx.KubeInformerFactory,
		ctx.LatticeInformerFactory,
	).Run(4, ctx.Stop)
}

func initializeServiceBuildController(ctx Context) {
	go servicebuild.NewController(
		ctx.NamespacePrefix,
		ctx.LatticeClientBuilder.ClientOrDie(controllerName(ServiceBuildController)),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
		ctx.LatticeInformerFactory.Lattice().V1().ComponentBuilds(),
	).Run(4, ctx.Stop)
}

func initializeSystemController(ctx Context) {
	go system.NewController(
		ctx.NamespacePrefix,
		ctx.LatticeID,
		ctx.CloudProviderOptions,
		ctx.ServiceMeshOptions,
		ctx.KubeClientBuilder.ClientOrDie(controllerName(SystemController)),
		ctx.LatticeClientBuilder.ClientOrDie(controllerName(SystemController)),
		ctx.KubeInformerFactory,
		ctx.LatticeInformerFactory,
	).Run(4, ctx.Stop)
}

func initializeSystemLifecycleController(ctx Context) {
	go systemlifecycle.NewController(
		ctx.NamespacePrefix,
		ctx.KubeClientBuilder.ClientOrDie(controllerName(SystemLifecycleController)),
		ctx.LatticeClientBuilder.ClientOrDie(controllerName(SystemLifecycleController)),
		ctx.LatticeInformerFactory.Lattice().V1().Deploys(),
		ctx.LatticeInformerFactory.Lattice().V1().Teardowns(),
		ctx.LatticeInformerFactory.Lattice().V1().Systems(),
		ctx.LatticeInformerFactory.Lattice().V1().Builds(),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceBuilds(),
		ctx.LatticeInformerFactory.Lattice().V1().ComponentBuilds(),
	).Run(4, ctx.Stop)
}
