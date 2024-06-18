data "tdh_eligible_data_planes" "all" {
  provider_name = "tkgs"
  org_id        = "ORG_ID" # leave out t filter shared data planes
}
output "resp" {
  value = data.tdh_eligible_data_planes.all
}

