

data "tdh_eligible_dedicated_dataplanes" "all"{
  provider_name= "tkgs"
  org_id="<<orgId>>"
}
output "resp" {
  value = data.tdh_eligible_dedicated_dataplanes.all
}

