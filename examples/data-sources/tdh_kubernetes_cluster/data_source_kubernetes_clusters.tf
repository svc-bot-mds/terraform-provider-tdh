


data "tdh_kubernetes_clusters" "all"{
  account_id = "58570f12-14b8-4f39-b226-dcd0ac4c4560"
}
output "resp" {
  value = data.tdh_kubernetes_clusters.all
}

