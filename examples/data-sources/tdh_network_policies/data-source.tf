data "tdh_network_policies" "all" {
  service_type = "POSTGRES" # optional to pass service_type for precise filtering
}