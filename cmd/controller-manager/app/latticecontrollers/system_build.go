package latticecontrollers

import (
	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
	"github.com/mlab-lattice/kubernetes-integration/pkg/controller/lattice/systembuild"
)

func initializeSystemBuildController(ctx controller.Context) {
	go systembuild.NewSystemBuildController(
		ctx.Provider,
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["system-build"],
		ctx.CRDInformers["service-build"],
	).Run(4, ctx.Stop)
}
