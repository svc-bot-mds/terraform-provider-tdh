data "tdh_data_planes" "all" {
}
output "resp" {
  value = data.tdh_data_planes.all
}

