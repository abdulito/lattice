package main

import (
	"fmt"

	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/client-go/kubernetes"
)

func seedLatticeControllerManager(kubeClientset *kubernetes.Clientset) {
	fmt.Println("Seeding lattice-controller-manager...")

	// TODO: for now we'll make a DaemonSet that runs on all the master nodes (aka all nodes in local)
	//		 and rely on the fact that we don't support multiple master nodes on AWS yet. Eventually we'll
	//		 either have to figure out how to have multiple lattice-controller-managers running (e.g. use leaderelect
	//		 in client-go) or find the best way to ensure there's at most one version of something running (maybe
	//		 StatefulSets?).
	latticeControllerManagerDaemonSet := &extensionsv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lattice-controller-manager",
			Namespace: constants.NamespaceLatticeInternal,
		},
		Spec: extensionsv1beta1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lattice-controller-manager",
					Labels: map[string]string{
						"master.lattice.mlab.com/controller-manager": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "controller-manager",
							Image:           latticeContainerRegistry + "/lattice-controller-manager",
							Command:         []string{"/app/cmd/controller-manager/go_image.binary"},
							Args:            []string{"-v", "5", "-logtostderr", "-provider", provider},
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLatticeControllerManager,
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
		return kubeClientset.
			ExtensionsV1beta1().
			DaemonSets(string(constants.NamespaceLatticeInternal)).
			Create(latticeControllerManagerDaemonSet)
	})
}
