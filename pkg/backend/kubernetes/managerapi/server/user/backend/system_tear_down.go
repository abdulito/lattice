package backend

import (
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) TearDownSystem(id types.SystemID) (types.SystemTeardownID, error) {
	systemTeardown, err := getSystemTeardown(id)
	if err != nil {
		return "", err
	}

	result, err := kb.LatticeClient.LatticeV1().SystemTeardowns(string(id)).Create(systemTeardown)
	if err != nil {
		return "", err
	}

	return types.SystemTeardownID(result.Name), err
}

func getSystemTeardown(id types.SystemID) (*crv1.SystemTeardown, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel: string(id),
	}

	sysT := &crv1.SystemTeardown{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: crv1.SystemTeardownSpec{},
		Status: crv1.SystemTeardownStatus{
			State: crv1.SystemTeardownStatePending,
		},
	}

	return sysT, nil
}

func (kb *KubernetesBackend) GetSystemTeardown(id types.SystemID, tid types.SystemTeardownID) (*types.SystemTeardown, bool, error) {
	result, err := kb.LatticeClient.LatticeV1().SystemTeardowns(string(id)).Get(string(tid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// TODO: add this to the query
	if strings.Compare(result.Labels[kubeconstants.LatticeNamespaceLabel], string(id)) != 0 {
		return nil, false, nil
	}

	sb := &types.SystemTeardown{
		ID:    tid,
		State: getSystemTeardownState(result.Status.State),
	}

	return sb, true, nil
}

func (kb *KubernetesBackend) ListSystemTeardowns(id types.SystemID) ([]types.SystemTeardown, error) {
	result, err := kb.LatticeClient.LatticeV1().SystemTeardowns(string(id)).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	teardowns := []types.SystemTeardown{}
	for _, b := range result.Items {
		// TODO: add this to the query
		if strings.Compare(b.Labels[kubeconstants.LatticeNamespaceLabel], string(id)) != 0 {
			continue
		}

		teardowns = append(teardowns, types.SystemTeardown{
			ID:    types.SystemTeardownID(b.Name),
			State: getSystemTeardownState(b.Status.State),
		})
	}

	return teardowns, nil
}

func getSystemTeardownState(state crv1.SystemTeardownState) types.SystemTeardownState {
	switch state {
	case crv1.SystemTeardownStatePending:
		return constants.SystemTeardownStatePending
	case crv1.SystemTeardownStateInProgress:
		return constants.SystemTeardownStateInProgress
	case crv1.SystemTeardownStateSucceeded:
		return constants.SystemTeardownStateSucceeded
	case crv1.SystemTeardownStateFailed:
		return constants.SystemTeardownStateFailed
	default:
		panic("unreachable")
	}
}
