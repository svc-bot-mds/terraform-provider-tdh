terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host = "https://console.tdh.vmware.com"

  // Authentication using username and password
  type     = "user_creds"
  username = "TDH_USERNAME"
  password = "TDH_PASSWORD"
  org_id   = "TDH_ORG_ID"
}