package infra_connector

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type CloudAccountsQuery struct {
	AccountType string `schema:"accountType,omitempty"`
	Name        string `schema:"name, omitempty"`
	model.PageQuery
}
