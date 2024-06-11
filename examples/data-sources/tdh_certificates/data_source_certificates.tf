terraform {
  required_providers {
    tdh = {
      source = "hashicorp.com/svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host      = "https://console.tdh.vmware.com"

  username = " < Username > "
  password = " < Password > "

  type = "user_creds"
}

data "tdh_certificates" "all"{
}
output "resp" {
  value = data.tdh_certificates.all
}

