package controller

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type RestoreQuery struct {
	ServiceType string `schema:"serviceType,omitempty"`
	model.PageQuery
}
