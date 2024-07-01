data "tdh_cloud_accounts" "all" {
}

data "tdh_k8s_clusters" "all" {
  account_id = "64bdec33-9224-4ce9-9d92-dc880cef7a87"
}

data "tdh_storage_policies" "all" {
  cloud_account_id = data.tdh_cloud_accounts.all.list[0].id # can be fetched from the datasource 'tdh_cloud_accounts'
  k8s_cluster_name = data.tdh_k8s_clusters.all.list[0].name # can be fetched from the datasource 'tdh_k8s_clusters'
}

output "resp" {
  value = data.tdh_storage_policies.all
}
