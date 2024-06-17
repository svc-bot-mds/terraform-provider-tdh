package controller

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type ClusterBackupQuery struct {
	ID string `schema:"id"`
	model.PageQuery
}
