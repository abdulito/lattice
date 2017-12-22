package main

import (
	"flag"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v1/pernode"
	"github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1/rest"
)

var (
	kubeconfig string
	namespace  string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&namespace, "namespace", "", "namespace the api manages")
	flag.Parse()
}

func main() {
	backend, err := pernode.NewKubernetesPerNodeBackend(kubeconfig, namespace)
	if err != nil {
		panic(err)
	}

	rest.RunNewRestServer(backend, 8080)
}
