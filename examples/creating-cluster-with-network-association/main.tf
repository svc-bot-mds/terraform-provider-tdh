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
  policy_with_update = ["custom-nw"]
  instance_type      = "XX-SMALL"
}

data "tdh_regions" "all" {
  provider_type = "aws"
  instance_size = "XX-SMALL"
}

output "regions" {
  value = data.tdh_regions.all
}

data "tdh_network_policies" "create" {
  names = local.policy_with_create
}

data "tdh_network_policies" "update" {
  names = local.policy_with_update
}

output "network_policies_data" {
  value = {
    update = data.tdh_network_policies.update
    create = data.tdh_network_policies.create
  }
}

resource "tdh_cluster" "test" {
  name               = "my-rmq-cls"
  service_type       = local.service_type
  provider_type      = local.provider
  instance_size      = local.instance_type
  region             = data.tdh_regions.all.regions[0].id
  network_policy_ids = data.tdh_network_policies.create.policies[*].id
  tags               = ["tdh-tf", "example", "new-tag"]
  timeouts = {
    create = "10m"
  }
}


resource "tdh_cluster_network_policies_association" "test" {
  id         = tdh_cluster.test.id
  policy_ids = data.tdh_network_policies.update.policies[*].id
}