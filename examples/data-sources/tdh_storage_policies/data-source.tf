data "tdh_cloud_accounts" "all" {
}

data "tdh_k8s_clusters" "all" {
}

data "tdh_storage_policies" "all" {
  cloud_account_id = data.tdh_cloud_accounts.all[0].id # can be fetched from the datasource 'tdh_cloud_accounts'
  k8s_cluster_name = "k8s cluster name"                #can be fetched from the datasource 'tdh_k8s_clusters'
}
output "resp" {
  value = data.tdh_storage_policies.all
}

