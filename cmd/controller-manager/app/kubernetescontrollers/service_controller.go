package kubernetescontrollers

import (
	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/kubernetes/service"
)

func initializeServiceController(ctx controller.Context) {
	go service.NewServiceController(
		ctx.ClientBuilder.ClientOrDie("kubernetes-service-controller"),
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["service"],
		ctx.CRDInformers["service-build"],
		ctx.CRDInformers["component-build"],
		ctx.InformerFactory.Extensions().V1beta1().Deployments(),
	).Run(4, ctx.Stop)
}
