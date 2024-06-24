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
		OAuthAppType: oauth_type.UserCredentials,
		Username:     "USERNAME",
		Password:     "PASSWORD",
		OrgId:        "ORG_ID",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.CustomerMetadata.CreateUser(&customer_metadata.CreateUserRequest{
		AccountType: account_type.USER_ACCOUNT,
		Usernames:   []string{"developer@vmware.com"},
		PolicyIds:   []string{"df4b263e-86e6-40c2-8705-350906ddafda", "e415202b-3967-46a9-a906-76527fd43f21"},
		ServiceRoles: []customer_metadata.RolesRequest{
			{RoleId: "ManagedDataService:Developer"},
			{RoleId: "ManagedDataService:Admin"},
		},
	})

	fmt.Println(err)
	apiErr := core.ApiError{}
	fmt.Println(errors.As(err, &apiErr), apiErr)
}
