data "tdh_certificates" "all" {
}

output "resp" {
  value = data.tdh_certificates.all
}