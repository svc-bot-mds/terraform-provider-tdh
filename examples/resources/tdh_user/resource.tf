terraform {
  required_providers {
    tdh = {
      source = "hashicorp.com/svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host = "https://tdh-cp-dev.tdh.ak.com"

  // Authentication using username and password
  username = "sre@broadcom.com"
  password = "Admin!23"
  org_id = "9ba341de-f018-42fa-8085-257032bfaf0e"
}

data "tdh_roles" "all" {
}

data "tdh_policies" "all" {
  identity_type = "USER_ACCOUNT"
}

output "values" {
  # view the output to decide on resource values
  value = {
    roles    = data.tdh_roles.all
    policies = data.tdh_policies.all
  }
}
data "tdh_organizations" "all" {}

output "organizations" {
  value = data.tdh_organizations.all
}

resource "tdh_user" "sample" {
  email      = "example-user@broadcom.com"
  tags       = ["new-user", "viewer"]
  role_ids   = data.tdh_roles.all.list[*].role_id
  policy_ids = data.tdh_policies.all.list[*].id # filter or select all policies

  organizations = [for s in data.tdh_organizations.all.list : s.id if !s.sre_org] # filter the organization data source to get the non sre organizations
  // non editable fields
  lifecycle {
    ignore_changes = [email]
  }
}
