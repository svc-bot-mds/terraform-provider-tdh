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
