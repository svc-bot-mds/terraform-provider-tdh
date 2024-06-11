package tdh_test

import (
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMdsServiceAccountsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_service_accounts" "service_accounts" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_service_accounts.service_accounts", "id"),
					resource.TestCheckResourceAttr("data.tdh_service_accounts.service_accounts", "service_accounts.#", "3"),
					resource.TestCheckResourceAttr("data.tdh_service_accounts.service_accounts", "service_accounts.0.name", "test-svc-tf-update1"),
					resource.TestCheckResourceAttr("data.tdh_service_accounts.service_accounts", "service_accounts.0.status", "ACTIVE"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.tdh_service_accounts.service_accounts", "id", common.DataSource+common.ServiceAccountsId),
				),
			},
		},
	})
}
