package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncAcceptedDeploy(deploy *latticev1.Deploy) error {
	build, err := c.buildLister.Builds(deploy.Namespace).Get(deploy.Spec.BuildName)
	if err != nil {
		return err
	}

	// Check to see if the system build controller has processed updates to its Spec.
	// If it hasn't, the build.Status.State is not up to date. Return no error
	// and wait until the System has been updated to resync.
	// TODO: don't think we actually need this here
	if !isSystemBuildStatusCurrent(build) {
		return nil
	}

	switch build.Status.State {
	case latticev1.BuildStatePending, latticev1.BuildStateRunning:
		return nil

	case latticev1.BuildStateFailed:
		_, err := c.updateDeployStatus(deploy, latticev1.DeployStateFailed, fmt.Sprintf("SystemBuild %v failed", build.Name))
		if err != nil {
			return err
		}

		return c.relinquishDeployOwningActionClaim(deploy)

	case latticev1.BuildStateSucceeded:
		system, err := c.getSystem(deploy.Namespace)
		if err != nil {
			return err
		}

		services, err := c.systemServices(deploy, build)
		if err != nil {
			return err
		}

		_, err = c.updateSystem(system, services)
		if err != nil {
			return err
		}

		_, err = c.updateDeployStatus(deploy, latticev1.DeployStateInProgress, "")
		return err

	default:
		return fmt.Errorf("SystemBuild %v/%v in unexpected state %v", build.Namespace, build.Name, build.Status.State)
	}
}
