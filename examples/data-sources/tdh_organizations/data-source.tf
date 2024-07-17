data "tdh_organizations" "all" {
}

output "resp" {
  value = data.tdh_organizations.all
}