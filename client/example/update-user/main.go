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
		ApiToken:     "API_TOKEN",
		OAuthAppType: oauth_type.ApiToken,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.CustomerMetadata.UpdateUser("64533d8a2cee5b76e7c5fa70", &customer_metadata.UserUpdateRequest{
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
