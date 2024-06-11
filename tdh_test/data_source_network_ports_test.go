package tdh_test

import (
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMdsNetworkPortsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_network_ports" "all" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_network_ports.all", "id"),
					resource.TestCheckResourceAttr("data.tdh_network_ports.all", "network_ports.#", "5"),
					resource.TestCheckResourceAttr("data.tdh_network_ports.all", "network_ports.0.name", "Metrics"),
					resource.TestCheckResourceAttr("data.tdh_network_ports.all", "network_ports.0.port", "4455"),
					resource.TestCheckResourceAttr("data.tdh_network_ports.all", "id", common.DataSource+common.NetworkPortsId),
				),
			},
		},
	})
}
