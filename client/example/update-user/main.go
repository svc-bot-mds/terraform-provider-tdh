package main

import (
	"errors"
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
)

func main() {
	host := "TDH_HOST_URL"
	client, err := tdh.NewClient(&host, &model.ClientAuth{
		OAuthAppType: oauth_type.UserCredentials,
		Username:     "USERNAME",
		Password:     "PASSWORD",
		OrgId:        "ORG_ID",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.CustomerMetadata.UpdateUser("df4b263e-86e6-40c2-8705-350906ddafda", &customer_metadata.UserUpdateRequest{
		//PolicyIds:   []string{"644a14ac4efa951adae6b7d3"},
		Tags: []string{"client-test"},
		ServiceRoles: []customer_metadata.RolesRequest{
			{RoleId: "ManagedDataService:Admin"},
		},
	})

	fmt.Println(err)
	apiErr := core.ApiError{}
	fmt.Println(err != nil && errors.As(err, &apiErr), apiErr)
}
