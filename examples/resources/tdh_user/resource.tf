resource "tdh_user" "example" {
  email      = "developer11@vmware.com"
  tags       = ["new-user", "viewer"]
  role_ids   = ["tdh:viewer"]
  policy_ids = ["f3c49288-7b17-4e78-a6af-257b49e35e53"]

  // non editable fields
  lifecycle {
    ignore_changes = [email]
  }
}