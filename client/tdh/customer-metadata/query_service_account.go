package customer_metadata

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type ServiceAccountsQuery struct {
	AccountType string   `schema:"accountType,omitempty"`
	Names       []string `schema:"name,omitempty"`
	model.PageQuery
}
