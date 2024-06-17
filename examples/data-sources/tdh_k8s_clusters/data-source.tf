data "tdh_k8s_clusters" "all" {
  account_id = "CLOUD_ACCOUNT_ID"
}
output "resp" {
  value = data.tdh_k8s_clusters.all
}
