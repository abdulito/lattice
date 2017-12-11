package base

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *DefaultBootstrapper) seedConfig() ([]interface{}, error) {
	namespace := kubeutil.GetFullNamespace(b.Options.Config.KubernetesNamespacePrefix, kubeconstants.NamespaceLatticeInternal)

	// Create config
	config := &crv1.Config{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Config",
			APIVersion: crv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ConfigGlobal,
			Namespace: namespace,
		},
		Spec: b.Options.Config,
	}

	if b.Options.DryRun {
		return []interface{}{config}, nil
	}

	fmt.Println("Seeding base lattice config")

	config, err := b.LatticeClient.LatticeV1().Configs(namespace).Create(config)
	return []interface{}{config}, err
}
