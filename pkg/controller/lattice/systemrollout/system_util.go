package systemrollout

import (
	"fmt"
	"reflect"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (src *SystemRolloutController) getNewSystem(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.System, error) {
	sysSpec, err := src.getNewSystemSpec(sysRollout, sysBuild)
	if err != nil {
		return nil, err
	}

	sys := &crv1.System{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(sysBuild.Spec.LatticeNamespace),
		},
		Spec: *sysSpec,
		Status: crv1.SystemStatus{
			State: crv1.SystemStateRollingOut,
		},
	}
	return sys, nil
}

func (src *SystemRolloutController) getNewSystemSpec(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.SystemSpec, error) {
	root, err := systemtree.NewNode(systemdefinition.Interface(&sysBuild.Spec.Definition), nil)
	if err != nil {
		return nil, err
	}

	// Create crv1.SystemServicesInfo for each service in the sysBuild.Spec.Definition
	services := map[systemtree.NodePath]crv1.SystemServicesInfo{}
	for path, service := range root.Services() {
		svcBuildInfo, ok := sysBuild.Spec.Services[path]
		if !ok {
			// FIXME: send warn event
			return nil, fmt.Errorf("SystemBuild does not have expected Service %v", path)
		}

		svcBuild, err := src.getSvcBuild(*svcBuildInfo.BuildName)
		if err != nil {
			return nil, err
		}

		// Create crv1.ComponentBuildArtifacts for each Component in the Service
		cBuildArtifacts := map[string]crv1.ComponentBuildArtifacts{}
		for component, cBuildInfo := range svcBuild.Spec.Components {
			if cBuildInfo.BuildName == nil {
				// FIXME: send warn event
				return nil, fmt.Errorf("svcBuild %v Component %v does not have a ComponentBuildName", svcBuild.Name, component)
			}

			cBuildName := *cBuildInfo.BuildName
			cBuildKey := svcBuild.Namespace + "/" + cBuildName
			cBuildObj, exists, err := src.componentBuildStore.GetByKey(cBuildKey)

			if err != nil {
				return nil, err
			}

			if !exists {
				// FIXME: send warn event
				return nil, fmt.Errorf("cBuild %v not in cBuild Store", cBuildKey)
			}

			cBuild := cBuildObj.(*crv1.ComponentBuild)

			if cBuild.Spec.Artifacts == nil {
				// FIXME: send warn event
				return nil, fmt.Errorf("cBuild %v does not have Artifacts", cBuildKey)
			}
			cBuildArtifacts[component] = *cBuild.Spec.Artifacts
		}

		services[path] = crv1.SystemServicesInfo{
			Definition:              *(service.Definition().(*systemdefinition.Service)),
			ComponentBuildArtifacts: cBuildArtifacts,
		}
	}

	sysSpec := &crv1.SystemSpec{
		Services: services,
	}

	return sysSpec, nil
}

func (src *SystemRolloutController) getSvcBuild(svcBuildName string) (*crv1.ServiceBuild, error) {
	svcBuildKey := constants.InternalNamespace + "/" + svcBuildName
	svcBuildObj, exists, err := src.serviceBuildStore.GetByKey(svcBuildKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("ServiceBuild %v is not in ServiceBuild Store", svcBuildKey)
	}

	svcBuild := svcBuildObj.(*crv1.ServiceBuild)
	return svcBuild, nil
}

func (src *SystemRolloutController) createSystem(sysRollout *crv1.SystemRollout, sysBuild *crv1.SystemBuild) (*crv1.System, error) {
	sys, err := src.getNewSystem(sysRollout, sysBuild)
	if err != nil {
		return nil, err
	}

	result := &crv1.System{}
	err = src.latticeResourceClient.Post().
		Namespace(string(sysRollout.Spec.LatticeNamespace)).
		Resource(crv1.SystemResourcePlural).
		Body(sys).
		Do().
		Into(result)
	return result, err
}

func (src *SystemRolloutController) updateSystem(sys *crv1.System, sysSpec *crv1.SystemSpec) (*crv1.System, error) {
	if reflect.DeepEqual(sys.Spec, sysSpec) {
		return sys, nil
	}

	sys.Spec = *sysSpec
	sys.Status.State = crv1.SystemStateRollingOut

	result := &crv1.System{}
	err := src.latticeResourceClient.Put().
		Namespace(sys.Namespace).
		Resource(crv1.SystemResourcePlural).
		Name(sys.Name).
		Body(sys).
		Do().
		Into(result)

	return result, err
}
