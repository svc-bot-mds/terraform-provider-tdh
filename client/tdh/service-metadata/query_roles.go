package service_metadata

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type RolesQuery struct {
	Type string `schema:"serviceType,omitempty"`
	model.PageQuery
}
