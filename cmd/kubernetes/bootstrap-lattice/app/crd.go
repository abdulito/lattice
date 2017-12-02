package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/kubernetes/customresource"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/client-go/rest"
)

func seedCrds(kubeconfig *rest.Config) {
	fmt.Println("Seeding CRDs...")

	apiextensionsclientset, err := apiextensionsclient.NewForConfig(kubeconfig)
	if err != nil {
		panic(err)
	}
	_, err = customresource.CreateCustomResourceDefinitions(apiextensionsclientset)
	if err != nil {
		panic(err)
	}
}
