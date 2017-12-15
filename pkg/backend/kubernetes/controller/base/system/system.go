package system

import (
	"fmt"
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

func (c *Controller) syncSystemStatus(system *crv1.System, services map[tree.NodePath]string, serviceStatuses map[string]crv1.ServiceStatus, deletedServices []string) error {
	hasFailedService := false
	hasUpdatingService := false
	hasScalingService := false

	for serviceName, status := range serviceStatuses {
		if status.State == crv1.ServiceStateFailed {
			hasFailedService = true
			continue
		}

		if status.State == crv1.ServiceStateUpdating || status.State == crv1.ServiceStatePending {
			hasUpdatingService = true
			continue
		}

		if status.State == crv1.ServiceStateScalingDown || status.State == crv1.ServiceStateScalingUp {
			hasScalingService = true
			continue
		}

		if status.State != crv1.ServiceStateStable {
			return fmt.Errorf("Service %v/%v had unexpected state: %v", system.Namespace, serviceName, status.State)
		}
	}

	state := crv1.SystemStateStable

	// A scaling status takes priority over a stable status
	if hasScalingService || len(deletedServices) != 0 {
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

func (c *Controller) updateSystemStatus(system *crv1.System, state crv1.SystemState, services map[tree.NodePath]string, serviceStatuses map[string]crv1.ServiceStatus) (*crv1.System, error) {
	status := crv1.SystemStatus{
		State:              state,
		ObservedGeneration: system.Generation,
		Services:           services,
		ServiceStatuses:    serviceStatuses,
	}

	if reflect.DeepEqual(system.Status, status) {
		return system, nil
	}

	// Copy so the shared cache isn't mutated
	system = system.DeepCopy()
	system.Status = status

	return c.latticeClient.LatticeV1().Systems(system.Namespace).UpdateStatus(system)
}
