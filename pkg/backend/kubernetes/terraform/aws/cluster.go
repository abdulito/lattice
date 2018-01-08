package aws

type Cluster struct {
	Source string `json:"source"`

	AWSAccountID string `json:"aws_account_id"`
	Region       string `json:"region"`

	AvailabilityZones []string `json:"availability_zones"`

	ClusterID           string `json:"cluster_id"`
	SystemDefinitionURL string `json:"system_definition_url"`

	BaseNodeAMIID          string `json:"base_node_ami_id"`
	MasterNodeAMIID        string `json:"master_node_ami_id"`
	MasterNodeInstanceType string `json:"master_node_instance_type"`
	KeyName                string `json:"key_name"`

	ClusterManagerAPIPort int32 `json:"cluster_manager_api_port"`
}