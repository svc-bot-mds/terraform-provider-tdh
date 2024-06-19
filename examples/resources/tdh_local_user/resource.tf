data "tdh_policies" "local_user" {
  identity_type = "LOCAL_USER_ACCOUNT"
}

output "local_user_policies" {
  value = data.tdh_policies.local_user.list
}

resource "tdh_local_user" "example" {
  username = "some_local_user"
  password = {
    new     = "Admin!23"
    confirm = "Admin!23"
  }
  tags       = ["new-user"]
  policy_ids = data.tdh_policies.local_user.list[0].id # available policies can be known with datasource "tdh_policies"
}