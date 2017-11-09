package latticecontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	"github.com/mlab-lattice/system/pkg/kubernetes/controller/lattice/systemrollout"
)

func initializeSystemRolloutController(ctx controller.Context) {
	go systemrollout.NewSystemRolloutController(
		ctx.LatticeResourceRestClient,
		ctx.CRDInformers["system-rollout"],
		ctx.CRDInformers["system"],
		ctx.CRDInformers["system-build"],
		ctx.CRDInformers["service-build"],
		ctx.CRDInformers["component-build"],
	).Run(4, ctx.Stop)
}
