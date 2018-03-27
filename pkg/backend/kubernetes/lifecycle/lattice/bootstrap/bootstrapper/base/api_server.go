package base

import (
	"fmt"
	"strconv"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func (b *DefaultBootstrapper) aPIServerResources(resources *bootstrapper.Resources) {
	internalNamespace := kubeutil.InternalNamespace(b.LatticeID)
	name := fmt.Sprintf("%v-%v", b.LatticeID, kubeconstants.ControlPlaneServiceAPIServer)

	clusterRole := &rbacv1.ClusterRole{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []rbacv1.PolicyRule{
			// lattice system read and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralSystem},
				Verbs:     readCreateAndDeleteVerbs,
			},
			// lattice config read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralConfig},
				Verbs:     readVerbs,
			},
			// lattice system build read and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralBuild},
				Verbs:     readAndCreateVerbs,
			},
			// lattice service build read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralServiceBuild},
				Verbs:     readVerbs,
			},
			// lattice component build read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralComponentBuild},
				Verbs:     readVerbs,
			},
			// lattice rollout build and create
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralDeploy},
				Verbs:     readAndCreateVerbs,
			},
			// lattice service read
			{
				APIGroups: []string{latticev1.GroupName},
				Resources: []string{latticev1.ResourcePluralService},
				Verbs:     readVerbs,
			},

			// kube pod read and delete
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"pods"},
				Verbs:     readAndDeleteVerbs,
			},
			// kube pod/log read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"pods/log"},
				Verbs:     readVerbs,
			},
			// kube job read
			{
				APIGroups: []string{batchv1.GroupName},
				Resources: []string{"jobs"},
				Verbs:     readVerbs,
			},
			// kube service read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"services"},
				Verbs:     readVerbs,
			},
			// kube node read
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"nodes"},
				Verbs:     readVerbs,
			},
			// kube secret
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"secrets"},
				Verbs:     readCreateUpdateAndDeleteVerbs,
			},
		},
	}
	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)

	serviceAccount := &corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountAPIServer,
			Namespace: internalNamespace,
		},
	}
	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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
			Name:     clusterRole.Name,
		},
	}
	resources.ClusterRoleBindings = append(resources.ClusterRoleBindings, clusterRoleBinding)

	args := []string{
		"--port", strconv.Itoa(int(b.Options.MasterComponents.APIServer.Port)),
		"--lattice-id", string(b.LatticeID),
	}
	args = append(args, b.Options.MasterComponents.APIServer.Args...)
	labels := map[string]string{
		kubeconstants.LabelKeyControlPlaneService: kubeconstants.ControlPlaneServiceAPIServer,
	}

	daemonSet := &appsv1.DaemonSet{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ControlPlaneServiceAPIServer,
			Namespace: internalNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   kubeconstants.ControlPlaneServiceAPIServer,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  kubeconstants.ControlPlaneServiceAPIServer,
							Image: b.Options.MasterComponents.APIServer.Image,
							Args:  args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      b.Options.MasterComponents.APIServer.Port,
									ContainerPort: b.Options.MasterComponents.APIServer.Port,
								},
							},
						},
					},
					HostNetwork:        b.Options.MasterComponents.APIServer.HostNetwork,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: kubeconstants.ServiceAccountAPIServer,
					Tolerations: []corev1.Toleration{
						kubeconstants.TolerationMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &kubeconstants.NodeAffinityMasterNode,
					},
				},
			},
		},
	}
	resources.DaemonSets = append(resources.DaemonSets, daemonSet)
}
