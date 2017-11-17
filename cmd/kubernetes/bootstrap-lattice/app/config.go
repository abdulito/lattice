package app

import (
	"fmt"
	"strings"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crdclient "github.com/mlab-lattice/system/pkg/kubernetes/customresource"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

func seedConfig(kubeconfig *rest.Config, userSystemUrl string) {
	fmt.Println("Seeding lattice config...")
	crClient, _, err := crdclient.NewClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	// Create config
	config := &crv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ConfigGlobal,
			Namespace: constants.NamespaceLatticeInternal,
		},
		Spec: crv1.ConfigSpec{
			SystemConfigs: map[coretypes.LatticeNamespace]crv1.SystemConfig{
				coreconstants.UserSystemNamespace: {
					Url: userSystemUrl,
				},
			},
			Envoy: crv1.EnvoyConfig{
				PrepareImage:      latticeContainerRegistry + "/envoy-prepare-envoy",
				Image:             "envoyproxy/envoy",
				RedirectCidrBlock: "172.16.29.0/16",
				XdsApiPort:        8080,
			},
			ComponentBuild: crv1.ComponentBuildConfig{
				DockerConfig: crv1.BuildDockerConfig{
					RepositoryPerImage: false,
					Repository:         constants.DockerRegistryComponentBuildsDefault,
					Push:               true,
					Registry:           componentBuildRegistry,
				},
				BuildDockerImage: latticeContainerRegistry + "/component-build-build-docker-image",
				GetEcrCredsImage: latticeContainerRegistry + "/component-build-get-ecr-creds",
				PullGitRepoImage: latticeContainerRegistry + "/component-build-pull-git-repo",
			},
		},
	}

	switch provider {
	case coreconstants.ProviderLocal:
		config.Spec.ComponentBuild.DockerConfig.Push = false

		localConfig, err := getLocalConfig()
		if err != nil {
			panic(err)
		}

		config.Spec.Provider.Local = localConfig
	case coreconstants.ProviderAWS:
		awsConfig, err := getAwsConfig()
		if err != nil {
			panic(err)
		}
		config.Spec.Provider.AWS = awsConfig
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return nil, crClient.Post().
			Namespace(constants.NamespaceLatticeInternal).
			Resource(crv1.ConfigResourcePlural).
			Body(config).
			Do().Into(nil)
	})
}

func getLocalConfig() (*crv1.ProviderConfigLocal, error) {
	// TODO: find a better way to do the parsing of the provider variables
	expectedVars := map[string]interface{}{
		"system-ip": nil,
	}

	for _, providerVar := range *providerVars {
		split := strings.Split(providerVar, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid provider variable " + providerVar)
		}

		key := split[0]
		var value interface{} = split[1]

		existingVal, ok := expectedVars[key]
		if !ok {
			return nil, fmt.Errorf("unexpected provider variable " + key)
		}
		if existingVal != nil {
			return nil, fmt.Errorf("provider variable " + key + " set multiple times")
		}

		expectedVars[key] = value
	}

	for k, v := range expectedVars {
		if v == nil {
			return nil, fmt.Errorf("missing required provider variable " + k)
		}
	}

	localConfig := &crv1.ProviderConfigLocal{
		IP: expectedVars["system-ip"].(string),
	}

	return localConfig, nil
}

func getAwsConfig() (*crv1.ProviderConfigAWS, error) {
	// TODO: find a better way to do the parsing of the provider variables
	expectedVars := map[string]interface{}{
		"account-id":       nil,
		"region":           nil,
		"vpc-id":           nil,
		"subnet-ids":       nil,
		"base-node-ami-id": nil,
		"key-name":         nil,
	}

	for _, providerVar := range *providerVars {
		split := strings.Split(providerVar, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid provider variable " + providerVar)
		}

		key := split[0]
		var value interface{} = split[1]

		existingVal, ok := expectedVars[key]
		if !ok {
			return nil, fmt.Errorf("unexpected provider variable " + key)
		}
		if existingVal != nil {
			return nil, fmt.Errorf("provider variable " + key + " set multiple times")
		}

		if key == "subnet-ids" {
			value = strings.Split(value.(string), ",")
		}

		expectedVars[key] = value
	}

	for k, v := range expectedVars {
		if v == nil {
			return nil, fmt.Errorf("missing required provider variable " + k)
		}
	}

	awsConfig := &crv1.ProviderConfigAWS{
		Region:        expectedVars["region"].(string),
		AccountId:     expectedVars["account-id"].(string),
		VPCId:         expectedVars["vpc-id"].(string),
		SubnetIds:     expectedVars["subnet-ids"].([]string),
		BaseNodeAMIId: expectedVars["base-node-ami-id"].(string),
		KeyName:       expectedVars["key-name"].(string),
	}

	return awsConfig, nil
}
