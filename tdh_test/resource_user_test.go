package tdh_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { /* Set up any prerequisites or check for required dependencies */ },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `locals {
  											account_type  = "USER_ACCOUNT"
  											service_roles = ["Developer", "Admin"]
  											policies = ["gya-policy","eu301"]
										}

										data "tdh_roles" "all" {
										}

										output "roles_data" {
  											value = data.tdh_roles.all
										}

										data "tdh_policies" "policies" {
										}

										output "policies_data" {
  											value = data.tdh_policies.policies
										}

										resource "tdh_user" "temp" {
 		 									email      = "developer-tf-user@vmware.com"
  											tags       = ["new-user-tf", "create-tf-user"]
  											role_ids   = [for role in data.tdh_roles.all.roles : role.role_id if contains(local.service_roles, role.name)]
  											policy_ids = [for policy in data.tdh_policies.policies.policies: policy.id if contains(local.policies, policy.name) ]
  											timeouts   = {
    											create = "1m"
  											}

  											// non editable fields
  											lifecycle {
    											ignore_changes = [email, status]
  											}
										}
data "tdh_users" "users" {
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_users.users", "id"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.0.email", "developer-tf-user@vmware.com"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.0.role_ids.#", "2"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.0.policy_ids.#", "2"),
				),
			},
			{
				Config: providerConfig + `locals {
  											account_type  = "USER_ACCOUNT"
  											service_roles = ["Admin"]
  											policies = ["gya-policy"]
										}

										data "tdh_roles" "all" {
										}

										output "roles_data" {
  											value = data.tdh_roles.all
										}

										data "tdh_policies" "policies" {
										}

										output "policies_data" {
  											value = data.tdh_policies.policies
										}

										resource "tdh_user" "temp" {
 		 									email      = "developer-tf-user@vmware.com"
  											tags       = ["new-user-tf", "update-tf-user"]
  											policy_ids = [for policy in data.tdh_policies.policies.policies: policy.id if contains(local.policies, policy.name) ]
  											timeouts   = {
    											create = "1m"
  											}

  											// non editable fields
  											lifecycle {
    											ignore_changes = [email, status]
  											}
										}
data "tdh_users" "users" {
}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_users.users", "id"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.0.email", "developer-tf-user@vmware.com"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.0.role_ids.#", "1"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.0.policy_ids.#", "1"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.0.tags.0", "update-tf-user"),
				),
			},
			{
				Config: providerConfig,
			},
		},
	})
}
