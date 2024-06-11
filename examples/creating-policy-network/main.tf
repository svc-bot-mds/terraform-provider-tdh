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

// network port IDs can be referred using this datasource
data "tdh_network_ports" "all" {
}

output "cluster_metadata" {
  value = data.tdh_network_ports.all
}

resource "tdh_network_policy" "network" {
  name         = "network-policy-from-m"
  service_type = "NETWORK"
  network_spec = {
    cidr             = "10.22.55.0/24",
    network_port_ids = ["rmq-streams", "rmq-amqps"]
  }
}