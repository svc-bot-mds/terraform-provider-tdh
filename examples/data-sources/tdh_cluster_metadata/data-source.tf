data "tdh_cluster_metadata" "example" {
  id = "CLUSTER_ID"
}
output "data" {
  value = {
    metadata = data.tdh_cluster_metadata.example
  }
}