data "tdh_backup" "all" {
  service_type = "< POSTGRES / RABBITMQ / MYSQL / REDIS >"
}

output "resp" {
  value = data.tdh_backup.all
}