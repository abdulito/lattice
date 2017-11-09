###############################################################################
# Variables

variable "aws_account_id" {}
variable "region" {}

variable "availability_zones" {
  type = "list"
}

variable "system_id" {}
variable "base_node_ami_id" {}
variable "master_node_ami_id" {}
variable "master_node_instance_type" {}
variable "key_name" {}

variable "kube_apiserver_port" {
  default = 6443
}
variable "system_environment_manager_api_port" {
  default = 80
}

###############################################################################
# Output

output "subnet_ids" {
  value = "${join(",", aws_subnet.subnet.*.id)}"
}

output "vpc_id" {
  value = "${aws_vpc.vpc.id}"
}

output "route53_private_zone_id" {
  value = "${aws_route53_zone.private_zone.id}"
}

output "system_environment_manager_address" {
  value = "${aws_alb.master.dns_name}"
}

###############################################################################
# Provider

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# S3
#

# FIXME: probably want to seperate out the system's bucket so it can be
#        cold-storaged after deleting the rest of the resources
resource "aws_s3_bucket" "system_bucket" {
  bucket = "lattice.system.${var.system_id}"
  acl    = "private"

  versioning {
    enabled = true
  }

  # FIXME: this will destroy all objects. only here so we can iterate quickly in dev
  force_destroy = true
}

###############################################################################
# Networking
#

###############################################################################
# VPC

resource "aws_vpc" "vpc" {
  cidr_block           = "10.240.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags {
    KubernetesCluster = "lattice.system-.${var.system_id}"
    Name              = "lattice.system.${var.system_id}"
  }
}

###############################################################################
# Internet Gateway

resource "aws_internet_gateway" "igw" {
  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    KubernetesCluster = "lattice.system.${var.system_id}"
    Name              = "lattice.system.${var.system_id}"
  }
}

###############################################################################
# Routing

# rotue table
resource "aws_route_table" "route_table" {
  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    KubernetesCluster = "lattice.system.${var.system_id}"
    Name              = "lattice.system.${var.system_id}"
  }
}

# route for igw
resource "aws_route" "igw" {
  route_table_id         = "${aws_route_table.route_table.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.igw.id}"
}

###############################################################################
# Subnets

resource "aws_subnet" "subnet" {
  count = 4

  vpc_id            = "${aws_vpc.vpc.id}"
  availability_zone = "${element(var.availability_zones, count.index % length(var.availability_zones))}"
  cidr_block        = "${cidrsubnet(aws_vpc.vpc.cidr_block, 2, count.index)}"

  tags {
    KubernetesCluster = "lattice.system.${var.system_id}"
    Name              = "lattice.system.${var.system_id}"
  }
}

###############################################################################
# Route Table-Subnet association

resource "aws_route_table_association" "route_table_association" {
  count = 4

  subnet_id      = "${element(aws_subnet.subnet.*.id, count.index)}"
  route_table_id = "${aws_route_table.route_table.id}"
}

###############################################################################
# Route53
#

# private zone
resource "aws_route53_zone" "private_zone" {
  name          = "system.internal"
  vpc_id        = "${aws_vpc.vpc.id}"
  force_destroy = true

  tags {
    KubernetesCluster = "lattice.system.${var.system_id}"
    Name              = "lattice.system.${var.system_id}"
  }
}

###############################################################################
# Masters
#

# Want to use count in modules once supported to add HA
# https://github.com/hashicorp/terraform/issues/953
# master node
module "master_node" {
  source = "../node/master"

  aws_account_id = "${var.aws_account_id}"
  region         = "${var.region}"

  system_id               = "${var.system_id}"
  system_s3_bucket        = "${aws_s3_bucket.system_bucket.id}"
  vpc_id                  = "${aws_vpc.vpc.id}"
  vpc_cidr_block          = "${aws_vpc.vpc.cidr_block}"
  subnet_id               = "${element(aws_subnet.subnet.*.id, 0)}"
  build_subnet_ids        = "${join(",", aws_subnet.subnet.*.id)}"
  base_node_ami_id        = "${var.base_node_ami_id}"
  route53_private_zone_id = "${aws_route53_zone.private_zone.id}"

  name          = "0"
  instance_type = "${var.master_node_instance_type}"
  ami_id        = "${var.master_node_ami_id}"
  key_name      = "${var.key_name}"
}

###############################################################################
# Security group

resource "aws_security_group" "master_alb" {
  name = "lattice.system.${var.system_id}.master-alb"

  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    KubernetesCluster = "lattice.system.${var.system_id}"
    Name              = "lattice.system.${var.system_id}.master-alb"
  }
}

resource "aws_security_group_rule" "master_node_allow_internal_ingress_to_kube_apiserver_port" {
  security_group_id = "${module.master_node.security_group_id}"

  type        = "ingress"
  from_port   = "${var.kube_apiserver_port}"
  to_port     = "${var.kube_apiserver_port}"
  protocol    = "tcp"
  cidr_blocks = ["${aws_vpc.vpc.cidr_block}"]
}

resource "aws_security_group_rule" "master_node_allow_ingress_from_alb_to_system_environment_manager_port" {
  security_group_id = "${module.master_node.security_group_id}"

  type                     = "ingress"
  from_port                = "${var.system_environment_manager_api_port}"
  to_port                  = "${var.system_environment_manager_api_port}"
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.master_alb.id}"
}

resource "aws_security_group_rule" "alb_allow_egress_to_master_node_system_environment_manager_port" {
  security_group_id = "${aws_security_group.master_alb.id}"

  type                     = "egress"
  from_port                = "${var.system_environment_manager_api_port}"
  to_port                  = "${var.system_environment_manager_api_port}"
  protocol                 = "tcp"
  source_security_group_id = "${module.master_node.security_group_id}"
}

###############################################################################
# ALB

resource "aws_alb" "master" {
  name            = "lattice-system-${var.system_id}-master"
  security_groups = ["${aws_security_group.master_alb.id}"]
  subnets         = ["${aws_subnet.subnet.*.id}"]
}

resource "aws_alb_target_group" "master" {
  name     = "lattice-system-${var.system_id}-master"
  port     = "${var.system_environment_manager_api_port}"
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.vpc.id}"
}

resource "aws_alb_listener" "master" {
  load_balancer_arn = "${aws_alb.master.arn}"
  port              = "${var.system_environment_manager_api_port}"

  "default_action" {
    target_group_arn = "${aws_alb_target_group.master.arn}"
    type             = "forward"
  }
}
