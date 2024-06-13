package main

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	customer_metadata "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
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

	response, err := client.CustomerMetadata.GetLocalUsers(&customer_metadata.LocalUsersQuery{})

	fmt.Println(response)
}
