package basecontrollers

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/service"
)

func initializeServiceController(ctx controller.Context) {
	go service.NewController(
		ctx.CloudProvider,
		ctx.NamespacePrefix,
		ctx.LatticeID,
		ctx.KubeClientBuilder.ClientOrDie("kubernetes-service-controller"),
		ctx.LatticeClientBuilder.ClientOrDie("kubernetes-service-controller"),
		ctx.LatticeInformerFactory.Lattice().V1().Configs(),
		ctx.LatticeInformerFactory.Lattice().V1().Services(),
		ctx.LatticeInformerFactory.Lattice().V1().NodePools(),
		ctx.KubeInformerFactory.Apps().V1().Deployments(),
		ctx.KubeInformerFactory.Core().V1().Pods(),
		ctx.KubeInformerFactory.Core().V1().Services(),
		ctx.LatticeInformerFactory.Lattice().V1().ServiceAddresses(),
		ctx.LatticeInformerFactory.Lattice().V1().LoadBalancers(),
	).Run(4, ctx.Stop)
}
