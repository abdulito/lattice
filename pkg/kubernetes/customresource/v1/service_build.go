package v1

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ServiceBuildResourceSingular  = "servicebuild"
	ServiceBuildResourcePlural    = "servicebuilds"
	ServiceBuildResourceShortName = "lsvcb"
	// TODO: should this be ClusterScoped?
	ServiceBuildResourceScope = apiextensionsv1beta1.NamespaceScoped
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ServiceBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ServiceBuildSpec   `json:"spec"`
	Status            ServiceBuildStatus `json:"status,omitempty"`
}

type ServiceBuildSpec struct {
	Components map[string]ServiceBuildComponentBuildInfo `json:"components"`
}

type ServiceBuildComponentBuildInfo struct {
	DefinitionBlock   block.ComponentBuild       `json:"definitionBlock"`
	DefinitionHash    *string                    `json:"definitionHash,omitempty"`
	BuildName         *string                    `json:"buildName,omitempty"`
	BuildState        *ComponentBuildState       `json:"buildState"`
	LastObservedPhase *types.ComponentBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureInfo       *ComponentBuildFailureInfo `json:"failureInfo,omitempty"`
}

type ServiceBuildStatus struct {
	State   ServiceBuildState `json:"state,omitempty"`
	Message string            `json:"message,omitempty"`
}

type ServiceBuildState string

const (
	ServiceBuildStatePending   ServiceBuildState = "Pending"
	ServiceBuildStateRunning   ServiceBuildState = "Running"
	ServiceBuildStateSucceeded ServiceBuildState = "Succeeded"
	ServiceBuildStateFailed    ServiceBuildState = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ServiceBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ServiceBuild `json:"items"`
}

// Below is taken from: https://github.com/kubernetes/apiextensions-apiserver/blob/master/examples/client-go/apis/cr/v1/zz_generated.deepcopy.go
// It's needed because runtime.Scheme.AddKnownTypes requires the type to implement runtime.interfaces.Object,
// which includes DeepCopyObject
// TODO: figure out how to autogen this

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceBuild) DeepCopyInto(out *ServiceBuild) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Example.
func (in *ServiceBuild) DeepCopy() *ServiceBuild {
	if in == nil {
		return nil
	}
	out := new(ServiceBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceBuild) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceBuildList) DeepCopyInto(out *ServiceBuildList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceBuild, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleList.
func (in *ServiceBuildList) DeepCopy() *ServiceBuildList {
	if in == nil {
		return nil
	}
	out := new(ServiceBuildList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceBuildList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceBuildSpec) DeepCopyInto(out *ServiceBuildSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleSpec.
func (in *ServiceBuildSpec) DeepCopy() *ServiceBuildSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceBuildSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceBuildStatus) DeepCopyInto(out *ServiceBuildStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleStatus.
func (in *ServiceBuildStatus) DeepCopy() *ServiceBuildStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceBuildStatus)
	in.DeepCopyInto(out)
	return out
}
