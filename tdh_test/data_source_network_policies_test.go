package tdh_test

import (
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMdsNetworkPoliciesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_network_policies" "network_policies" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_network_policies.network_policies", "id"),
					resource.TestCheckResourceAttr("data.tdh_network_policies.network_policies", "policies.#", "4"),
					resource.TestCheckResourceAttr("data.tdh_network_policies.network_policies", "policies.0.name", "open-to-all"),
					resource.TestCheckResourceAttr("data.tdh_network_policies.network_policies", "id", common.DataSource+common.NetworkPoliciesId),
				),
			},
		},
	})
}
