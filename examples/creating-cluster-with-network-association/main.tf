terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host     = "TDH_HOST_URL"
  type     = "user_creds" # Authentication using username and password
  username = "TDH_USERNAME"
  password = "TDH_PASSWORD"
  org_id   = "TDH_ORG_ID"
}

locals {
  service_type       = "RABBITMQ"
  provider           = "tkgs"
  policy_with_create = ["open-to-all"]
  policy_with_update = ["custom-nw"]
  instance_type      = "XX-SMALL" # complete list can be got using datasource "tdh_instance_types"
}

data "tdh_regions" "all" {
  provider_type = "tkgs"
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
  data_plane_id      = "DP_ID" # can get using datasource "tdh_regions" based on instance size selected there
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