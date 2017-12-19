package local

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

func NewLocalCloudProvider() *DefaultLocalCloudProvider {
	return &DefaultLocalCloudProvider{}
}

type DefaultLocalCloudProvider struct {
}

func (cp *DefaultLocalCloudProvider) TransformComponentBuildJobSpec(spec *batchv1.JobSpec) *batchv1.JobSpec {
	spec.Template.Spec.Affinity = nil
	return spec
}

func (cp *DefaultLocalCloudProvider) TransformServiceDeploymentSpec(service *crv1.Service, spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	spec.Template.Spec.Affinity = nil
	return spec
}

func (sm *DefaultLocalCloudProvider) IsDeploymentSpecUpdated(
	service *crv1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string, *appsv1.DeploymentSpec) {
	return true, "", nil
}
