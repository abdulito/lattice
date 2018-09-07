package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncInProgressDeploy(deploy *latticev1.Deploy) error {
	system, err := c.getSystem(deploy.Namespace)
	if err != nil {
		return err
	}

	// Check to see if the system controller has processed updates to its Spec.
	// If it hasn't, the system.Status.State is not up to date. Return no error
	// and wait until the System has been updated to resync.
	if !system.UpdateProcessed() {
		return nil
	}

	var state latticev1.DeployState
	switch system.Status.State {
	case latticev1.SystemStateUpdating, latticev1.SystemStateScaling:
		// Still in progress, nothing more to do
		return nil

	case latticev1.SystemStateStable:
		state = latticev1.DeployStateSucceeded

	case latticev1.SystemStateDegraded:
		state = latticev1.DeployStateFailed

	default:
		return fmt.Errorf("%v in unexpected state %v", system.Description(), system.Status.State)
	}

	deploy, err = c.updateDeployStatus(deploy, state, "", deploy.Status.BuildID)
	if err != nil {
		// FIXME: is it possible that the deploy is locked forever now?
		return err
	}

	if deploy.Status.State == latticev1.DeployStateSucceeded || deploy.Status.State == latticev1.DeployStateFailed {
		return c.relinquishDeployOwningActionClaim(deploy)
	}

	return nil
}
