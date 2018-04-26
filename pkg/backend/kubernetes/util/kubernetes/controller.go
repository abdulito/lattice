package kubernetes

const (
	controllerLabel = "controller.lattice.mlab.com"
	finalizerSuffix = "/finalizer"

	AddressControllerFinalizer      = "address." + controllerLabel + finalizerSuffix
	BuildControllerFinalizer        = "build." + controllerLabel + finalizerSuffix
	NodePoolControllerFinalizer     = "nodepool." + controllerLabel + finalizerSuffix
	ServiceBuildControllerFinalizer = "servicebuild." + controllerLabel + finalizerSuffix
)
