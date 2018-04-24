package backend

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListServices(systemID v1.SystemID) ([]v1.Service, error) {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	services, err := kb.latticeClient.LatticeV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalServices []v1.Service
	for _, service := range services.Items {
		// FIXME: should use service.PathLabel()
		servicePath, err := tree.NodePathFromDomain(service.Name)
		if err != nil {
			return nil, err
		}

		externalService, err := kb.transformService(servicePath, &service.Status)
		if err != nil {
			return nil, err
		}

		externalServices = append(externalServices, externalService)
	}

	return externalServices, nil
}

func (kb *KubernetesBackend) GetService(systemID v1.SystemID, path tree.NodePath) (*v1.Service, error) {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	service, err := kb.latticeClient.LatticeV1().Services(namespace).Get(path.ToDomain(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	servicePath, err := tree.NodePathFromDomain(service.Name)
	if err != nil {
		return nil, err
	}
	externalService, err := kb.transformService(servicePath, &service.Status)
	if err != nil {
		return nil, err
	}

	return &externalService, nil
}

func (kb *KubernetesBackend) transformService(path tree.NodePath, status *latticev1.ServiceStatus) (v1.Service, error) {
	state, err := getServiceState(status.State)
	if err != nil {
		return v1.Service{}, err
	}

	service := v1.Service{
		Path: path,

		State:  state,
		Reason: status.Reason,

		UpdatedInstances: status.UpdatedInstances,
		StaleInstances:   status.StaleInstances,

		Ports: status.Ports,
	}

	var failureMessage *string
	if status.FailureInfo != nil {
		internalError := "internal error"
		failureMessage = &internalError

		if !status.FailureInfo.Internal {
			errorMessage := fmt.Sprintf("%v: %v", status.FailureInfo.Time, status.FailureInfo.Message)
			failureMessage = &errorMessage
		}
	}
	service.FailureMessage = failureMessage

	return service, nil
}

func getServiceState(state latticev1.ServiceState) (v1.ServiceState, error) {
	switch state {
	case latticev1.ServiceStatePending:
		return v1.ServiceStatePending, nil
	case latticev1.ServiceStateScaling:
		return v1.ServiceStateScaling, nil
	case latticev1.ServiceStateUpdating:
		return v1.ServiceStateUpdating, nil
	case latticev1.ServiceStateStable:
		return v1.ServiceStateStable, nil
	case latticev1.ServiceStateFailed:
		return v1.ServiceStateFailed, nil
	default:
		return "", fmt.Errorf("invalid service state: %v", state)
	}
}
