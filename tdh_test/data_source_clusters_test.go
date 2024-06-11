package tdh_test

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"
)

func TestMdsClustersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_clusters" "cluster_list"{
  											service_type = "RABBITMQ"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_clusters.cluster_list", "id"),
					resource.TestCheckResourceAttrSet("data.tdh_clusters.cluster_list", "service_type"),
					resource.TestCheckResourceAttr("data.tdh_clusters.cluster_list", "clusters.#", "26"),
					resource.TestCheckResourceAttr("data.tdh_clusters.cluster_list", "clusters.0.name", "audit-test-dnd"),
					resource.TestCheckResourceAttr("data.tdh_clusters.cluster_list", "id", common.DataSource+common.ClusterId),
				),
			},
		},
	})
}
