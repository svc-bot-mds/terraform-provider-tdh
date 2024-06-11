provider "tdh" {
  host = "https://console.tdh.vmware.com"

  // Get the authentication with "username and password"
  type = "user_creds"

  username = "< Username >"
  password = "< Password >"
  org_id   = "< ORG_ID >"
}