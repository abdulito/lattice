package base

import (
	"fmt"
	"strconv"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
)

func (b *DefaultBootstrapper) seedMasterComponents() error {
	fmt.Println("Seeding master components")

	seedMasterComponentFuncs := []func() error{
		b.seedLatticeControllerManager,
		b.seedManagerAPI,
	}

	for _, seedFunc := range seedMasterComponentFuncs {
		if err := seedFunc(); err != nil {
			return err
		}
	}
	return nil
}

func (b *DefaultBootstrapper) seedLatticeControllerManager() error {
	fmt.Println("Seeding lattice-controller-manager")

	// TODO: for now we'll make a DaemonSet that runs on all the master nodes (aka all nodes in local)
	//		 and rely on the fact that we don't support multiple master nodes on AWS yet. Eventually we'll
	//		 either have to figure out how to have multiple lattice-controller-managers running (e.g. use leaderelect
	//		 in client-go) or find the best way to ensure there's at most one version of something running (maybe
	//		 StatefulSets?).
	namespace := kubeutil.GetFullNamespace(b.Options.Config.KubernetesNamespacePrefix, constants.NamespaceLatticeInternal)
	args := []string{"--provider", b.Provider}
	args = append(args, b.Options.MasterComponents.LatticeControllerManager.Args...)
	labels := map[string]string{
		constants.MasterNodeLabelComponent: constants.MasterNodeComponentLatticeControllerManager,
	}

	latticeControllerManagerDaemonSet := &appsv1beta2.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeComponentLatticeControllerManager,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.MasterNodeComponentLatticeControllerManager,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.MasterNodeComponentLatticeControllerManager,
							Image: b.Options.MasterComponents.LatticeControllerManager.Image,
							Args:  args,
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLatticeControllerManager,
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &constants.NodeAffinityMasterNode,
					},
				},
			},
		},
	}

	_, err := b.KubeClient.AppsV1beta2().DaemonSets(namespace).Create(latticeControllerManagerDaemonSet)
	return err
}

func (b *DefaultBootstrapper) seedManagerAPI() error {
	fmt.Println("Seeding manager-api")

	// TODO: for now we'll make a DaemonSet that runs on all the master nodes (aka all nodes in local)
	//		 and rely on the fact that we don't support multiple master nodes on AWS yet. Eventually we'll
	//		 either have to figure out how to have multiple lattice-controller-managers running (e.g. use leaderelect
	//		 in client-go) or find the best way to ensure there's at most one version of something running (maybe
	//		 StatefulSets?).
	namespace := kubeutil.GetFullNamespace(b.Options.Config.KubernetesNamespacePrefix, constants.NamespaceLatticeInternal)
	args := []string{"--port", strconv.Itoa(int(b.Options.MasterComponents.ManagerAPI.Port))}
	args = append(args, b.Options.MasterComponents.ManagerAPI.Args...)
	labels := map[string]string{
		constants.MasterNodeLabelComponent: constants.MasterNodeComponentLatticeControllerManager,
	}

	latticeControllerManagerDaemonSet := &appsv1beta2.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeComponentManagerAPI,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.MasterNodeComponentManagerAPI,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.MasterNodeComponentManagerAPI,
							Image: b.Options.MasterComponents.ManagerAPI.Image,
							Args:  args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      b.Options.MasterComponents.ManagerAPI.Port,
									ContainerPort: b.Options.MasterComponents.ManagerAPI.Port,
								},
							},
						},
					},
					HostNetwork:        b.Options.MasterComponents.ManagerAPI.HostNetwork,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLatticeControllerManager,
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &constants.NodeAffinityMasterNode,
					},
				},
			},
		},
	}

	_, err := b.KubeClient.AppsV1beta2().DaemonSets(namespace).Create(latticeControllerManagerDaemonSet)
	return err
}
