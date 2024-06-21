data "tdh_provider_types" "create" {
}

output "resp" {
  value = data.tdh_provider_types.create
}
resource "tdh_certificate" "example" {
  name            = "tf-example-certificate-name"
  domain_name     = "tdh.domain.com"
  provider_type   = data.tdh_provider_types.create.list[2] # can be fetched using 'tdh_provider_types' datasource
  certificate     = <<EOF
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
EOF
  certificate_ca  = <<EOF
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
EOF
  certificate_key = <<EOF
-----BEGIN PRIVATE KEY-----
-----END PRIVATE KEY-----
EOF
  // non editable fields during the update
  lifecycle {
    ignore_changes = [name, domain_name, provider_type, certificate_ca, certificate_key, certificate]
  }
}

