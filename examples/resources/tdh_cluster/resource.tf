
// to get the Storage Policies and Eligible Dataplane for the given Provider
data "tdh_eligible_shared_dataplanes" "all"{
  provider_name= "tkgs"
}

resource "tdh_cluster" "example" {
  name               = "test-terraform"
  cloud_provider     = "aws"
  service_type       = "RABBITMQ"
  instance_size      = "XX-SMALL"
  region             = "eu-west-1"
  network_policy_ids = ["policy id"]
  tags               = ["tdh-tf", "example"]
  dedicated          = false
  shared             = false

  data_plane_id = "dataplane id"
  // non editable fields
  lifecycle {
    ignore_changes = [instance_size, name, cloud_provider, region, service_type]
  }
}