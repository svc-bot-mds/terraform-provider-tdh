# For onboarding the data plane use the k8s cluster where the attribute "available" is set to true
# If onboarding data Plane on TDH Control Plane, use the k8s cluster where the attribute "cp_present" is set to true and "dp_present" is set to false

data "tdh_k8s_clusters" "all" {
  account_id = "CLOUD_ACCOUNT_ID" # can be fetched using 'tdh_cloud_accounts" datasource
}
output "resp" {
  value = data.tdh_k8s_clusters.all
}
