package v1

import (
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularSystemTeardown  = "systemteardown"
	ResourcePluralSystemTeardown    = "systemteardowns"
	ResourceShortNameSystemTeardown = "lsyst"
	ResourceScopeSystemTeardown     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemTeardown struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemTeardownSpec   `json:"spec"`
	Status            SystemTeardownStatus `json:"status,omitempty"`
}

type SystemTeardownSpec struct {
	LatticeNamespace types.LatticeNamespace
}

type SystemTeardownStatus struct {
	State   SystemTeardownState `json:"state,omitempty"`
	Message string              `json:"message,omitempty"`
}

type SystemTeardownState string

const (
	SystemTeardownStatePending    SystemTeardownState = "Pending"
	SystemTeardownStateInProgress SystemTeardownState = "InProgress"
	SystemTeardownStateSucceeded  SystemTeardownState = "Succeeded"
	SystemTeardownStateFailed     SystemTeardownState = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemTeardownList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemTeardown `json:"items"`
}
