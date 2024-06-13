package tdh_test

import (
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMdsPoliciesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_policies" "policies" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_policies.policies", "id"),
					resource.TestCheckResourceAttr("data.tdh_policies.policies", "list.#", "27"),
					resource.TestCheckResourceAttr("data.tdh_policies.policies", "list.0.name", "test-tfddwqe"),
					resource.TestCheckResourceAttr("data.tdh_policies.policies", "id", common.DataSource+common.PoliciesId),
				),
			},
		},
	})
}
