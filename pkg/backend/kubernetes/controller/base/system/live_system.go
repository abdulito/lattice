package system

import (
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncLiveSystem(system *latticev1.System) error {
	services, serviceStatuses, deletedServices, err := c.syncSystemServices(system)
	if err != nil {
		return err
	}

	return c.syncSystemStatus(system, services, serviceStatuses, deletedServices)
}
