data "tdh_provider_types" "all" {
}

data "tdh_cluster_versions" "name" {
  service_type  = "REDIS"
  provider_type = "tkgs"
}
