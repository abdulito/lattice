package v1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "lattice.mlab.com"

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme

	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

	Resources = []struct {
		Singular   string
		Plural     string
		ShortNames []string
		Scope      apiextensionsv1beta1.ResourceScope
		Kind       string
		ListKind   string
		Type       runtime.Object
		ListType   runtime.Object
	}{
		{
			Singular:   ResourceSingularComponentBuild,
			Plural:     ResourcePluralComponentBuild,
			ShortNames: []string{ResourceShortNameComponentBuild},
			Scope:      ResourceScopeComponentBuild,
			Kind:       "ComponentBuild",
			ListKind:   "ComponentBuildList",
			Type:       &ComponentBuild{},
			ListType:   &ComponentBuildList{},
		},
		{
			Singular:   ResourceSingularConfig,
			Plural:     ResourcePluralConfig,
			ShortNames: []string{},
			Scope:      ResourceScopeConfig,
			Kind:       "Config",
			ListKind:   "ConfigList",
			Type:       &Config{},
			ListType:   &ConfigList{},
		},
		{
			Singular:   ResourceSingularService,
			Plural:     ResourcePluralService,
			ShortNames: []string{ResourceShortNameService},
			Scope:      ResourceScopeService,
			Kind:       "Service",
			ListKind:   "ServiceList",
			Type:       &Service{},
			ListType:   &ServiceList{},
		},
		{
			Singular:   ResourceSingularServiceBuild,
			Plural:     ResourcePluralServiceBuild,
			ShortNames: []string{ResourceShortNameServiceBuild},
			Scope:      ResourceScopeServiceBuild,
			Kind:       "ServiceBuild",
			ListKind:   "ServiceBuildList",
			Type:       &ServiceBuild{},
			ListType:   &ServiceBuildList{},
		},
		{
			Singular:   ResourceSingularSystem,
			Plural:     ResourcePluralSystem,
			ShortNames: []string{ResourceShortNameSystem},
			Scope:      ResourceScopeSystem,
			Kind:       "System",
			ListKind:   "SystemList",
			Type:       &System{},
			ListType:   &SystemList{},
		},
		{
			Singular:   ResourceSingularSystemBuild,
			Plural:     ResourcePluralSystemBuild,
			ShortNames: []string{ResourceShortNameSystemBuild},
			Scope:      ResourceScopeSystemBuild,
			Kind:       "SystemBuild",
			ListKind:   "SystemBuildList",
			Type:       &SystemBuild{},
			ListType:   &SystemBuildList{},
		},
		{
			Singular:   ResourceSingularSystemRollout,
			Plural:     ResourcePluralSystemRollout,
			ShortNames: []string{ResourceShortNameSystemRollout},
			Scope:      ResourceScopeSystemRollout,
			Kind:       "SystemRollout",
			ListKind:   "SystemRolloutList",
			Type:       &SystemRollout{},
			ListType:   &SystemRolloutList{},
		},
		{
			Singular:   ResourceSingularSystemTeardown,
			Plural:     ResourcePluralSystemTeardown,
			ShortNames: []string{ResourceShortNameSystemTeardown},
			Scope:      ResourceScopeSystemTeardown,
			Kind:       "SystemTeardown",
			ListKind:   "SystemTeardownList",
			Type:       &SystemTeardown{},
			ListType:   &SystemTeardownList{},
		},
	}
)

// Resource takes an unqualified resource and returns a Group-qualified GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	for _, resource := range Resources {
		scheme.AddKnownTypes(
			SchemeGroupVersion,
			resource.Type.(runtime.Object),
			resource.ListType.(runtime.Object),
		)
	}
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
