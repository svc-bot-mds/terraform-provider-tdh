data "tdh_dns" "all" {
}
output "resp" {
  value = data.tdh_dns.all
}
