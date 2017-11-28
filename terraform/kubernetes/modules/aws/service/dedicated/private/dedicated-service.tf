###############################################################################
# Variables

variable "aws_account_id" {}
variable "region" {}

variable "vpc_id" {}
variable "subnet_ids" {}
variable "base_node_ami_id" {}
variable "key_name" {}

variable "system_id" {}
variable "service_id" {}
variable "num_instances" {}
variable "instance_type" {}

variable "master_node_security_group_id" {}

###############################################################################
# Provider

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# Service node

module "service_node" {
  source = "../../../node/service"

  aws_account_id = "${var.aws_account_id}"
  region         = "${var.region}"

  system_id  = "${var.system_id}"
  vpc_id     = "${var.vpc_id}"
  subnet_ids = "${var.subnet_ids}"

  service_id       = "${var.service_id}"
  num_instances    = "${var.num_instances}"
  instance_type    = "${var.instance_type}"
  base_node_ami_id = "${var.base_node_ami_id}"
  key_name         = "${var.key_name}"

  master_node_security_group_id = "${var.master_node_security_group_id}"
}
