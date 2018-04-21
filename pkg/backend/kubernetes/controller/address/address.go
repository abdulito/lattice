package address

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncAddressStatus(
	address *latticev1.Address,
	endpoint *latticev1.Endpoint,
) (*latticev1.Address, error) {
	var state latticev1.AddressState
	switch endpoint.Status.State {
	case latticev1.EndpointStatePending:
		state = latticev1.ServiceAddressStatePending

	case latticev1.EndpointStateCreated:
		state = latticev1.ServiceAddressStateCreated

	case latticev1.EndpointStateFailed:
		state = latticev1.ServiceAddressStateFailed

	default:
		return nil, fmt.Errorf("Endpoint %v/%v in unexpected state %v", endpoint.Namespace, endpoint.Name, endpoint.Status.State)
	}

	return c.updateServiceStatus(address, state)
}

func (c *Controller) updateServiceStatus(
	address *latticev1.Address,
	state latticev1.AddressState,
) (*latticev1.Address, error) {
	status := latticev1.AddressStatus{
		State:              state,
		ObservedGeneration: address.Generation,
	}

	if reflect.DeepEqual(address.Status, status) {
		return address, nil
	}

	// Copy the address so the shared cache isn't mutated
	address = address.DeepCopy()
	address.Status = status

	return c.latticeClient.LatticeV1().Addresses(address.Namespace).Update(address)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().Services(address.Namespace).UpdateStatus(address)
}
