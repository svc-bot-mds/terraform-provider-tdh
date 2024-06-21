# to create cluster backup
resource "tdh_cluster_backup" "example" {
  name        = "tf-first-backup"
  description = "my postgres cluster backup using TDH provider"
  cluster_id  = "CLUSTER_ID" # ID of the cluster to backup, use datasource "tdh_clusters" to see available clusters

  // non editable fields
  lifecycle {
    ignore_changes = [name]
  }
}

# to restore a backup, either use same created resource or import a backup using backup ID. Then initialize the restore config like so:
resource "tdh_cluster_backup" "example" {
  name       = "created-backup" # if importing, name has to match imported backup name
  cluster_id = "CLUSTER_ID"     # if importing, ID has to match with existing state

  restore = {
    cluster_name = "tf-restore-6"
    # this will be the name of cluster that will be created with this restore. Not Applicable for "REDIS"
    storage_policy = "tdh-k8s-cluster-policy"
    # can get using datasource "tdh_storage_policies". Not Applicable for "REDIS"
    network_policy_ids = [
      # can get using datasource "tdh_network_policies"
      # Not Applicable for "REDIS"
      "6ad0cf49-81be-48e3-bab4-2a13b9de0c95"
    ]
  }
  // non editable fields
  lifecycle {
    ignore_changes = [name]
  }
}