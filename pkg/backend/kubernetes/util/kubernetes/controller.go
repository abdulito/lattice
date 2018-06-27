package kubernetes

const (
	controllerLabel = "controller.lattice.mlab.com"
	finalizerSuffix = "/finalizer"

	AddressControllerFinalizer      = "address." + controllerLabel + finalizerSuffix
	BuildControllerFinalizer        = "build." + controllerLabel + finalizerSuffix
	JobControllerFinalizer          = "job." + controllerLabel + finalizerSuffix
	NodePoolControllerFinalizer     = "nodepool." + controllerLabel + finalizerSuffix
	ServiceBuildControllerFinalizer = "servicebuild." + controllerLabel + finalizerSuffix
	ServiceControllerFinalizer      = "service." + controllerLabel + finalizerSuffix
	SystemControllerFinalizer       = "system." + controllerLabel + finalizerSuffix

	UserResourcePrefix = "lattice-user-"
)
