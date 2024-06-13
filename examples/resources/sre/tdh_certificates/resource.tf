
resource "tdh_certificate" "example" {
  name              = "tf-example-object-store"
  domain_name       = "tdh.nn.com"
  provider_type          = "openshift"
  certificate            = <<EOF
  "-----BEGIN CERTIFICATE-----

-----END CERTIFICATE-----"
EOF
certificate_ca     = <<EOF
"-----BEGIN PRIVATE KEY-----

-----END PRIVATE KEY-----"
EOF
  certificate_key = <<EOF
  "-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----"
EOF

  // non editable fields during the update
  lifecycle {
    ignore_changes = [name]
  }
}

