package tdh_test

import (
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMdsUsersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "tdh_users" "users" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tdh_users.users", "id"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.#", "10"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "users.0.email", "developer-tf-user@vmware.com"),
					resource.TestCheckResourceAttr("data.tdh_users.users", "id", common.DataSource+common.UsersId),
				),
			},
		},
	})
}
