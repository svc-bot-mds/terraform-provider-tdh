data "tdh_roles" "all" {
}

data "tdh_policies" "all" {
  identity_type = "USER_ACCOUNT"
}

output "values" { # view the output to decide on resource values
  value = {
    roles    = data.tdh_roles.all
    policies = data.tdh_policies.all
  }
}

resource "tdh_user" "sample" {
  email      = "example-user@vmware.com"
  tags       = ["new-user", "viewer"]
  role_ids   = data.tdh_roles.all.list[*].role_id
  policy_ids = data.tdh_policies.all.list[*].id # filter or select all policies

  // non editable fields
  lifecycle {
    ignore_changes = [email]
  }
}
