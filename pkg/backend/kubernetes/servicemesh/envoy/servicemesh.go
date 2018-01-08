package envoy

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
	systembootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	envoyConfigDirectory           = "/etc/envoy"
	envoyConfigDirectoryVolumeName = kubeconstants.DeploymentResourcePrefixServiceMesh + "envoyconfig"

	initContainerNamePrepareEnvoy = kubeconstants.DeploymentResourcePrefixServiceMesh + "prepare-envoy"
	containerNameEnvoy            = kubeconstants.DeploymentResourcePrefixServiceMesh + "envoy"

	envoyXDSAPI         = "envoy-xds-api"
	labelKeyEnvoyXDSAPI = "envoy.servicemesh.lattice.mlab.com/xds-api"
)

func NewEnvoyServiceMesh(config *crv1.ConfigEnvoy) *DefaultEnvoyServiceMesh {
	return &DefaultEnvoyServiceMesh{
		Config: config,
	}
}

type DefaultEnvoyServiceMesh struct {
	Config *crv1.ConfigEnvoy
}

func (sm *DefaultEnvoyServiceMesh) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: envoyXDSAPI,
		},
		Rules: []rbacv1.PolicyRule{
			// Read kube endpoints
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"endpoints"},
				Verbs:     []string{"get", "watch", "list"},
			},
			// Read lattice services
			{
				APIGroups: []string{crv1.GroupName},
				Resources: []string{crv1.ResourcePluralService},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}

	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)
}

func (sm *DefaultEnvoyServiceMesh) BootstrapSystemResources(resources *systembootstrapper.SystemResources) {
	namespace := resources.Namespace.Name

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      envoyXDSAPI,
			Namespace: namespace,
		},
	}

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      envoyXDSAPI,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      serviceAccount.Name,
				Namespace: serviceAccount.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     envoyXDSAPI,
		},
	}

	labels := map[string]string{
		labelKeyEnvoyXDSAPI: "true",
	}

	daemonSet := &appsv1.DaemonSet{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      envoyXDSAPI,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   envoyXDSAPI,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "envoy-xds-api",
							Args: []string{
								"-v", "5",
								"-logtostderr",
								"-namespace", namespace,
							},
							Image: sm.Config.XDSAPIImage,
							Ports: []corev1.ContainerPort{
								{
									HostPort:      8080,
									ContainerPort: 8080,
								},
							},
						},
					},
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: serviceAccount.Name,
					Tolerations: []corev1.Toleration{
						kubeconstants.TolerationNodePool,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &kubeconstants.NodeAffinityNodePool,
					},
				},
			},
		},
	}

	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)
	resources.RoleBindings = append(resources.RoleBindings, roleBinding)
	resources.DaemonSets = append(resources.DaemonSets, daemonSet)
}

func (sm *DefaultEnvoyServiceMesh) TransformServiceDeploymentSpec(
	service *crv1.Service,
	spec *appsv1.DeploymentSpec,
	services []*crv1.Service,
) *appsv1.DeploymentSpec {
	prepareEnvoyContainer, envoyContainer := sm.envoyContainers(service)

	configVolume := corev1.Volume{
		Name: envoyConfigDirectoryVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	initContainers := []corev1.Container{prepareEnvoyContainer}
	initContainers = append(initContainers, spec.Template.Spec.InitContainers...)

	containers := []corev1.Container{envoyContainer}
	containers = append(containers, spec.Template.Spec.Containers...)

	volumes := []corev1.Volume{configVolume}
	volumes = append(volumes, spec.Template.Spec.Volumes...)

	// FIXME: remove this when local dns is working
	var hostnames []string
	for _, service := range services {
		hostnames = append(hostnames, service.Spec.Path.ToDomain(true))
	}
	sort.Strings(hostnames)

	ip, _, _ := net.ParseCIDR(sm.Config.RedirectCIDRBlock)

	hostAlias := corev1.HostAlias{
		IP:        ip.String(),
		Hostnames: hostnames,
	}
	spec.Template.Spec.HostAliases = []corev1.HostAlias{hostAlias}

	spec.Template.Spec.InitContainers = initContainers
	spec.Template.Spec.Containers = containers
	spec.Template.Spec.Volumes = volumes
	return spec
}

func (sm *DefaultEnvoyServiceMesh) IsDeploymentSpecUpdated(
	service *crv1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string, *appsv1.DeploymentSpec) {
	// make sure the init containers are correct
	updated, reason := checkExpectedContainers(current.Template.Spec.InitContainers, desired.Template.Spec.InitContainers, true)
	if !updated {
		return false, reason, nil
	}

	// make sure the containers are correct
	updated, reason = checkExpectedContainers(current.Template.Spec.Containers, desired.Template.Spec.Containers, false)
	if !updated {
		return false, reason, nil
	}

	// make sure the volumes are correct
	updated, reason = checkExpectedVolumes(current.Template.Spec.Volumes, desired.Template.Spec.Volumes)
	if !updated {
		return false, reason, nil
	}

	// get the init containers that are not a part of the serviceMesh
	var initContainers []corev1.Container
	for _, container := range current.Template.Spec.InitContainers {
		if isServiceMeshResource(container.Name) {
			continue
		}

		initContainers = append(initContainers, container)
	}

	// get the containers that are not a part of the serviceMesh
	var containers []corev1.Container
	for _, container := range current.Template.Spec.Containers {
		if isServiceMeshResource(container.Name) {
			continue
		}

		containers = append(containers, container)
	}

	// get the volumes that are not a part of the serviceMesh
	var volumes []corev1.Volume
	for _, volume := range current.Template.Spec.Volumes {
		if isServiceMeshResource(volume.Name) {
			continue
		}

		volumes = append(volumes, volume)
	}

	// make a copy of the desired spec, and set the initContainers, containers, and volumes
	// to be the slices without the service mesh resources
	spec := desired.DeepCopy()
	spec.Template.Spec.InitContainers = initContainers
	spec.Template.Spec.Containers = containers
	spec.Template.Spec.Volumes = volumes

	return true, "", spec
}

func checkExpectedContainers(currentContainers, desiredContainers []corev1.Container, init bool) (bool, string) {
	// Collect all of the expected containers
	desiredEnvoyContainers := map[string]corev1.Container{}
	for _, container := range desiredContainers {
		if !isServiceMeshResource(container.Name) {
			// not a service-mesh init container
			continue
		}

		desiredEnvoyContainers[container.Name] = container
	}

	containerType := ""
	if init {
		containerType = " init"
	}

	// Check to make sure all of the envoy containers exist
	currentEnvoyContainers := map[string]struct{}{}
	for _, container := range currentContainers {
		if !isServiceMeshResource(container.Name) {
			// not a service-mesh init container
			continue
		}

		desiredContainer, ok := desiredEnvoyContainers[container.Name]
		if !ok {
			return false, fmt.Sprintf("has extra envoy%v container %v", containerType, container.Name)
		}

		if !kubeutil.ContainersSemanticallyEqual(&container, &desiredContainer) {
			return false, fmt.Sprintf("has out of date envoy%v container %v", containerType, container.Name)
		}

		currentEnvoyContainers[container.Name] = struct{}{}
	}

	// Make sure there aren't extra containers
	numDesiredContainers := len(desiredEnvoyContainers)
	numCurrentContainers := len(currentEnvoyContainers)
	if numDesiredContainers != numCurrentContainers {
		return false, fmt.Sprintf("expected %v envoy%v containers, had %v", numDesiredContainers, containerType, numCurrentContainers)
	}

	return true, ""
}

func checkExpectedVolumes(currentVolumes, desiredVolumes []corev1.Volume) (bool, string) {
	// Collect all of the expected volumes
	desiredEnvoyVolumes := map[string]corev1.Volume{}
	for _, volume := range desiredVolumes {
		if !isServiceMeshResource(volume.Name) {
			// not a service-mesh init volume
			continue
		}

		desiredEnvoyVolumes[volume.Name] = volume
	}

	// Check to make sure all of the volumes exist
	currentEnvoyVolumes := map[string]struct{}{}
	for _, volume := range currentVolumes {
		if !isServiceMeshResource(volume.Name) {
			// not a service-mesh init volume
			continue
		}

		desiredVolume, ok := desiredEnvoyVolumes[volume.Name]
		if !ok {
			return false, fmt.Sprintf("has extra envoy volume %v", volume.Name)
		}

		if !kubeutil.VolumesSemanticallyEqual(&volume, &desiredVolume) {
			return false, fmt.Sprintf("has out of date envoy volume %v", volume.Name)
		}

		currentEnvoyVolumes[volume.Name] = struct{}{}
	}

	numDesiredVolumes := len(desiredEnvoyVolumes)
	numCurrentVolumes := len(currentEnvoyVolumes)
	if numDesiredVolumes != numCurrentVolumes {
		return false, fmt.Sprintf("expected %v envoy volumes, had %v", numDesiredVolumes, numCurrentVolumes)
	}

	return true, ""
}

func isServiceMeshResource(name string) bool {
	parts := strings.Split(name, kubeconstants.DeploymentResourcePrefixServiceMesh)
	return len(parts) >= 2
}

func (sm *DefaultEnvoyServiceMesh) envoyContainers(service *crv1.Service) (corev1.Container, corev1.Container) {
	prepareEnvoy := corev1.Container{
		Name:  initContainerNamePrepareEnvoy,
		Image: sm.Config.PrepareImage,
		Env: []corev1.EnvVar{
			{
				Name:  "EGRESS_PORT",
				Value: strconv.Itoa(int(service.Spec.EnvoyEgressPort)),
			},
			{
				Name:  "REDIRECT_EGRESS_CIDR_BLOCK",
				Value: sm.Config.RedirectCIDRBlock,
			},
			{
				Name:  "CONFIG_DIR",
				Value: envoyConfigDirectory,
			},
			{
				Name:  "ADMIN_PORT",
				Value: strconv.Itoa(int(service.Spec.EnvoyAdminPort)),
			},
			{
				Name: "XDS_API_HOST",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			},
			{
				Name:  "XDS_API_PORT",
				Value: fmt.Sprintf("%v", sm.Config.XDSAPIPort),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      envoyConfigDirectoryVolumeName,
				MountPath: envoyConfigDirectory,
			},
		},
		// Need CAP_NET_ADMIN to manipulate iptables
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"NET_ADMIN"},
			},
		},
	}

	var envoyPorts []corev1.ContainerPort
	for component, ports := range service.Spec.Ports {
		for _, port := range ports {
			envoyPorts = append(
				envoyPorts,
				corev1.ContainerPort{
					Name:          component + "-" + port.Name,
					ContainerPort: port.EnvoyPort,
				},
			)
		}
	}

	envoy := corev1.Container{
		Name:            containerNameEnvoy,
		Image:           sm.Config.Image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/local/bin/envoy"},
		Args: []string{
			"-c",
			fmt.Sprintf("%v/config.json", envoyConfigDirectory),
			"--service-cluster",
			service.Namespace,
			"--service-node",
			service.Spec.Path.ToDomain(false),
		},
		Ports: envoyPorts,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      envoyConfigDirectoryVolumeName,
				MountPath: envoyConfigDirectory,
				ReadOnly:  true,
			},
		},
	}

	return prepareEnvoy, envoy
}

func (sm *DefaultEnvoyServiceMesh) GetEndpointSpec(address *crv1.ServiceAddress) (*crv1.EndpointSpec, error) {
	ip, _, err := net.ParseCIDR(sm.Config.RedirectCIDRBlock)
	if err != nil {
		return nil, err
	}

	ipStr := ip.String()
	spec := &crv1.EndpointSpec{
		Path: address.Spec.Path,
		IP:   &ipStr,
	}
	return spec, nil
}
