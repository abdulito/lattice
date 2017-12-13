package system

import (
	"fmt"
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncSystemStatus(system *crv1.System) error {
	hasFailedService := false
	hasUpdatingService := false
	hasScalingService := false

	for path, service := range system.Spec.Services {
		if service.State == nil {
			return fmt.Errorf("System %v's Service %v had no ServiceBuildState in syncSystemStatus", system.Namespace, path)
		}

		// If there's a failed service, no need to look any further, our System has failed.
		if *service.State == crv1.ServiceStateFailed {
			hasFailedService = true
			break
		}

		if *service.State == crv1.ServiceStateUpdating || *service.State == crv1.ServiceStatePending {
			hasUpdatingService = true
			continue
		}

		if *service.State == crv1.ServiceStateScalingDown || *service.State == crv1.ServiceStateScalingUp {
			hasScalingService = true
			continue
		}

		if *service.State != crv1.ServiceStateStable {
			return fmt.Errorf("System %v's Service %v had unexpected state: %v", system.Namespace, path, *service.State)
		}
	}

	state := crv1.SystemStateStable

	// A scaling status takes priority over a stable status
	if hasScalingService {
		state = crv1.SystemStateScaling
	}

	// An updating status takes priority over a scaling status
	if hasUpdatingService {
		state = crv1.SystemStateUpdating
	}

	// A failed status takes priority over an updating status
	if hasFailedService {
		state = crv1.SystemStateFailed
	}

	status := crv1.SystemStatus{
		State: state,
	}

	if reflect.DeepEqual(system.Status, status) {
		return nil
	}

	// Copy the system so the shared cache is not mutated
	system = system.DeepCopy()
	system.Status = status

	_, err := c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
	return err
}
