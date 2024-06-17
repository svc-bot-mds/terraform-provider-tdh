data "tdh_network_policies" "pg" {
  service_type = "POSTGRES"
}

output "data" {
  value = {
    policies = data.tdh_network_policies.pg
  }
}

# It is a good idea to first import the existing associations that may have been created during cluster creation.
resource "tdh_cluster_network_policies_association" "pg" {
  id = "df4b263e-86e6-40c2-8705-350906ddafda"
  policy_ids = [
    "EXISTING_POLICY_ID",
    "ANOTHER_POLICY_ID",
  ]
}
