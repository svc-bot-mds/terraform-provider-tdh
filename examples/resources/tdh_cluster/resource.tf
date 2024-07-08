data "tdh_provider_types" "all" {
}
data "tdh_instance_types" "pg" {
  service_type = local.service_type
}
locals {
  service_type        = "POSTGRES"
  provider_type       = "tkgs"        # can get using datasource "tdh_provider_types"
  instance_type       = "XX-SMALL"    # can get using datasource "tdh_instance_types"
  version             = "postgres-13" # complete list can be got using datasource "tdh_service_versions"
  storage_policy_name = "tdh-k8s-cluster-policy"
  # can get using datasource "tdh_eligible_data_planes", in the field 'list'
}
data "tdh_regions" "shared" {
  instance_size = local.instance_type
  provider_type = local.provider_type
  service_type  = local.service_type
}
data "tdh_object_storages" "all" {
}
data "tdh_network_ports" "all" {
}
# to get the storage policies and eligible data planes for the given provider, although it may not be available if given size doesn't meet resource requirement in this data plane
data "tdh_eligible_data_planes" "all" {
  provider_name = local.provider_type
  org_id        = "ORG_ID" # leave out to filter shared data planes
}

data "tdh_service_versions" "name" {
  service_type  = local.service_type
  provider_type = local.provider_type
}
output "data" {
  value = {
    providers       = data.tdh_provider_types.all
    instance_types  = data.tdh_instance_types.pg
    regions         = data.tdh_regions.shared
    object_storages = data.tdh_object_storages.all
    network_ports   = data.tdh_network_ports.all
    data_planes     = data.tdh_eligible_data_planes.all
  }
}

resource "tdh_network_policy" "network" {
  name = "tf-pg-nw-policy"
  network_spec = {
    cidr = "0.0.0.0/32",
    network_port_ids = [
      for port in data.tdh_network_ports.all.list : port.id if strcontains(port.id, "postgres")
    ]
  }
}

resource "tdh_cluster" "test" {
  name                = "tf-pg-cls"
  service_type        = local.service_type
  provider_type       = local.provider_type
  instance_size       = "XX-SMALL"    # complete list can be got using datasource "tdh_instance_types"
  region              = "REGION_NAME" # can get using datasource "tdh_regions"
  data_plane_id       = "DP_ID"       # can get using datasource "tdh_regions" based on instance size selected there
  network_policy_ids  = [tdh_network_policy.network.id]
  tags                = ["tdh-tf", "new-tag"]
  version             = local.version             # available values can be fetched using datasource "tdh_service_versions"
  storage_policy_name = local.storage_policy_name # complete list can be got using datasource "tdh_eligible_data_planes"
  cluster_metadata = {
    username          = "test"
    password          = "Admin!23"
    database          = "test"
    object_storage_id = "OBJECT_STORE_ID" # can be used from datasource "tdh_object_storages"
  }
  // non editable fields
  lifecycle {
    ignore_changes = [instance_size, name, provider_type, region, service_type]
  }
}