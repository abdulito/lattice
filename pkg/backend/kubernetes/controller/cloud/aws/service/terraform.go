package service

import (
	"encoding/json"
	"fmt"
	"strings"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubetf "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	tf "github.com/mlab-lattice/system/pkg/terraform"
	tfconfig "github.com/mlab-lattice/system/pkg/terraform/config"
	awstf "github.com/mlab-lattice/system/pkg/terraform/config/aws"

	"github.com/mlab-lattice/system/pkg/types"
	corev1 "k8s.io/api/core/v1"
)

const (
	terraformStatePathService = "/services"
)

func (c *Controller) provisionService(svc *crv1.Service) error {
	var svcTfConfig interface{}
	{
		// Need a consistent view of our config while generating the config
		c.configLock.RLock()
		defer c.configLock.RUnlock()

		svcTf, err := c.getServiceTerraformConfig(svc)
		if err != nil {
			return err
		}

		svcTfConfig = svcTf
	}

	tec, err := tf.NewTerrafromExecContext(getWorkingDirectory(svc), nil)
	if err != nil {
		return err
	}

	svcTfConfigBytes, err := json.Marshal(svcTfConfig)
	if err != nil {
		return err
	}

	err = tec.AddFile("config.tf", svcTfConfigBytes)
	if err != nil {
		return err
	}

	result, _, err := tec.Init()
	if err != nil {
		return err
	}

	err = result.Wait()
	if err != nil {
		return err
	}

	result, _, err = tec.Apply(nil)
	if err != nil {
		return err
	}

	return result.Wait()
}

func (c *Controller) deprovisionService(svc *crv1.Service) error {
	var svcTfConfig interface{}
	{
		// Need a consistent view of our config while generating the config
		c.configLock.RLock()
		defer c.configLock.RUnlock()

		svcTf, err := c.getServiceTerraformConfig(svc)
		if err != nil {
			return err
		}

		svcTfConfig = svcTf
	}

	tec, err := tf.NewTerrafromExecContext(getWorkingDirectory(svc), nil)
	if err != nil {
		return err
	}

	svcTfConfigBytes, err := json.Marshal(svcTfConfig)
	if err != nil {
		return err
	}

	err = tec.AddFile("config.tf", svcTfConfigBytes)
	if err != nil {
		return err
	}

	result, _, err := tec.Init()
	if err != nil {
		return err
	}

	err = result.Wait()
	if err != nil {
		return err
	}

	result, _, err = tec.Destroy(nil)
	if err != nil {
		return err
	}

	err = result.Wait()
	if err != nil {
		return err
	}

	return c.removeFinalizer(svc)
}

func (c *Controller) getServiceTerraformConfig(service *crv1.Service) (interface{}, error) {
	systemName, err := kubeutil.SystemID(service.Namespace)
	if err != nil {
		return nil, err
	}

	kubeSvc, necessary, err := c.getKubeServiceForService(service)
	if err != nil {
		return nil, err
	}

	var serviceModule interface{}
	if necessary {
		if kubeSvc == nil {
			return nil, fmt.Errorf("Service %v requires kubeSvc but it does not exist", service.Name)
		}

		serviceModule, err = c.getServiceDedicatedPublicHTTPTerraformModule(service, kubeSvc)
	} else {
		serviceModule, err = c.getServiceDedicatedPrivateTerraformModule(service)
	}
	if err != nil {
		return nil, err
	}

	awsConfig := c.config.Provider.AWS

	config := tfconfig.Config{
		Provider: awstf.Provider{
			Region: awsConfig.Region,
		},
		Backend: awstf.S3Backend{
			Region: awsConfig.Region,
			Bucket: c.config.Terraform.Backend.S3.Bucket,
			Key: fmt.Sprintf("%v%v/%v",
				kubetf.GetS3BackendStatePathRoot(c.clusterID, types.SystemID(systemName)),
				terraformStatePathService,
				service.Name),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"service": serviceModule,
		},
	}

	return config, nil
}

func (c *Controller) getServiceDedicatedPrivateTerraformModule(service *crv1.Service) (interface{}, error) {
	awsConfig := c.config.Provider.AWS

	systemID, err := kubeutil.SystemID(service.Namespace)
	if err != nil {
		return nil, err
	}

	module := kubetf.ServiceDedicatedPrivate{
		Source: c.terraformModulePath + kubetf.ModulePathServiceDedicatedPrivate,

		AWSAccountID: awsConfig.AccountID,
		Region:       awsConfig.Region,

		VPCID:                     awsConfig.VPCID,
		SubnetIDs:                 strings.Join(awsConfig.SubnetIDs, ","),
		MasterNodeSecurityGroupID: awsConfig.MasterNodeSecurityGroupID,
		BaseNodeAmiID:             awsConfig.BaseNodeAMIID,
		KeyName:                   awsConfig.KeyName,

		SystemID:  string(systemID),
		ServiceID: service.Name,
		// FIXME: support min/max instances
		NumInstances: *service.Spec.Definition.Resources.NumInstances,
		InstanceType: *service.Spec.Definition.Resources.InstanceType,
	}
	return module, nil
}

func (c *Controller) getServiceDedicatedPublicHTTPTerraformModule(service *crv1.Service, kubeSvc *corev1.Service) (interface{}, error) {
	awsConfig := c.config.Provider.AWS

	systemName, err := kubeutil.SystemID(service.Namespace)
	if err != nil {
		return nil, err
	}

	publicComponentPorts := map[int32]bool{}
	for _, component := range service.Spec.Definition.Components {
		for _, port := range component.Ports {
			if port.ExternalAccess != nil && port.ExternalAccess.Public {
				publicComponentPorts[port.Port] = true
			}
		}
	}

	ports := map[int32]int32{}
	for _, port := range kubeSvc.Spec.Ports {
		if _, ok := publicComponentPorts[port.Port]; ok {
			ports[port.Port] = port.NodePort
		}
	}

	module := kubetf.ServiceDedicatedPublicHTTP{
		Source: c.terraformModulePath + kubetf.ModulePathServiceDedicatedPublicHTTP,

		AWSAccountID: awsConfig.AccountID,
		Region:       awsConfig.Region,

		VPCID:                     awsConfig.VPCID,
		SubnetIDs:                 strings.Join(awsConfig.SubnetIDs, ","),
		MasterNodeSecurityGroupID: awsConfig.MasterNodeSecurityGroupID,
		BaseNodeAmiID:             awsConfig.BaseNodeAMIID,
		KeyName:                   awsConfig.KeyName,

		SystemID:  string(systemName),
		ServiceID: service.Name,
		// FIXME: support min/max instances
		NumInstances: *service.Spec.Definition.Resources.NumInstances,
		InstanceType: *service.Spec.Definition.Resources.InstanceType,

		Ports: ports,
	}
	return module, nil
}

func getWorkingDirectory(svc *crv1.Service) string {
	return "/tmp/lattice-controller-manager/controllers/cloud/aws/service/terraform/" + svc.Name
}
