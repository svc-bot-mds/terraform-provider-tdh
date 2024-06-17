data "tdh_smtp" "all" {
}
output "resp" {
  value = data.tdh_smtp.all
}

