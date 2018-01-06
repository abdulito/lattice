package main

import (
	"flag"
	"time"

	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
	dnsconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local/controller"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

var (
	kubeconfig        string
	hostsFilePath     string
	dnsmasqConfigPath string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&dnsmasqConfigPath, "dnsmasq-config-path", dnsconstants.DNSSharedConfigDirectory+dnsconstants.DnsmasqConfigFile, "path to the additional dnsmasq configuration file")
	flag.StringVar(&hostsFilePath, "hosts-file-path", dnsconstants.DNSSharedConfigDirectory+dnsconstants.DNSHostsFile, "path to the additional dnsmasq hosts")
	flag.Parse()
}

func main() {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		panic(err)
	}

	stop := make(chan struct{})

	lcb := controller.LatticeClientBuilder{
		Kubeconfig: config,
	}

	versionedLatticeClient := lcb.ClientOrDie("shared-latticeinformers")
	latticeInformers := latticeinformers.NewSharedInformerFactory(versionedLatticeClient, time.Duration(12*time.Hour))

	if err != nil {
		panic(err)
	}

	glog.V(1).Info("Starting dns controller")

	go dnscontroller.NewController(
		dnsmasqConfigPath,
		hostsFilePath,
		lcb.ClientOrDie("local-dns-lattice-address"),
		latticeInformers.Lattice().V1().Endpoints(),
	).Run(4, stop)

	glog.V(1).Info("Starting informer factory")
	latticeInformers.Start(stop)

	select {}
}
