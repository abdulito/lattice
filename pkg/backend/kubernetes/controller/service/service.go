package service

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

const (
	finalizerName = "service.lattice.mlab.com"

	reasonTimedOut           = "ProgressDeadlineExceeded"
	reasonLoadBalancerFailed = "LoadBalancerFailed"
)

func (c *Controller) syncServiceStatus(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
	nodePoolReady bool,
	deploymentStatus *deploymentStatus,
	extraNodePoolsExist bool,
	kubeService *corev1.Service,
	serviceAddress *latticev1.ServiceAddress,
	loadBalancer *latticev1.LoadBalancer,
	loadBalancerNeeded bool,
) (*latticev1.Service, error) {
	failed := false
	failureReason := ""
	failureMessage := ""
	var failureTime *metav1.Time

	var state latticev1.ServiceState
	if !deploymentStatus.UpdateProcessed {
		state = latticev1.ServiceStateUpdating
	} else if deploymentStatus.State == deploymentStateFailed {
		failed = true
		failureReason = deploymentStatus.FailureInfo.Reason
		failureMessage = deploymentStatus.FailureInfo.Message
		failureTime = &deploymentStatus.FailureInfo.Time
	} else if deploymentStatus.State == deploymentStateScaling {
		state = latticev1.ServiceStateScaling
	} else {
		state = latticev1.ServiceStateStable
	}

	// The cloud controller is responsible for creating the Kubernetes Service.
	if kubeService == nil {
		state = latticev1.ServiceStateUpdating
	}

	publicPorts := latticev1.ServiceStatusPublicPorts{}
	if loadBalancerNeeded {
		switch loadBalancer.Status.State {
		case latticev1.LoadBalancerStatePending, latticev1.LoadBalancerStateProvisioning:
			state = latticev1.ServiceStateUpdating

		case latticev1.LoadBalancerStateCreated:
			for port, portInfo := range loadBalancer.Status.Ports {
				publicPorts[port] = latticev1.ServiceStatusPublicPort{
					Address: portInfo.Address,
				}
			}

		case latticev1.LoadBalancerStateFailed:
			// Only create new failure info if we didn't fail above
			if !failed {
				now := metav1.Now()
				failed = true
				failureReason = reasonLoadBalancerFailed
				failureMessage = ""
				failureTime = &now
			}

		default:
			err := fmt.Errorf(
				"LoadBalancer %v/%v has unexpected state %v",
				loadBalancer.Namespace,
				loadBalancer.Name,
				loadBalancer.Status.State,
			)
			return nil, err
		}
	}

	// But if we have a failure, our updating or scaling has failed
	// A failed status takes priority over an updating status
	var failureInfo *latticev1.ServiceFailureInfo
	if failed {
		state = latticev1.ServiceStateFailed
		switch failureReason {
		case reasonTimedOut:
			failureInfo = &latticev1.ServiceFailureInfo{
				Internal: false,
				Message:  "timed out",
				Time:     *failureTime,
			}

		case reasonLoadBalancerFailed:
			failureInfo = &latticev1.ServiceFailureInfo{
				Internal: false,
				Message:  "load balancer failed",
				Time:     *failureTime,
			}

		default:
			failureInfo = &latticev1.ServiceFailureInfo{
				Internal: true,
				Message:  fmt.Sprintf("%v: %v", failureReason, failureMessage),
				Time:     *failureTime,
			}
		}
	}

	return c.updateServiceStatus(service, state, deploymentStatus.UpdatedInstances, deploymentStatus.StaleInstances, publicPorts, failureInfo)
}

func (c *Controller) updateServiceStatus(
	service *latticev1.Service,
	state latticev1.ServiceState,
	updatedInstances, staleInstances int32,
	publicPorts latticev1.ServiceStatusPublicPorts,
	failureInfo *latticev1.ServiceFailureInfo,
) (*latticev1.Service, error) {
	status := latticev1.ServiceStatus{
		State:              state,
		ObservedGeneration: service.Generation,
		UpdateProcessed:    true,
		UpdatedInstances:   updatedInstances,
		StaleInstances:     staleInstances,
		PublicPorts:        publicPorts,
		FailureInfo:        failureInfo,
	}

	if reflect.DeepEqual(service.Status, status) {
		return service, nil
	}

	// Copy the service so the shared cache isn't mutated
	service = service.DeepCopy()
	service.Status = status

	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().Services(service.Namespace).UpdateStatus(service)
}

type lookupDelete struct {
	lookup func() (interface{}, error)
	delete func() error
}

// FIXME: remove this, was a remnant of cascading garbage collection not working for CRDs, add owner reference instead
func (c *Controller) syncDeletedService(service *latticev1.Service) error {
	lookupDeletes := []lookupDelete{
		// node pool
		// FIXME: need to change this to support system etc level node pools
		// FIXME: should potentially wait until deployment is cleaned up before deleting node pool
		//        to allow for graceful termination
		{
			lookup: func() (interface{}, error) {
				return c.nodePoolLister.NodePools(service.Namespace).Get(service.Name)
			},
			delete: func() error {
				return c.latticeClient.LatticeV1().NodePools(service.Namespace).Delete(service.Name, nil)
			},
		},
		// deployment
		// FIXME: is any of this even working? we don't name the deployment this
		{
			lookup: func() (interface{}, error) {
				return c.deploymentLister.Deployments(service.Namespace).Get(service.Name)
			},
			delete: func() error {
				return c.kubeClient.AppsV1().Deployments(service.Namespace).Delete(service.Name, nil)
			},
		},
		// kube service
		{
			lookup: func() (interface{}, error) {
				name := kubeutil.GetKubeServiceNameForService(service.Name)
				return c.kubeServiceLister.Services(service.Namespace).Get(name)
			},
			delete: func() error {
				name := kubeutil.GetKubeServiceNameForService(service.Name)
				return c.kubeClient.CoreV1().Services(service.Namespace).Delete(name, nil)
			},
		},
		// service address
		{
			lookup: func() (interface{}, error) {
				return c.serviceAddressLister.ServiceAddresses(service.Namespace).Get(service.Name)
			},
			delete: func() error {
				return c.latticeClient.LatticeV1().ServiceAddresses(service.Namespace).Delete(service.Name, nil)
			},
		},
		// load balancer
		{
			lookup: func() (interface{}, error) {
				return c.loadBalancerLister.LoadBalancers(service.Namespace).Get(service.Name)
			},
			delete: func() error {
				return c.latticeClient.LatticeV1().LoadBalancers(service.Namespace).Delete(service.Name, nil)
			},
		},
	}

	existingResource := false
	for _, lookupDelete := range lookupDeletes {
		exists, err := resourceExists(lookupDelete.lookup)
		if err != nil {
			return err
		}

		if exists {
			existingResource = true
			if err := lookupDelete.delete(); err != nil {
				return err
			}

			continue
		}
	}

	if existingResource {
		return nil
	}

	// All of the children resources have been cleaned up
	_, err := c.removeFinalizer(service)
	return err
}

func resourceExists(lookupFunc func() (interface{}, error)) (bool, error) {
	_, err := lookupFunc()
	if err == nil {
		// resource still exists, wait until it is deleted
		return true, nil
	}

	if !errors.IsNotFound(err) {
		return false, err
	}

	return false, nil
}

func controllerRef(service *latticev1.Service) *metav1.OwnerReference {
	return metav1.NewControllerRef(service, controllerKind)
}

// FIXME: don't think we need a finalizer anymore if cascading garbage collection works
func (c *Controller) addFinalizer(service *latticev1.Service) (*latticev1.Service, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range service.Finalizers {
		if finalizer == finalizerName {
			glog.V(5).Infof("service %v has %v finalizer", service.Name, finalizerName)
			return service, nil
		}
	}

	// Add the finalizer to the list and update.
	// If this fails due to a race the Endpoint should get requeued by the controller, so
	// not a big deal.
	service.Finalizers = append(service.Finalizers, finalizerName)
	glog.V(5).Infof("service %v missing %v finalizer, adding it", service.Name, finalizerName)

	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}

func (c *Controller) removeFinalizer(service *latticev1.Service) (*latticev1.Service, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	found := false
	var finalizers []string
	for _, finalizer := range service.Finalizers {
		if finalizer == finalizerName {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return service, nil
	}

	// The finalizer was in the list, so we should remove it.
	service.Finalizers = finalizers
	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}
