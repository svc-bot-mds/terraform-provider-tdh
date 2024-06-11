package main

import (
	"errors"
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
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

	_, err = client.Controller.GetCluster("12376yhsjdasd")

	if err != nil {
		fmt.Println(err)
		var apiError core.ApiError
		if errors.As(err, &apiError) {
			fmt.Println("recognized")
			fmt.Println(apiError.ErrorMessage)
		}
		return
	}
}
