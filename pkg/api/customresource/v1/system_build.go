package v1

import (
	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	SystemBuildResourceSingular  = "systembuild"
	SystemBuildResourcePlural    = "systembuilds"
	SystemBuildResourceShortName = "lsysb"
	// TODO: should this be ClusterScoped?
	SystemBuildResourceScope = apiextensionsv1beta1.NamespaceScoped

	SystemBuildVersionLabelKey = "build.system.lattice.mlab.com/version"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SystemBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemBuildSpec   `json:"spec"`
	Status            SystemBuildStatus `json:"status,omitempty"`
}

type SystemBuildSpec struct {
	Services []SystemBuildServicesInfo `json:"services"`
}

type SystemBuildServicesInfo struct {
	Path              string                   `json:"path"`
	Definition        systemdefinition.Service `json:"definition"`
	ServiceBuildName  *string                  `json:"serviceBuildName,omitempty"`
	ServiceBuildState *ServiceBuildState       `json:"serviceBuildState"`
}

type SystemBuildStatus struct {
	State   SystemBuildState `json:"state,omitempty"`
	Message string           `json:"message,omitempty"`
}

type SystemBuildState string

const (
	SystemBuildStatePending   SystemBuildState = "Pending"
	SystemBuildStateRunning   SystemBuildState = "Running"
	SystemBuildStateSucceeded SystemBuildState = "Succeeded"
	SystemBuildStateFailed    SystemBuildState = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SystemBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemBuild `json:"items"`
}

// Below is taken from: https://github.com/kubernetes/apiextensions-apiserver/blob/master/examples/client-go/apis/cr/v1/zz_generated.deepcopy.go
// It's needed because runtime.Scheme.AddKnownTypes requires the type to implement runtime.interfaces.Object,
// which includes DeepCopyObject
// TODO: figure out how to autogen this

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemBuild) DeepCopyInto(out *SystemBuild) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Example.
func (in *SystemBuild) DeepCopy() *SystemBuild {
	if in == nil {
		return nil
	}
	out := new(SystemBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemBuild) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemBuildList) DeepCopyInto(out *SystemBuildList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SystemBuild, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleList.
func (in *SystemBuildList) DeepCopy() *SystemBuildList {
	if in == nil {
		return nil
	}
	out := new(SystemBuildList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemBuildList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemBuildSpec) DeepCopyInto(out *SystemBuildSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleSpec.
func (in *SystemBuildSpec) DeepCopy() *SystemBuildSpec {
	if in == nil {
		return nil
	}
	out := new(SystemBuildSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemBuildStatus) DeepCopyInto(out *SystemBuildStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExampleStatus.
func (in *SystemBuildStatus) DeepCopy() *SystemBuildStatus {
	if in == nil {
		return nil
	}
	out := new(SystemBuildStatus)
	in.DeepCopyInto(out)
	return out
}
