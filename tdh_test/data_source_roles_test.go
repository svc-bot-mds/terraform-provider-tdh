package tdh_test

import (
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMdsRolesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_roles" "roles" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_roles.roles", "id"),
					resource.TestCheckResourceAttr("data.tdh_roles.roles", "roles.#", "5"),
					resource.TestCheckResourceAttr("data.tdh_roles.roles", "roles.0.name", "Operator"),
					resource.TestCheckResourceAttr("data.tdh_roles.roles", "roles.0.role_id", "StgManagedDataService:Operator"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.tdh_roles.roles", "id", common.DataSource+common.RolesId),
				),
			},
		},
	})
}
