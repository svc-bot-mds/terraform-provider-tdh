package tdh_test

import (
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMdsPolicyTypesDataSource(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_policy_types" "typesList" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tdh_policy_types.typesList", "policy_types.#", "2"),
					resource.TestCheckResourceAttr("data.tdh_policy_types.typesList", "policy_types.0", "NETWORK"),
					resource.TestCheckResourceAttr("data.tdh_policy_types.typesList", "policy_types.1", "RABBITMQ"),
					resource.TestCheckResourceAttr("data.tdh_policy_types.typesList", "id", common.DataSource+common.PolicyTypesId),
				),
			},
		},
	})
}
