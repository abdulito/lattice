package v1

import (
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularSystemRollout  = "systemrollout"
	ResourcePluralSystemRollout    = "systemrollouts"
	ResourceShortNameSystemRollout = "lsysr"
	ResourceScopeSystemRollout     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemRollout struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemRolloutSpec   `json:"spec"`
	Status            SystemRolloutStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=false
type SystemRolloutSpec struct {
	BuildName        string `json:"buildName"`
	LatticeNamespace types.LatticeNamespace
}

type SystemRolloutStatus struct {
	State   SystemRolloutState `json:"state,omitempty"`
	Message string             `json:"message,omitempty"`
}

type SystemRolloutState string

const (
	SystemRolloutStatePending    SystemRolloutState = "Pending"
	SystemRolloutStateAccepted   SystemRolloutState = "Accepted"
	SystemRolloutStateInProgress SystemRolloutState = "InProgress"
	SystemRolloutStateSucceeded  SystemRolloutState = "Succeeded"
	SystemRolloutStateFailed     SystemRolloutState = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemRolloutList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemRollout `json:"items"`
}
