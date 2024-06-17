data "tdh_dataplanes" "all" {
}
output "resp" {
  value = data.tdh_dataplanes.all
}

