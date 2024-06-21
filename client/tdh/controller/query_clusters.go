package controller

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type ClustersQuery struct {
	ServiceType   string `schema:"serviceType,omitempty"`
	Name          string `schema:"name,omitempty"`
	FullNameMatch bool   `schema:"MATCH_FULL_WORD,omitempty"`
	model.PageQuery
}
