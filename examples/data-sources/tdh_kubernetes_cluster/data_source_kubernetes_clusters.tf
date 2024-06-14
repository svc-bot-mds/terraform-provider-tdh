
terraform {
  required_providers {
    tdh = {
      source = "hashicorp.com/svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host      = "https://tdh-cp-dev.tdh.nn.com/"

  username = "sre@broadcom.com"
  password = "Admin!23"
  org_id = "32fb14b6-13c8-4821-a4b3-e7b693327061"

  type = "user_creds"
}

data "tdh_kubernetes_clusters" "all"{
  account_id = "58570f12-14b8-4f39-b226-dcd0ac4c4560"
}
output "resp" {
  value = data.tdh_kubernetes_clusters.all
}

