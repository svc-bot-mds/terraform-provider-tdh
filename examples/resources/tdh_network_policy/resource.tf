// network port IDs can be referred using this datasource
data "tdh_network_ports" "pg" {
  service_type = "POSTGRES"
}
output "network_ports" {
  value = data.tdh_network_ports.pg
}

resource "tdh_network_policy" "pg" {
  name = "tf-pg-nw-policy"
  network_spec = {
    cidr             = "0.0.0.0/32",
    network_port_ids = [for port in data.tdh_network_ports.pg.list : port.id]
  }
}
