package systemlifecycle

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncAcceptedDeploy(deploy *latticev1.Deploy) error {
	build, err := c.buildLister.Builds(deploy.Namespace).Get(deploy.Spec.BuildName)
	if err != nil {
		return err
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

		version := v1.SystemVersion("unknown")
		if label, ok := deploy.DefinitionVersionLabel(); ok {
			version = label
		}

		buildID := v1.BuildID("unknown")
		if label, ok := deploy.BuildIDLabel(); ok {
			buildID = label
		}

		deployID := v1.DeployID(deploy.Name)

		system, err = c.updateSystemLabels(system, &version, &deployID, &buildID)
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
