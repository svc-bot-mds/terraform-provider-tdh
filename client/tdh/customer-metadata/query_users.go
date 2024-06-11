package customer_metadata

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type UsersQuery struct {
	AccountType string   `schema:"accountType"`
	Emails      []string `schema:"email,omitempty"`
	Names       []string `schema:"name,omitempty"`
	model.PageQuery
}
