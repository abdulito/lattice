package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

func seedLatticeSystemEnvironmentManagerAPI() {
	fmt.Println("Seeding lattice-system-environment-manager...")

	latticeSystemEnvironmentManagerAPIDaemonSet := &extensionsv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeComponentManagerAPI,
			Namespace: constants.NamespaceLatticeInternal,
		},
		Spec: extensionsv1beta1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.MasterNodeComponentManagerAPI,
					Labels: map[string]string{
						constants.MasterNodeLabelComponent: constants.MasterNodeComponentManagerAPI,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "api",
							Image: getContainerImageFQN(constants.DockerImageManagerAPIRest),
							Args:  []string{"-port", "80"},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      80,
									ContainerPort: 80,
								},
							},
						},
					},
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountManagementAPI,
					// Can tolerate the master-node taint in the local case when it's not applied harmlessly
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
				},
			},
		},
	}

	// FIXME: add NodeSelector for cloud providers
	//switch coretypes.Provider(provider) {
	//case coreconstants.ProviderLocal:
	//
	//default:
	//	panic("unsupported provider")
	//}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClient.
			ExtensionsV1beta1().
			DaemonSets(string(constants.NamespaceLatticeInternal)).
			Create(latticeSystemEnvironmentManagerAPIDaemonSet)
	})
}
