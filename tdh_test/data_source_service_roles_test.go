package tdh_test

import (
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMdsServiceRolesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_service_roles" "roles"{
  type = "RABBITMQ"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_service_roles.roles", "id"),
					resource.TestCheckResourceAttrSet("data.tdh_service_roles.roles", "type"),
					resource.TestCheckResourceAttr("data.tdh_service_roles.roles", "roles.#", "6"),
					resource.TestCheckResourceAttr("data.tdh_service_roles.roles", "roles.0.name", "write"),
					resource.TestCheckResourceAttr("data.tdh_service_roles.roles", "roles.0.role_id", "sample"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.tdh_service_roles.roles", "id", common.DataSource+common.ServiceRolesId),
				),
			},
		},
	})
}
