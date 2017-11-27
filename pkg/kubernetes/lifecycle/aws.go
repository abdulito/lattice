package lifecycle

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mlab-lattice/system/pkg/terraform"

	awsterraform "github.com/mlab-lattice/system/pkg/kubernetes/terraform/aws"
)

type AWSProvisioner struct {
	latticeImageDockerRepository string
	tec                          *terraform.ExecContext
	config                       AWSProvisionerConfig
}

type AWSProvisionerConfig struct {
	TerraformModulePath string

	AccountId         string
	Region            string
	AvailabilityZones []string
	KeyName           string

	MasterNodeInstanceType string
	MasterNodeAMIId        string
	BaseNodeAMIId          string
}

const (
	systemModulePath = "aws/system"
	// FIXME: move to constants
	systemEnvironmentManagerAPIPort    = 80
	systemManagerAddressOutputVariable = "system_environment_manager_address"
)

func NewAWSProvisioner(latticeImageDockerRepository, workingDir string, config AWSProvisionerConfig) (*AWSProvisioner, error) {
	tec, err := terraform.NewTerrafromExecContext(workingDir, nil)
	if err != nil {
		return nil, err
	}

	ap := &AWSProvisioner{
		latticeImageDockerRepository: latticeImageDockerRepository,
		tec:    tec,
		config: config,
	}
	return ap, nil
}

func (ap *AWSProvisioner) Provision(name, url string) error {
	// Add system json to working dir
	sysJSON, err := ap.getSystemTerraformJSON(name, url)
	if err != nil {
		return err
	}

	err = ap.tec.AddFile("config.tf.json", sysJSON)
	if err != nil {
		return err
	}

	// Run `terraform init`
	result, logFilename, err := ap.tec.Init()
	if err != nil {
		return err
	}

	fmt.Printf("Running terraform init (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*ap.tec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return err
	}

	// Run `terraform apply`
	result, logFilename, err = ap.tec.Apply(nil)
	if err != nil {
		return err
	}

	fmt.Printf("Running terraform apply (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*ap.tec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return err
	}

	address, err := ap.Address(name)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for System Environment Manager to be ready...")
	return pollForSystemEnvironmentReadiness(address)
}

func (ap *AWSProvisioner) getSystemTerraformJSON(name, url string) ([]byte, error) {
	sysModule := awsterraform.System{
		Source: filepath.Join(ap.config.TerraformModulePath, systemModulePath),

		AWSAccountId:      ap.config.AccountId,
		Region:            ap.config.Region,
		AvailabilityZones: ap.config.AvailabilityZones,
		KeyName:           ap.config.KeyName,

		SystemId:            name,
		SystemDefinitionUrl: url,

		MasterNodeInstanceType:          ap.config.MasterNodeInstanceType,
		MasterNodeAMIId:                 ap.config.MasterNodeAMIId,
		SystemEnvironmentManagerAPIPort: systemEnvironmentManagerAPIPort,

		BaseNodeAmiId: ap.config.BaseNodeAMIId,
	}

	jsonMap := map[string]interface{}{
		"provider": map[string]interface{}{
			"aws": map[string]interface{}{
				"region": sysModule.Region,
			},
		},
		"module": map[string]interface{}{
			"system": sysModule,
		},
		"output": map[string]interface{}{
			systemManagerAddressOutputVariable: map[string]interface{}{
				"value": fmt.Sprintf("${module.system.%v}", systemManagerAddressOutputVariable),
			},
		},
	}

	return json.MarshalIndent(jsonMap, "", "  ")
}

func (ap *AWSProvisioner) Address(name string) (string, error) {
	return ap.tec.Output(systemManagerAddressOutputVariable)
}

func (ap *AWSProvisioner) Deprovision(name string) error {
	address, err := ap.Address(name)
	if err != nil {
		return err
	}

	err = tearDownAndWaitForSuccess(address)
	if err != nil {
		return err
	}

	// Run `terraform destroy`
	result, logFilename, err := ap.tec.Destroy(nil)
	if err != nil {
		return err
	}

	fmt.Printf("Running terraform destroy (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*ap.tec.LogPath, logFilename))
	return result.Wait()
}
