resource "tdh_cluster_backup" "example" {
  cluster_id = "< Cluster ID for Backup>"
  name      = " < Backup Name > "
  description       = " < Description for cluster backup >"
  service_type = " < POSTGRES / MYSQL / REDIS > "

  // non editable fields
  lifecycle {
    ignore_changes = [name ]
  }
}