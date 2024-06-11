data "tdh_object_storages" "all" {
}

output "resp" {
  value = data.tdh_object_storages.all
}

