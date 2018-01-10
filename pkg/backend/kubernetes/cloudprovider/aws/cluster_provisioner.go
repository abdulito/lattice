package aws

import (
	"fmt"
	"path/filepath"
	"time"

	awsterraform "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
	"github.com/mlab-lattice/system/pkg/terraform"
	awstfprovider "github.com/mlab-lattice/system/pkg/terraform/provider/aws"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	clusterModulePath = "aws/cluster"
	// FIXME: move to constants
	clusterManagerAPIPort                = 80
	terraformOutputclusterManagerAddress = "cluster_manager_address"
)

type ClusterProvisionerOptions struct {
	TerraformModulePath string

	AccountID         string
	Region            string
	AvailabilityZones []string
	KeyName           string

	MasterNodeInstanceType string
	MasterNodeAMIID        string
	BaseNodeAMIID          string
}

type DefaultAWSClusterProvisioner struct {
	workDirectory string

	latticeContainerRegistry   string
	latticeContainerRepoPrefix string

	terraformModulePath string

	accountID         string
	region            string
	availabilityZones []string
	keyName           string

	masterNodeInstanceType string
	masterNodeAMIID        string
	baseNodeAMIID          string
}

func NewClusterProvisioner(latticeImageDockerRepository, latticeContainerRepoPrefix, workingDir string, options *ClusterProvisionerOptions) *DefaultAWSClusterProvisioner {
	return &DefaultAWSClusterProvisioner{
		workDirectory: workingDir,

		latticeContainerRegistry:   latticeImageDockerRepository,
		latticeContainerRepoPrefix: latticeContainerRepoPrefix,

		terraformModulePath: options.TerraformModulePath,

		accountID:         options.AccountID,
		region:            options.Region,
		availabilityZones: options.AvailabilityZones,
		keyName:           options.KeyName,

		masterNodeInstanceType: options.MasterNodeInstanceType,
		masterNodeAMIID:        options.MasterNodeAMIID,
		baseNodeAMIID:          options.BaseNodeAMIID,
	}
}

func (p *DefaultAWSClusterProvisioner) Provision(clusterID, url string) error {
	clusterModule := p.clusterModule(clusterID, url)
	clusterConfig := p.clusterConfig(clusterModule)

	err := terraform.Apply(p.workDirectory, clusterConfig)
	if err != nil {
		return err
	}

	address, err := p.Address(clusterID)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for Cluster Manager to be ready...")
	clusterClient := rest.NewClient(fmt.Sprintf("http://%v", address))
	return wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		ok, _ := clusterClient.Status()
		return ok, nil
	})
}

func (p *DefaultAWSClusterProvisioner) clusterConfig(clusterModule *awsterraform.Cluster) *terraform.Config {
	config := &terraform.Config{
		Provider: awstfprovider.Provider{
			Region: p.region,
		},
	}

	if clusterModule != nil {
		config.Modules = map[string]interface{}{
			"cluster": clusterModule,
		}

		config.Output = map[string]terraform.ConfigOutput{
			terraformOutputclusterManagerAddress: {
				Value: fmt.Sprintf("${module.cluster.%v}", terraformOutputclusterManagerAddress),
			},
		}
	}

	return config
}

func (p *DefaultAWSClusterProvisioner) clusterModule(clusterID, url string) *awsterraform.Cluster {
	return &awsterraform.Cluster{
		Source: filepath.Join(p.terraformModulePath, clusterModulePath),

		AWSAccountID:      p.accountID,
		Region:            p.region,
		AvailabilityZones: p.availabilityZones,
		KeyName:           p.keyName,

		ClusterID:           clusterID,
		SystemDefinitionURL: url,

		MasterNodeInstanceType: p.masterNodeInstanceType,
		MasterNodeAMIID:        p.masterNodeAMIID,
		ClusterManagerAPIPort:  clusterManagerAPIPort,

		BaseNodeAMIID: p.baseNodeAMIID,
	}
}

func (p *DefaultAWSClusterProvisioner) Address(name string) (string, error) {
	tec, err := terraform.NewTerrafromExecContext(p.workDirectory, nil)
	if err != nil {
		return "", err
	}

	return tec.Output(terraformOutputclusterManagerAddress)
}

func (p *DefaultAWSClusterProvisioner) Deprovision(clusterID string) error {
	address, err := p.Address(clusterID)
	if err != nil {
		return err
	}

	clusterClient := rest.NewClient(fmt.Sprintf("http://%v", address))
	systems, err := clusterClient.Systems().List()
	if err != nil {
		return err
	}

	teardowns := map[types.SystemID]types.SystemTeardownID{}
	for _, system := range systems {
		teardownID, err := clusterClient.Systems().Teardowns(system.ID).Create()
		if err != nil {
			return err
		}

		teardowns[system.ID] = teardownID
	}

	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		for systemID, teardownID := range teardowns {
			teardown, err := clusterClient.Systems().Teardowns(systemID).Get(teardownID)
			if err != nil {
				return false, err
			}

			if teardown.State != types.SystemTeardownStateSucceeded {
				return false, nil
			}
		}

		return true, nil
	})

	//clusterConfig := p.clusterConfig(nil)
	return terraform.Destroy(p.workDirectory, nil)
}
