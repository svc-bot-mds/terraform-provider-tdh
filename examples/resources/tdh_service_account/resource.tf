data "tdh_roles" "all" {
}

data "tdh_policies" "all" {
  identity_type = "SERVICE_ACCOUNT"
}

output "values" {
  # view the output to decide on resource values
  value = {
    roles    = data.tdh_roles.all
    policies = data.tdh_policies.all
  }
}

resource "tdh_service_account" "example" {
  name       = "example-acc@vmware.com"
  tags       = ["new-user", "viewer"]
  policy_ids = data.tdh_policies.all.list[*].id # filter or select all policies

  // non editable fields
  lifecycle {
    ignore_changes = [name]
  }
  //Oauth app details
  oauth_app = {
    description = "Oauth app created for example-acc service account"
    ttl_spec = {
      ttl       = "1"
      time_unit = "HOURS"
    }
  }
}
