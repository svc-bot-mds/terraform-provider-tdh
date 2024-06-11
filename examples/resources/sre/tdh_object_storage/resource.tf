terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host = "TDH_HOST_URL"

  //Get the authentication with "username and password"
  username = "USERNAME"
  password = "PASSWORD"
  org_id   = "ORG-ID"
  type     = "user_creds"
}

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

