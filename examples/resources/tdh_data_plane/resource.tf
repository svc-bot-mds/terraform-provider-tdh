data "tdh_provider_types" "all" {
}

data "tdh_cloud_accounts" "all" {
}

# For onboarding the data plane use the k8s cluster with the attribute "available" is set to true
# If onboarding data Plane on TDH Control Plane, use the k8s cluster with the attribute "cp_present" is set to true and "dp_present" is set to false
data "tdh_k8s_clusters" "all" {
  account_id = data.tdh_cloud_accounts.all.list[0].id
}

data "tdh_data_plane_helm_releases" "all" {

}

data "tdh_storage_policies" "all" {
  cloud_account_id = data.tdh_cloud_accounts.all.list[0].id
  k8s_cluster_name = data.tdh_k8s_clusters.all.list[0].name

}
output "resp" {
  value = data.tdh_storage_policies.all
}

output "data" {
  value = {
    provider_type  = data.tdh_cloud_accounts.all
    cloud_accounts = data.tdh_cloud_accounts.all
    k8s_clusters   = data.tdh_k8s_clusters.all
    helm_release   = data.tdh_data_plane_helm_releases.all
    storage_class  = data.tdh_storage_policies.all
  }
}


resource "tdh_data_plane" "example" {
  name       = "dp_name-new"
  account_id = data.tdh_cloud_accounts.all.list[0].id
  # this ID can be fetched from the datasource "tdh_cloud_accounts" . Provider type can be verifies using the 'provider_type' field
  k8s_cluster_name = data.tdh_k8s_clusters.all.list[5].name
  # use datasource "tdh_k8s_clusters" to get the list of available K8s clusters.# For onboarding the data plane use the k8s cluster with the attribute "available" is set to tru. If onboarding Data Plane on TDH Control Plane, use the k8s cluster with the attribute "cp_present" is set to true and "dp_present" is set to false
  storage_classes       = [for storage_class in data.tdh_storage_policies.all.list : storage_class.name]
  backup_storage_class  = data.tdh_storage_policies.all.list[0].name # name of the storage class to use for backups
  data_plane_release_id = data.tdh_data_plane_helm_releases.all.list[0].id
  # use datasource "tdh_data_plane_helm_releases" to select one of the IDs
  shared       = true
  org_id       = null # setting this to particular Org ID will make it available to only that Org
  tags         = ["dev-dp-terraform"]
  auto_upgrade = false
  services     = data.tdh_data_plane_helm_releases.all.list[0].services
  # can be fetched from the response of "tdh_data_plane_helm_releases" services field
  cp_bootstrapped_cluster = false
  # set to true to Onboard Data Plane on TDH Control Plane
  configure_core_dns = true

  // non editable fields, edit is not allowed
  lifecycle {
    ignore_changes = [
      k8s_cluster_name, account_id, provider_name, storage_classes, backup_storage_class, data_plane_release_id, shared,
      org_id, configure_core_dns, services, cp_bootstrapped_cluster
    ]
  }
}
