package v1

import (
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularSystem  = "system"
	ResourcePluralSystem    = "systems"
	ResourceShortNameSystem = "lsys"
	ResourceScopeSystem     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type System struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemSpec   `json:"spec"`
	Status            SystemStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=false
type SystemSpec struct {
	Services map[tree.NodePath]SystemServicesInfo `json:"services"`
}

// +k8s:deepcopy-gen=false
type SystemServicesInfo struct {
	Definition definition.Service `json:"definition"`

	// ComponentBuildArtifacts maps Component names to the artifacts created by their build
	ComponentBuildArtifacts map[string]ComponentBuildArtifacts `json:"componentBuildArtifacts"`

	// Name is the name of the Service CustomResource that is created by the lattice-system-controller
	Name *string `json:"name,omitempty"`
	// Status is the last observed status of the Service CustomResource
	Status *ServiceStatus `json:"status"`
}

type SystemStatus struct {
	State SystemState `json:"state,omitempty"`
}

type SystemState string

const (
	SystemStateScaling  SystemState = "scaling"
	SystemStateUpdating SystemState = "updating"
	SystemStateStable   SystemState = "stable"
	SystemStateFailed   SystemState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []System `json:"items"`
}
