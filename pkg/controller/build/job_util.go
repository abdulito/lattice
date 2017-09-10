package build

import (
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	providerutils "github.com/mlab-lattice/core/pkg/provider"
	"github.com/mlab-lattice/core/pkg/system/definition/block"
)

const (
	jobLocalWorkingDirectoryVolumePathPrefix = "/data/builder"
	jobWorkingDirectory                      = "/var/run/builder"
	jobWorkingDirectoryVolumeName            = "workdir"

	jobDockerSocketVolumePath = "/var/run/docker.sock"
	jobDockerSocketPath       = "/var/run/docker.sock"
	jobDockerSocketVolumeName = "dockersock"

	jobDockerFqnAnnotationKey = "docker-image-fqn"
)

func getBuildJobName(b *crv1.Build) string {
	return fmt.Sprintf("lattice-build-%s", b.Name)
}

func (bc *BuildController) getBuildJob(b *crv1.Build) *batchv1.Job {
	// Need a consistent view of our config while generating the Job
	bc.configLock.RLock()
	defer bc.configLock.RUnlock()

	name := getBuildJobName(b)

	// FIXME: get job spec for build.DockerImage as well
	jobSpec, dockerImageFqn := bc.getGitRepositoryBuildJobSpec(b)

	labels := map[string]string{
		"mlab.lattice.com/build": "true",
	}
	annotations := map[string]string{
		jobDockerFqnAnnotationKey: dockerImageFqn,
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Annotations:     annotations,
			Labels:          labels,
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(b, controllerKind)},
		},
		Spec: jobSpec,
	}

	return job
}

func (bc *BuildController) getGitRepositoryBuildJobSpec(build *crv1.Build) (batchv1.JobSpec, string) {
	pullGitRepoContainer := bc.getPullGitRepoContainer(build)
	authorizeDockerContainer := bc.getAuthorizeDockerContainer()
	buildDockerImageContainer, dockerImageFqn := bc.getBuildDockerImageContainer(build)
	name := getBuildJobName(build)

	var workingDirectoryVolumeSource corev1.VolumeSource
	switch bc.provider {
	case providerutils.Local:
		workingDirectoryVolumeSource = corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: fmt.Sprintf("%v/%v", jobLocalWorkingDirectoryVolumePathPrefix, name),
			},
		}
	default:
		panic("unreachable")
	}

	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						Name:         jobWorkingDirectoryVolumeName,
						VolumeSource: workingDirectoryVolumeSource,
					},
					{
						Name: jobDockerSocketVolumeName,
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: jobDockerSocketVolumePath,
							},
						},
					},
				},
				InitContainers: []corev1.Container{
					pullGitRepoContainer,
				},
				Containers: []corev1.Container{
					buildDockerImageContainer,
				},
				// TODO: add failure policy once it is supported: https://github.com/kubernetes/kubernetes/issues/30243
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}

	if authorizeDockerContainer != nil {
		jobSpec.Template.Spec.InitContainers = append(
			jobSpec.Template.Spec.InitContainers,
			*authorizeDockerContainer,
		)
	}

	return jobSpec, dockerImageFqn
}

func (bc *BuildController) getPullGitRepoContainer(build *crv1.Build) corev1.Container {
	pullGitRepoContainer := corev1.Container{
		Name:    "pull-git-repo",
		Image:   bc.config.PullGitRepoImage,
		Command: []string{"./pull-git-repo.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "WORK_DIR",
				Value: jobWorkingDirectory,
			},
			{
				Name:  "GIT_URL",
				Value: build.Spec.BuildDefinitionBlock.GitRepository.Url,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      jobWorkingDirectoryVolumeName,
				MountPath: jobWorkingDirectory,
			},
		},
	}

	if build.Spec.BuildDefinitionBlock.GitRepository.Commit != nil {
		pullGitRepoContainer.Env = append(
			pullGitRepoContainer.Env,
			corev1.EnvVar{
				Name:  "GIT_CHECKOUT_TARGET",
				Value: *build.Spec.BuildDefinitionBlock.GitRepository.Commit,
			},
		)
	} else {
		pullGitRepoContainer.Env = append(
			pullGitRepoContainer.Env,
			corev1.EnvVar{
				Name:  "GIT_CHECKOUT_TARGET",
				Value: *build.Spec.BuildDefinitionBlock.GitRepository.Tag,
			},
		)
	}

	return pullGitRepoContainer
}

func (bc *BuildController) getAuthorizeDockerContainer() *corev1.Container {
	switch bc.provider {
	case providerutils.AWS:
		authorizeEcrContainer := bc.getAuthorizeEcrContainer()
		return &authorizeEcrContainer
	case providerutils.Local:
		return nil
	default:
		panic("unreachable")
	}
}

func (bc *BuildController) getAuthorizeEcrContainer() corev1.Container {
	return corev1.Container{
		Name:    "get-ecr-creds",
		Image:   bc.config.AuthorizeDockerImage,
		Command: []string{"./get-ecr-creds.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "WORK_DIR",
				Value: jobWorkingDirectory,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      jobWorkingDirectoryVolumeName,
				MountPath: jobWorkingDirectory,
			},
		},
	}
}

func (bc *BuildController) getBuildDockerImageContainer(build *crv1.Build) (corev1.Container, string) {
	buildDockerImageContainer := corev1.Container{
		Name:    "build-docker-image",
		Image:   bc.config.BuildDockerImage,
		Command: []string{"./build-docker-image.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "WORK_DIR",
				Value: jobWorkingDirectory,
			},
			{
				Name:  "DOCKER_REGISTRY",
				Value: bc.config.DockerConfig.Registry,
			},
			{
				Name:  "BUILD_CMD",
				Value: *build.Spec.BuildDefinitionBlock.Command,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      jobWorkingDirectoryVolumeName,
				MountPath: jobWorkingDirectory,
			},
			{
				Name:      jobDockerSocketVolumeName,
				MountPath: jobDockerSocketPath,
			},
		},
	}

	repo := bc.config.DockerConfig.Repository
	tag := build.Name
	if bc.config.DockerConfig.RepositoryPerImage {
		repo = build.Name
		tag = fmt.Sprint(time.Now().Unix())
	}

	dockerImageFqn := fmt.Sprintf(
		"%v/%v:%v",
		bc.config.DockerConfig.Registry,
		repo,
		tag,
	)

	buildDockerImageContainer.Env = append(
		buildDockerImageContainer.Env,
		// TODO: should this be Namespace/Name? should builds be namespaced?
		corev1.EnvVar{
			Name:  "DOCKER_REPOSITORY",
			Value: repo,
		},
		corev1.EnvVar{
			Name:  "DOCKER_IMAGE_TAG",
			Value: tag,
		},
	)

	push := "0"
	if bc.config.DockerConfig.Push {
		push = "1"
	}
	buildDockerImageContainer.Env = append(
		buildDockerImageContainer.Env,
		corev1.EnvVar{
			Name:  "DOCKER_PUSH",
			Value: push,
		},
	)

	var baseImage string
	if build.Spec.BuildDefinitionBlock.Language != nil {
		// TODO: insert custom language images when we have them
		baseImage = *build.Spec.BuildDefinitionBlock.Language
	} else {
		baseImage = getDockerImageFqn(build.Spec.BuildDefinitionBlock.DockerImage)
	}
	buildDockerImageContainer.Env = append(
		buildDockerImageContainer.Env,
		corev1.EnvVar{
			Name:  "DOCKER_BASE_IMAGE",
			Value: baseImage,
		},
	)

	return buildDockerImageContainer, dockerImageFqn
}

func getDockerImageFqn(dockerImage *block.DockerImage) string {
	return fmt.Sprintf("%v/%v:%v", dockerImage.Registry, dockerImage.Repository, dockerImage.Tag)
}

func jobStatus(j *batchv1.Job) (finished bool, succeeded bool) {
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			finished = true
			succeeded = true
			return
		}
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			finished = true
			return
		}
	}
	return
}
