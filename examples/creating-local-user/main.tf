terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host = "TDH_HOST"
  # Authentication using username and password
  username = "TDH_USERNAME"
  password = "TDH_PASSWORD"
  org_id   = "TDH_ORG_ID"
}

locals {
  identity_type = "LOCAL_USER_ACCOUNT"
  policies      = ["psql-viewer"]
}

data "tdh_policies" "all" {
  identity_type = local.identity_type
}

output "policies_data" {
  value = data.tdh_policies.all
}

resource "tdh_local_user" "temp" {
  username   = "tf_local_user"
  policy_ids = [for policy in data.tdh_policies.all.list : policy.id if contains(local.policies, policy.name)]
  password = {
    new = "myP4$$word"
    new = "myP4$$word"
  }
}
