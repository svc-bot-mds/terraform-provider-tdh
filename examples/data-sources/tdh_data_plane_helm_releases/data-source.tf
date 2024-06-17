data "tdh_data_plane_helm_releases" "all" {
}
output "resp" {
  value = data.tdh_data_plane_helm_releases.all
}

