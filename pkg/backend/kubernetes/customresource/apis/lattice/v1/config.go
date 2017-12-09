package v1

import (
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularConfig = "config"
	ResourcePluralConfig   = "configs"
	ResourceScopeConfig    = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ConfigSpec `json:"spec"`
}

type ConfigSpec struct {
	SystemID       string                                  `json:"systemID"`
	Provider       ConfigProvider                          `json:"providerConfig"`
	ComponentBuild ConfigComponentBuild                    `json:"componentBuild"`
	Envoy          ConfigEnvoy                             `json:"envoy"`
	SystemConfigs  map[types.LatticeNamespace]ConfigSystem `json:"userSystem"`
	Terraform      *ConfigTerraform                        `json:"terraform,omitempty"`
}

type ConfigProvider struct {
	Local *ConfigProviderLocal `json:"local,omitempty"`
	AWS   *ConfigProviderAWS   `json:"aws,omitempty"`
}

type ConfigProviderLocal struct {
	IP string `json:"ip"`
}

type ConfigProviderAWS struct {
	Region                    string   `json:"region"`
	AccountID                 string   `json:"accountID"`
	VPCID                     string   `json:"vpcID"`
	SubnetIDs                 []string `json:"subnetIDs"`
	MasterNodeSecurityGroupID string   `json:"masterNodeSecurityGroupID"`
	BaseNodeAMIID             string   `json:"baseNodeAmiID"`
	KeyName                   string   `json:"keyName"`
}

type ConfigSystem struct {
	URL string `json:"url"`
}

type ConfigComponentBuild struct {
	DockerConfig ConfigBuildDocker `json:"dockerConfig"`
	BuildImage   string            `json:"buildImage"`
}

type ConfigBuildDocker struct {
	// Registry used to tag images.
	Registry string `json:"registry"`

	// If true, make a new repository for the image.
	// If false, use Repository as the repository for the image and give it
	// a unique tag.
	RepositoryPerImage bool   `json:"repositoryPerImage"`
	Repository         string `json:"repository"`

	// If true push the image to the repository.
	// Set to false for the local case.
	Push bool `json:"push"`

	// Version of the docker API used by the build node docker daemons
	APIVersion string `json:"apiVersion"`
}

type ConfigEnvoy struct {
	PrepareImage      string `json:"prepareImage"`
	Image             string `json:"image"`
	RedirectCidrBlock string `json:"redirectCidrBlock"`
	XDSAPIPort        int32  `json:"xdsApiPort"`
}

type ConfigTerraform struct {
	S3Backend *ConfigTerraformBackendS3 `json:"s3Backend,omitempty"`
}

type ConfigTerraformBackendS3 struct {
	Bucket string `json:"bucket"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Config `json:"items"`
}