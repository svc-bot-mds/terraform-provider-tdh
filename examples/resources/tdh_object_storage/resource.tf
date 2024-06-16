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

