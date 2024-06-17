data "tdh_provider_types" "all" {
}

data "tdh_cloud_accounts" "all" {
}

data "tdh_k8s_clusters" "all" {
}

data "tdh_data_plane_helm_releases" "all" {

}

output "data" {
  value = {
    provider_type = data.tdh_cloud_accounts.all
    cloud_accounts       = data.tdh_cloud_accounts.all
    k8s_clusters = data.tdh_k8s_clusters.all
    helm_release = data.tdh_data_plane_helm_releases
  }
}


resource "tdh_data_plane" "example" {
  name                    = "name"
  account_id              = data.tdh_cloud_accounts.all[0].id       # this ID can be fetched from the datasource "tdh_cloud_accounts" . Provider type can be verifies using the 'provider_type' field
  k8s_cluster_name        = "k8s_cluster_name" # use datasource "tdh_k8s_clusters" to get the list of available K8s clusters available from an account
  storage_classes         = ["tdh-k8s-storage-policy", "default"]
  backup_storage_class    = "backup_storage_class"  # name of the storage class to use for backups
  data_plane_release_id   = "data_plane_release_id" # use datasource "tdh_data_plane_helm_releases" to select one of the IDs
  shared                  = true
  org_id                  = null # setting this to particular Org ID will make it available to only that Org
  tags                    = ["dev-dp"]
  auto_upgrade            = false
  services                = [] # can be fetched from the response of "tdh_data_plane_helm_releases" services field
  cp_bootstrapped_cluster = false
  configure_core_dns      = true

  // non editable fields, edit is not allowed
  lifecycle {
    ignore_changes = [
      k8s_cluster_name, account_id, provider_name, storage_classes, backup_storage_class, data_plane_release_id, shared,
      org_id, configure_core_dns, services, cp_bootstrapped_cluster
    ]
  }
}
