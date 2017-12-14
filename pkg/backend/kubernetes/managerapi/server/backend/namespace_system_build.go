package backend

import (
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) BuildSystem(ln types.LatticeNamespace, definitionRoot tree.Node, v types.SystemVersion) (types.SystemBuildID, error) {
	systemBuild, err := getNewSystemBuild(ln, definitionRoot, v)
	if err != nil {
		return "", err
	}

	result, err := kb.LatticeClient.LatticeV1().SystemBuilds(kubeconstants.NamespaceLatticeInternal).Create(systemBuild)
	if err != nil {
		return "", err
	}

	return types.SystemBuildID(result.Name), err
}

func getNewSystemBuild(ln types.LatticeNamespace, definitionRoot tree.Node, v types.SystemVersion) (*crv1.SystemBuild, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel: string(ln),
		kubeconstants.LabelKeySystemVersion: string(v),
	}

	services := map[tree.NodePath]crv1.SystemBuildSpecServiceInfo{}
	for path, svcNode := range definitionRoot.Services() {
		services[path] = crv1.SystemBuildSpecServiceInfo{
			Definition: *(svcNode.Definition().(*definition.Service)),
		}
	}

	sysB := &crv1.SystemBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: crv1.SystemBuildSpec{
			LatticeNamespace: ln,
			DefinitionRoot:   definitionRoot,
			Services:         services,
		},
		Status: crv1.SystemBuildStatus{
			State: crv1.SystemBuildStatePending,
		},
	}

	return sysB, nil
}

func (kb *KubernetesBackend) ListSystemBuilds(ln types.LatticeNamespace) ([]types.SystemBuild, error) {
	result, err := kb.LatticeClient.LatticeV1().SystemBuilds(kubeconstants.NamespaceLatticeInternal).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	builds := []types.SystemBuild{}
	for _, build := range result.Items {
		// TODO: add this to the query
		if strings.Compare(build.Labels[kubeconstants.LatticeNamespaceLabel], string(ln)) != 0 {
			continue
		}

		builds = append(builds, transformSystemBuild(&build))
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetSystemBuild(ln types.LatticeNamespace, bid types.SystemBuildID) (*types.SystemBuild, bool, error) {
	build, exists, err := kb.getInternalSystemBuild(ln, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	coreBuild := transformSystemBuild(build)
	return &coreBuild, true, nil
}

func (kb *KubernetesBackend) getInternalSystemBuild(ln types.LatticeNamespace, bid types.SystemBuildID) (*crv1.SystemBuild, bool, error) {
	result, err := kb.LatticeClient.LatticeV1().SystemBuilds(kubeconstants.NamespaceLatticeInternal).Get(string(bid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// TODO: add this to the query
	if strings.Compare(result.Labels[kubeconstants.LatticeNamespaceLabel], string(ln)) != 0 {
		return nil, false, nil
	}

	return result, true, nil
}

func transformSystemBuild(build *crv1.SystemBuild) types.SystemBuild {
	sysb := types.SystemBuild{
		ID:            types.SystemBuildID(build.Name),
		State:         getSystemBuildState(build.Status.State),
		Version:       types.SystemVersion(build.Labels[kubeconstants.LabelKeySystemVersion]),
		ServiceBuilds: map[tree.NodePath]*types.ServiceBuild{},
	}

	for service, svcbInfo := range build.Spec.Services {
		if svcbInfo.Name == nil {
			sysb.ServiceBuilds[service] = nil
			continue
		}
		id := types.ServiceBuildID(*svcbInfo.Name)

		state := constants.ServiceBuildStatePending
		if svcbInfo.State != nil {
			state = getServiceBuildState(*svcbInfo.State)
		}

		svcb := &types.ServiceBuild{
			ID:              id,
			State:           state,
			ComponentBuilds: map[string]*types.ComponentBuild{},
		}

		for component, cbInfo := range svcbInfo.Components {
			if cbInfo.Name == nil {
				svcb.ComponentBuilds[component] = nil
				continue
			}
			id := types.ComponentBuildID(*cbInfo.Name)

			state := constants.ComponentBuildStatePending
			if cbInfo.Status != nil {
				state = getComponentBuildState(*cbInfo.Status)
			}

			var failureMessage *string
			if cbInfo.FailureInfo != nil {
				failMessage := getComponentBuildFailureMessage(*cbInfo.FailureInfo)
				failureMessage = &failMessage
			}

			svcb.ComponentBuilds[component] = &types.ComponentBuild{
				ID:                id,
				State:             state,
				LastObservedPhase: cbInfo.LastObservedPhase,
				FailureMessage:    failureMessage,
			}
		}

		sysb.ServiceBuilds[service] = svcb
	}

	return sysb
}

func getSystemBuildState(state crv1.SystemBuildState) types.SystemBuildState {
	switch state {
	case crv1.SystemBuildStatePending:
		return constants.SystemBuildStatePending
	case crv1.SystemBuildStateRunning:
		return constants.SystemBuildStateRunning
	case crv1.SystemBuildStateSucceeded:
		return constants.SystemBuildStateSucceeded
	case crv1.SystemBuildStateFailed:
		return constants.SystemBuildStateFailed
	default:
		panic("unreachable")
	}
}
