terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host = "TDH_HOST_URL"
  # Authentication using username and password
  username = "TDH_USERNAME"
  password = "TDH_PASSWORD"
  org_id   = "TDH_ORG_ID"
}

locals {
  cluster_name = "test"
  queue_name   = "dc"
}

data "tdh_clusters" "cluster_list" {
  service_type = "RABBITMQ"
}

data "tdh_service_roles" "roles" {
  type = "RABBITMQ"
}

data "tdh_cluster_metadata" "rmq" {
  id = "d3c49288-7b17-4e78-a6af-257b49e34e53"
}

// queue and other RMQ resources can be referred from the output to craft a permission on resources
output "cluster_metadata" {
  value = data.tdh_cluster_metadata.rmq
}

resource "tdh_policy" "rabbitmq" {
  name         = "test-tf"
  service_type = "RABBITMQ"
  permission_specs = [
    {
      permissions = ["read"],
      role        = "read",
      resource    = "cluster:${local.cluster_name}"
    },
    {
      "permissions" = ["write"],
      "role"        = "write",
      "resource"    = "cluster:${local.cluster_name}/queue:${local.queue_name}"
    }
  ]
}
