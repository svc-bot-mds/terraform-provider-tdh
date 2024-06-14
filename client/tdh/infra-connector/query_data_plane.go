package infra_connector

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type DataPlaneQuery struct {
	Name string `schema:"dataplaneName,omitempty"`
	model.PageQuery
}

type EligibleSharedDataPlaneQuery struct {
	Provider          string `schema:"provider"`
	InfraResourceType string `schema:"infraResourceType"`
	model.PageQuery
}

type EligibleDedicatedDataPlaneQuery struct {
	Provider          string `schema:"provider"`
	InfraResourceType string `schema:"infraResourceType"`
	OrgId             string `schema:"orgId"`
	model.PageQuery
}
