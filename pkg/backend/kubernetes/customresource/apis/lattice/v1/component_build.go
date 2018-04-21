package v1

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/block"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularComponentBuild = "componentbuild"
	ResourcePluralComponentBuild   = "componentbuilds"
	ResourceScopeComponentBuild    = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ComponentBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ComponentBuildSpec   `json:"spec"`
	Status            ComponentBuildStatus `json:"status"`
}

// +k8s:deepcopy-gen=false
type ComponentBuildSpec struct {
	BuildDefinitionBlock block.ComponentBuild `json:"definitionBlock"`
}

type ComponentBuildStatus struct {
	State              ComponentBuildState           `json:"state"`
	ObservedGeneration int64                         `json:"observedGeneration"`
	Artifacts          *ComponentBuildArtifacts      `json:"artifacts,omitempty"`
	LastObservedPhase  *v1.ComponentBuildPhase       `json:"lastObservedPhase,omitempty"`
	FailureInfo        *v1.ComponentBuildFailureInfo `json:"failureInfo,omitempty"`
}

type ComponentBuildState string

const (
	ComponentBuildStatePending   ComponentBuildState = "pending"
	ComponentBuildStateQueued    ComponentBuildState = "queued"
	ComponentBuildStateRunning   ComponentBuildState = "running"
	ComponentBuildStateSucceeded ComponentBuildState = "succeeded"
	ComponentBuildStateFailed    ComponentBuildState = "failed"
)

type ComponentBuildArtifacts struct {
	DockerImageFQN string `json:"dockerImageFqn"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ComponentBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ComponentBuild `json:"items"`
}
