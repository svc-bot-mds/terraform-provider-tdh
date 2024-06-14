package tdh_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { /* Set up any prerequisites or check for required dependencies */ },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `locals {
  											account_type  = "SERVICE_ACCOUNT"
  											policies = ["gya-policy","eu301"]
										}

										data "tdh_policies" "policies" {
										}

										output "policies_data" {
  											value = data.tdh_policies.policies
										}

										resource "tdh_service_account" "svc_account" {
  											name = "test-svc-tf-create-sa"
  											tags = ["create-svc-acct","from-tf"]
  											policy_ids =  [for policy in data.tdh_policies.policies.list: policy.id if contains(local.policies, policy.name) ]

  											// non editable fields
  											lifecycle {
   											 ignore_changes = [name]
  											}
									}

data "tdh_service_accounts" "service_accounts" {
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_service_accounts.service_accounts", "id"),
					resource.TestCheckResourceAttr("data.tdh_service_accounts.service_accounts", "service_accounts.0.name", "test-svc-tf-create-sa"),
				),
			},
			{
				Config: providerConfig + `locals {
  											account_type  = "SERVICE_ACCOUNT"
										}

										resource "tdh_service_account" "svc_account" {
  											name = "test-svc-tf-create-sa"
  											tags = ["update-svc-acct"]
  											policy_ids =  [for policy in data.tdh_policies.policies.list: policy.id if contains(local.policies, policy.name) ]

  											// non editable fields
  											lifecycle {
   											 ignore_changes = [name]
  											}
									}
data "tdh_service_accounts" "service_accounts" {
}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_service_accounts.service_accounts", "id"),
					resource.TestCheckResourceAttr("data.tdh_service_accounts.service_accounts", "service_accounts.0.name", "test-svc-tf-create-sa"),
				),
			},
			{
				Config: providerConfig,
			},
		},
	})
}
