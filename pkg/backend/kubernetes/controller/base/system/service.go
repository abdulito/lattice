package system

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/system/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"

	"github.com/golang/glog"
	"github.com/satori/go.uuid"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) syncSystemServices(
	system *latticev1.System,
) (map[tree.NodePath]string, map[string]latticev1.ServiceStatus, []string, error) {
	// Maps Service path to Service.Name of the Service
	services := map[tree.NodePath]string{}

	// Maps Service.Name to Service.Status
	serviceStatuses := map[string]latticev1.ServiceStatus{}

	systemNamespace := kubeutil.SystemNamespace(c.latticeID, v1.SystemID(system.Name))

	// Loop through the Services defined in the System's Spec, and create/update any that need it
	for path, serviceInfo := range system.Spec.Services {
		var service *latticev1.Service

		serviceName, ok := system.Status.Services[path]
		if !ok {
			pathDomain := path.ToDomain(true)
			// We don't have the name of the Service in our Status, but it may still have been created already.
			// First, look in the cache for a Service with the proper label.
			selector := kubelabels.NewSelector()
			requirement, err := kubelabels.NewRequirement(kubeconstants.LabelKeyServicePathDomain, selection.Equals, []string{pathDomain})
			if err != nil {
				return nil, nil, nil, err
			}

			selector = selector.Add(*requirement)
			services, err := c.serviceLister.Services(systemNamespace).List(selector)
			if err != nil {
				return nil, nil, nil, err
			}

			if len(services) > 1 {
				err := fmt.Errorf(
					"multiple Services in the %v namespace are labeled with %v = %v",
					systemNamespace,
					kubeconstants.LabelKeyServicePathDomain,
					pathDomain,
				)
				return nil, nil, nil, err
			}

			if len(services) == 1 {
				service = services[0]
			}

			if len(services) == 0 {
				// The cache did not have a Service matching the label.
				// However, it would be a constraint violation to have multiple Services for the same path,
				// so we'll have to do a quorum read from the API to make sure that the Service does not exist.
				services, err := c.latticeClient.LatticeV1().Services(systemNamespace).List(metav1.ListOptions{LabelSelector: selector.String()})
				if err != nil {
					return nil, nil, nil, err
				}

				if len(services.Items) > 1 {
					err := fmt.Errorf(
						"multiple Services in the %v namespace are labeled with %v = %v",
						systemNamespace,
						kubeconstants.LabelKeyServicePathDomain,
						pathDomain,
					)
					return nil, nil, nil, err
				}

				if len(services.Items) == 1 {
					service = &services.Items[0]
				}

				if len(services.Items) == 0 {
					// We are now sure that the Service does not exist, so now we can create it.
					service, err = c.createNewService(system, &serviceInfo, path)
				}
			}
		}

		if service == nil {
			var err error
			service, err = c.serviceLister.Services(systemNamespace).Get(serviceName)
			if err != nil {
				if !errors.IsNotFound(err) {
					return nil, nil, nil, err
				}

				// the Service wasn't in our cache, so check with the API
				service, err = c.latticeClient.LatticeV1().Services(systemNamespace).Get(serviceName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						// FIXME: send warn event
						// TODO: should we just create a new Service here?
						return nil, nil, nil, fmt.Errorf(
							"Service %v in namespace %v has Name %v but Service does not exist",
							path,
							systemNamespace,
							serviceName,
						)
					}

					return nil, nil, nil, err
				}
			}
		}

		// Otherwise, get a new spec and update the service
		spec, err := c.serviceSpec(system, &serviceInfo, path)
		if err != nil {
			return nil, nil, nil, err
		}

		service, err = c.updateService(service, spec)
		if err != nil {
			return nil, nil, nil, err
		}

		services[path] = service.Name
		serviceStatuses[service.Name] = service.Status
	}

	// Loop through all of the Services that exist in the System's namespace, and delete any
	// that are no longer a part of the System's Spec
	// TODO(kevinrosendahl): should we wait until all other services are successfully rolled out before deleting these?
	// need to figure out what the rollout/automatic roll-back strategy is
	allServices, err := c.serviceLister.Services(systemNamespace).List(kubelabels.Everything())
	if err != nil {
		return nil, nil, nil, err
	}

	var deletedServices []string
	for _, service := range allServices {
		if _, ok := serviceStatuses[service.Name]; !ok {
			glog.V(4).Infof(
				"Found Service %q in Namespace %q that is no longer in the System Spec",
				service.Name,
				service.Namespace,
			)
			deletedServices = append(deletedServices, service.Name)

			if service.DeletionTimestamp == nil {
				err := c.latticeClient.LatticeV1().Services(service.Namespace).Delete(service.Name, &metav1.DeleteOptions{})
				if err != nil {
					return nil, nil, nil, err
				}
			}
		}
	}

	return services, serviceStatuses, deletedServices, nil
}

func (c *Controller) createNewService(
	system *latticev1.System,
	serviceInfo *latticev1.SystemSpecServiceInfo,
	path tree.NodePath,
) (*latticev1.Service, error) {
	service, err := c.newService(system, serviceInfo, path)
	if err != nil {
		return nil, err
	}

	return c.latticeClient.LatticeV1().Services(service.Namespace).Create(service)
}

func (c *Controller) newService(
	system *latticev1.System,
	serviceInfo *latticev1.SystemSpecServiceInfo,
	path tree.NodePath,
) (*latticev1.Service, error) {
	labels := map[string]string{
		kubeconstants.LabelKeyServicePathDomain: path.ToDomain(true),
	}

	spec, err := c.serviceSpec(system, serviceInfo, path)
	if err != nil {
		return nil, err
	}

	systemNamespace := kubeutil.SystemNamespace(c.latticeID, v1.SystemID(system.Name))

	service := &latticev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			Namespace:       systemNamespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(system, controllerKind)},
			Labels:          labels,
		},
		Spec: spec,
		Status: latticev1.ServiceStatus{
			State: latticev1.ServiceStatePending,
		},
	}

	annotations, err := c.serviceMesh.ServiceAnnotations(service)
	if err != nil {
		return nil, err
	}

	service.Annotations = annotations

	return service, nil
}

func (c *Controller) serviceSpec(
	system *latticev1.System,
	serviceInfo *latticev1.SystemSpecServiceInfo,
	path tree.NodePath,
) (latticev1.ServiceSpec, error) {
	var numInstances int32
	if serviceInfo.Definition.Resources().NumInstances != nil {
		numInstances = *(serviceInfo.Definition.Resources().NumInstances)
	} else if serviceInfo.Definition.Resources().MinInstances != nil {
		numInstances = *(serviceInfo.Definition.Resources().MinInstances)
	} else {
		systemNamespace := kubeutil.SystemNamespace(c.latticeID, v1.SystemID(system.Name))
		err := fmt.Errorf(
			"System %v/%v Service %v invalid Service definition: num_instances or min_instances must be set",
			systemNamespace,
			system.Name,
			path,
		)
		return latticev1.ServiceSpec{}, err
	}

	componentPorts := map[string][]latticev1.ComponentPort{}

	for _, component := range serviceInfo.Definition.Components() {
		var ports []latticev1.ComponentPort
		for _, port := range component.Ports {
			componentPort := latticev1.ComponentPort{
				Name:     port.Name,
				Port:     int32(port.Port),
				Protocol: port.Protocol,
				Public:   false,
			}

			if port.ExternalAccess != nil && port.ExternalAccess.Public {
				componentPort.Public = true
			}

			ports = append(ports, componentPort)
		}

		componentPorts[component.Name] = ports
	}

	spec := latticev1.ServiceSpec{
		Path:                    path,
		Definition:              serviceInfo.Definition,
		ComponentBuildArtifacts: serviceInfo.ComponentBuildArtifacts,
		Ports:        componentPorts,
		NumInstances: numInstances,
	}
	return spec, nil
}

func (c *Controller) updateService(service *latticev1.Service, spec latticev1.ServiceSpec) (*latticev1.Service, error) {
	if reflect.DeepEqual(service.Spec, spec) {
		return service, nil
	}

	// Copy so the cache isn't mutated
	service = service.DeepCopy()
	service.Spec = spec

	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}
