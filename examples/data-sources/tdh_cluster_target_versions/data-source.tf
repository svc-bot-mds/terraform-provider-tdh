data "tdh_cluster_target_versions" "all" {
  id = "CLUSTER_ID" # required field
}

output "resp" {
  value = data.tdh_cluster_target_versions.all
}

