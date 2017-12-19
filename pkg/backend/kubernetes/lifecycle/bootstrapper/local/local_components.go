package local

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
)

func (b *DefaultBootstrapper) seedDNS() ([]interface{}, error) {
	if !b.Options.DryRun {
		fmt.Println("Seeding local DNS server")
	}

	// TODO :: Handle namespace
	namespace := kubeutil.InternalNamespace("lattice")

	controller_args := []string{"--provider", b.Provider, "--cluster-id", string(b.ClusterID)}
	controller_args = append(controller_args, b.Options.LocalComponents.LocalDNSController.Args...)

	server_args := []string{}
	server_args = append(server_args, b.Options.LocalComponents.LocalDNSServer.Args...)

	labels := map[string]string{
		"key" : constants.MasterNodeDNSServer,
	}

	// Create a daemon set for my image
	localDNSDaemonSet := &appsv1beta2.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1beta2.GroupName + "/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:		constants.MasterNodeDNSServer,
			Namespace: 	namespace,
			Labels:		labels,
		},
		Spec: appsv1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.MasterNodeDNSServer,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.MasterNodeDNSSController,
							Image: b.Options.LocalComponents.LocalDNSController.Image,
							Args:  controller_args,
						},
						{
							Name:	constants.MasterNodeDNSServer,
							Image:	b.Options.LocalComponents.LocalDNSServer.Image,
							Args:	server_args,
							// TODO :: Ports
							//Ports: []corev1.ContainerPort{
							//	{
							//		ContainerPort: 53,
							//		Name: "dns",
							//		Protocol: "UDP",
							//	},
							//	{
							//		ContainerPort: 53,
							//		Name: "dns-tcp",
							//		Protocol: "TCP",
							//	},
							//},
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					// TODO :: This is default until I know what SA, if any, to use for the DNS.
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

	if b.Options.DryRun {
		return []interface{}{localDNSDaemonSet}, nil
	}

	localDNSDaemonSet, err := b.KubeClient.AppsV1beta2().DaemonSets(namespace).Create(localDNSDaemonSet)
	return []interface{}{localDNSDaemonSet}, err
}
