variable "tkgs_cred" {
  description = "TKGs CRED JSON"
  type        = string
  default     = <<EOF
{
    "userName": "test",
    "password": "REPLACE",
    "supervisorManagementIP": "SOME_IP",
    "vsphereNamespace": "NAMESPACE"
}
EOF
}

data "tdh_provider_types" "create" {
}

output "provider_types" {
  value = {
    create = data.tdh_provider_types.create
  }
}
resource "tdh_cloud_account" "example" {
  name          = "tf-cloud-account1"
  provider_type = element(data.tdh_provider_types.create.list, 0)
  credentials   = var.tkgs_cred
  shared        = true
  tags          = ["tag1", "tag2"]

  //non editable fields during the update
  lifecycle {
    ignore_changes = [name]
  }
}

