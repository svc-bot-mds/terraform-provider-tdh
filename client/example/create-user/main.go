package main

import (
	"errors"
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/account_type"
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

	err = client.CustomerMetadata.CreateUser(&customer_metadata.CreateUserRequest{
		AccountType: account_type.USER_ACCOUNT,
		Usernames:   []string{"developer@vmware.com"},
		PolicyIds:   []string{"6446112a8710fc120cbdc8ff", "6438cbd364740d4d48dc2673"},
		ServiceRoles: []customer_metadata.RolesRequest{
			{RoleId: "ManagedDataService:Developer"},
			{RoleId: "ManagedDataService:Admin"},
		},
	})

	fmt.Println(err)
	apiErr := core.ApiError{}
	fmt.Println(errors.As(err, &apiErr), apiErr)
}
