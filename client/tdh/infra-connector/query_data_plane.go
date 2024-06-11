package infra_connector

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type DataPlaneQuery struct {
	Name string `schema:"name,omitempty"`
	model.PageQuery
}
