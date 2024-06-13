terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host      = "TDH_HOST_URL"
  api_token = "API_TOKEN"
}

locals {
  service_type       = "RABBITMQ"
  provider           = "aws"
  policy_with_create = ["open-to-all"]
  instance_type      = "XX-SMALL"
}

data "tdh_regions" "aws_small" {
  provider_type = "aws"
  instance_size = "XX-SMALL"
}

output "regions" {
  value = data.tdh_regions.aws_small
}

data "tdh_network_policies" "create" {
  names = local.policy_with_create
}

output "network_policies_data" {
  value = {
    create = data.tdh_network_policies.create
  }
}

resource "tdh_cluster" "test" {
  name               = "my-rmq-cls"
  service_type       = local.service_type
  provider_type      = local.provider
  instance_size      = local.instance_type
  region             = data.tdh_regions.aws_small.regions[0].id
  network_policy_ids = data.tdh_network_policies.create.policies[*].id
  tags               = ["tdh-tf", "example", "new-tag"]
  timeouts = {
    create = "1m"
    delete = "1m"
  }
  // non editable fields
  lifecycle {
    ignore_changes = [instance_size, name, provider_type, region, service_type]
  }
}
