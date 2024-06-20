# Credential format for 'tkgs' provider
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

# Credential format for 'tkgm' provider
variable "tkgm_cred" {
  description = "TKGM CRED JSON"
  type        = string
  default     = <<EOF
  {
    "kubeconfigBase64": "REPLACE"
  }
EOF
}

# Credential format for 'openshift' provider
variable "openshift_cred" {
  description = "OPENSHIFT CRED JSON"
  type        = string
  default     = <<EOF
  {
    "domain" : "<<domain>>",
    "userName" : "<<user name>>",
    "password" : "<<password>>"
  }
EOF
}

# Credential format for 'tas' provider
variable "tas_cred" {
  description = "TAS CRED JSON"
  type        = string
  default     = <<EOF
  {
    "operationManagerIp":"",
    "userName":"",
    "password":"",
    "cfUserName":"",
    "cfPassword":"",
    "cfApiHost":""
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
    ignore_changes = [name, provider_type, org_id, shared]
  }
}

