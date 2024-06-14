package customer_metadata

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type LocalUsersQuery struct {
	Username string `schema:"username,omitempty"`
	model.PageQuery
}
