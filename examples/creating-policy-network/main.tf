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

# network port IDs can be referred using this datasource
data "tdh_network_ports" "all" {
}

output "cluster_metadata" {
  value = data.tdh_network_ports.all
}

resource "tdh_network_policy" "network" {
  name = "network-policy-from-tf"
  network_spec = {
    cidr             = "10.22.55.0/24",
    network_port_ids = ["rmq-streams", "rmq-amqps"]
  }
}