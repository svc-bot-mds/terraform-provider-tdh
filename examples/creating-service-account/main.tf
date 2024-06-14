terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host      = "TDH_HOST_URL"
  api_token = "API_TOKEN"
}

locals {
  policies = ["test-svc-pol"]
}

data "tdh_policies" "all" {
}

output "policies_data" {
  value = data.tdh_policies.all
}

resource "tdh_service_account" "test" {
  name       = "test-svc-tf-testing-131"
  tags       = ["update-svc-acct", "from-tf"]
  policy_ids = [for policy in data.tdh_policies.all.list : policy.id if contains(local.policies, policy.name)]

  //Oauth app details
  oauth_app = {
    description = " description1"
    ttl_spec = {
      ttl       = "1"
      time_unit = "HOURS"
    }
  }
}