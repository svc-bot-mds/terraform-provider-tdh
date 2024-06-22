data "tdh_service_roles" "pg" {
  # to view what all role can be used
  type = "POSTGRES"
}

data "tdh_cluster_metadata" "pg" {
  # to view which database/schema/table to use in resource regex
  id = "CLUSTER_ID"
}

resource "tdh_policy" "sample" {
  name         = "tf-pg-policy"
  description  = "to allow login and create DB"
  service_type = "POSTGRES"
  permission_specs = [
    {
      resource   = "cluster:${data.tdh_cluster_metadata.pg.name}"
      role       = "login" # use any value from tdh_service_roles.all.list[*].name
      permission = "login" # optional, must be same as role for services other than REDIS
    },
    {
      resource = "cluster:${data.tdh_cluster_metadata.pg.name}/database:broadcom"
      # use any value from tdh_cluster_metadata.pg.databases[*].name
      role = "create"
    },
  ]
}
