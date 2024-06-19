terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host     = "TDH_HOST"
  type     = "user_creds" # Authentication using username and password
  username = "TDH_USERNAME"
  password = "TDH_PASSWORD"
  org_id   = "TDH_ORG_ID"
}

locals {
  account_type  = "USER_ACCOUNT"
  service_roles = ["Developer", "Admin"]
  policies      = ["viewer-policy", "eu301"]
}

data "tdh_roles" "all" {
}

output "roles_data" {
  value = data.tdh_roles.all
}

data "tdh_policies" "all" {
}

output "policies_data" {
  value = data.tdh_policies.all
}

resource "tdh_user" "temp" {
  email      = "developer11@vmware.com"
  tags       = ["new-user-tf", "update-tf-user"]
  role_ids   = [for role in data.tdh_roles.all.roles : role.role_id if contains(local.service_roles, role.name)]
  policy_ids = [for policy in data.tdh_policies.all.list : policy.id if contains(local.policies, policy.name)]

  // non editable fields
  lifecycle {
    ignore_changes = [email, status]
  }
}
