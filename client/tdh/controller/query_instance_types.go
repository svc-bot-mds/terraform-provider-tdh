package controller

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type InstanceTypesQuery struct {
	ServiceType string `schema:"serviceType"`
	model.PageQuery
}
