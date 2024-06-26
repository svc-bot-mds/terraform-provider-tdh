---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tdh_object_storage Resource - tdh"
subcategory: ""
description: |-
  Represents a object storage created on TDH, can be used to create/update/delete/import a object storage.
  Note: For SRE only.
---

# tdh_object_storage (Resource)

Represents a object storage created on TDH, can be used to create/update/delete/import a object storage.
**Note:** For SRE only.

## Example Usage

```terraform
resource "tdh_object_storage" "example" {
  name              = "tf-example-object-store"
  bucket_name       = "my-bucket"
  endpoint          = "https://s3.amazonaws.com"
  region            = "us-east-1"
  access_key_id     = "ACCESS_KEY"
  secret_access_key = "SECRET_KEY"

  // non editable fields during the update
  lifecycle {
    ignore_changes = [name]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `access_key_id` (String) Access Key Id for the authentication of object storage.
- `bucket_name` (String) Name of the initial bucket to create. Modifying this field is not allowed.
- `endpoint` (String) Endpoint of the object storage to use. Modifying this field is not allowed.
- `name` (String) Name of the object storage. Updating it will result in creating new object store.
- `region` (String) Region where object storage is created. Modifying this field is not allowed.
- `secret_access_key` (String, Sensitive) Secret Access Key for the authentication of object storage.

### Read-Only

- `id` (String) Auto-generated ID after creating a object storage, and can be passed to import an existing user from TDH to terraform state.

## Import

Import is supported using the following syntax:

```shell
terraform import tdh_object_storage.example c31457a4-1641-4a07-ab43-b471e0563fbd
```
