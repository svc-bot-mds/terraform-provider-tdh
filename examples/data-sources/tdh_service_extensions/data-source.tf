data "tdh_service_extensions" "example" {
  service_type = "POSTGRES" # optional field
}

output "resp" {
  value = data.tdh_service_extensions.example
}
