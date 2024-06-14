
data "tdh_dataplane_helm_release" "all"{
}
output "resp" {
  value = data.tdh_dataplane_helm_release.all
}

