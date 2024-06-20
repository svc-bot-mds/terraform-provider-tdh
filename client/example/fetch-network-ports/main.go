package main

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/service-metadata"
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
		fmt.Println("error on getting token")
		fmt.Println(err)
		return
	}

	query := service_metadata.NetworkPortsQuery{}
	fmt.Printf("query: %+v\n", query)
	response, err := client.ServiceMetadata.GetNetworkPorts(&query)
	if err != nil {
		panic(err)
	}
	fmt.Println(response)
}
