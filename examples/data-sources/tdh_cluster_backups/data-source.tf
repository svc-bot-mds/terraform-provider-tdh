data "tdh_cluster_backups" "all" {
  service_type = "< POSTGRES / MYSQL / REDIS >"
}

output "resp" {
  value = data.tdh_cluster_backups.all
}