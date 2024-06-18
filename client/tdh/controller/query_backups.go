package controller

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type BackupsQuery struct {
	ServiceType string `schema:"serviceType,omitempty"`
	Name        string `schema:"name,omitempty"`
	ClusterId   string `schema:"clusterId,omitempty"`
	model.PageQuery
}
