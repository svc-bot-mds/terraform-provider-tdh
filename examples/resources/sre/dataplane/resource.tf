
resource "tdh_dataplane" "example" {
  account_id="account_id"
  name = "name"
  k8s_cluster_name ="k8s_cluster_name"
  provider_name = "provider_name"
  storage_classes= ["storage_classesy", "storage_classes"]
  backup_storage_class= "backup_storage_classg"
  data_plane_release_id = "data_plane_release_id"
  shared = true
  org_id = "null"
  tags = ["tags"]
  auto_upgrade = false
  services = []
  cp_bootstrapped_cluster = true
  configure_core_dns = true


  // non editable fields , edit is not allowed
  lifecycle {
    ignore_changes = [k8s_cluster_name, account_id, provider_name,storage_classes,backup_storage_class,data_plane_release_id,shared,org_id,configure_core_dns,services,cp_bootstrapped_cluster]
  }
}
