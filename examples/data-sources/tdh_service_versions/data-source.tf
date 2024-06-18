data "tdh_provider_types" "all" {
}

data "tdh_service_versions" "pg" {
  service_type  = "POSTGRES"
  provider_type = "tkgs" # available values can be known with datasource "tdh_provider_types" above
}
