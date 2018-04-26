package componentbuild

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	batchv1 "k8s.io/api/batch/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) handleConfigAdd(obj interface{}) {
	config := obj.(*latticev1.Config)
	glog.V(4).Infof("Adding Config %s", config.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = config.DeepCopy().Spec

	err := c.newCloudProvider()
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}

	if !c.configSet {
		c.configSet = true
		close(c.configSetChan)
	}
}

func (c *Controller) handleConfigUpdate(old, cur interface{}) {
	oldConfig := old.(*latticev1.Config)
	curConfig := cur.(*latticev1.Config)
	glog.V(4).Infof("Updating Config %s", oldConfig.Name)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = curConfig.DeepCopy().Spec

	err := c.newCloudProvider()
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return
	}
}

func (c *Controller) newCloudProvider() error {
	options, err := cloudprovider.OverlayConfigOptions(c.staticCloudProviderOptions, &c.config.CloudProvider)
	if err != nil {
		return err
	}

	cloudProvider, err := cloudprovider.NewCloudProvider(c.namespacePrefix, nil, nil, options)
	if err != nil {
		return err
	}

	c.cloudProvider = cloudProvider
	return nil
}

func (c *Controller) handleComponentBuildAdd(obj interface{}) {
	build := obj.(*latticev1.ComponentBuild)

	if build.DeletionTimestamp != nil {
		// nothing to be done for deleted component builds
		return
	}

	c.handleComponentBuildEvent(build, "added")
}

func (c *Controller) handleComponentBuildUpdate(old, cur interface{}) {
	build := cur.(*latticev1.ComponentBuild)
	c.handleComponentBuildEvent(build, "updated")
}

func (c *Controller) handleComponentBuildEvent(build *latticev1.ComponentBuild, verb string) {
	glog.V(4).Infof("%v %v", build.Description(c.namespacePrefix), verb)
	c.enqueue(build)
}

func (c *Controller) handleJobAdd(obj interface{}) {
	job := obj.(*batchv1.Job)

	if job.DeletionTimestamp != nil {
		// jobs we care about should only be deleted if the component build is already
		// deleted
		return
	}

	c.handleJobEvent(job, "added")
}

func (c *Controller) handleJobUpdate(old, cur interface{}) {
	job := cur.(*batchv1.Job)
	c.handleJobEvent(job, "updated")
}

func (c *Controller) handleJobEvent(job *batchv1.Job, verb string) {
	glog.V(4).Infof("job %v/%v %v", job.Namespace, job.Name, verb)

	// see if the deployment has a service as a controller owning reference
	if controllerRef := metav1.GetControllerOf(job); controllerRef != nil {
		componentBuild := c.resolveControllerRef(job.Namespace, controllerRef)

		// not a component build job
		if componentBuild == nil {
			return
		}

		c.enqueue(componentBuild)
		return
	}
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *latticev1.ComponentBuild {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != latticev1.ComponentBuildKind.Kind {
		return nil
	}

	build, err := c.componentBuildLister.ComponentBuilds(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	if build.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return build
}
