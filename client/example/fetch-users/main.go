package main

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/account_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
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

	response, err := client.CustomerMetadata.GetUsers(&customer_metadata.UsersQuery{
		AccountType: account_type.USER_ACCOUNT,
		Emails:      []string{"admin@vmware.com", "developer@vmware.com"},
	})

	fmt.Println(response.Get())
	for _, dto := range *response.Get() {
		fmt.Println(dto)
	}
}
