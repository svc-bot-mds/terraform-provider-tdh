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
		OAuthAppType: oauth_type.UserCredentials,
		Username:     "USERNAME",
		Password:     "PASSWORD",
		OrgId:        "ORG_ID",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = client.Controller.GetCluster("e415202b-3967-46a9-a906-76527fd43f21")

	if err != nil {
		fmt.Println(err)
		var apiError core.ApiError
		if errors.As(err, &apiError) {
			fmt.Println(apiError.ErrorMessage)
		}
		return
	}
}
