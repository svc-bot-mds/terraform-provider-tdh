# pass valid data with respect to the instance type selected
data "tdh_regions" "dedicated_dp" {
  instance_size        = "SMALL-LITE"
  dedicated_data_plane = true
  provider_type = "tkgs" # can be fetched using datasource "tdh_provider_types"
  service_type = "POSTGRES" # can get using datasource "tdh_provider_types"
}

output "resp" {
  value = data.tdh_regions.dedicated_dp.regions
}