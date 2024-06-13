package task

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type TasksQuery struct {
	ResourceName string `schema:"resourceName"`
	model.PageQuery
}
