resource "tdh_smtp" "custom" {
  host      = "<<host>>"
  port      = "<<port>>"
  from      = "<<email_address>>"
  user_name = "<<username>>"
  password  = "<<password>>"
  tls       = true
  auth      = true
}

