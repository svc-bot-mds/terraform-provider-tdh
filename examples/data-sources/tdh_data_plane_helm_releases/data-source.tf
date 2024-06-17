data "tdh_data_plane_helm_release" "all" {
}
output "resp" {
  value = data.tdh_data_plane_helm_release.all
}

