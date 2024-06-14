

data "tdh_eligible_shared_dataplanes" "all"{
  provider_name= "tkgs"
}
output "resp" {
  value = data.tdh_eligible_shared_dataplanes.all
}

