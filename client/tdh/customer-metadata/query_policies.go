package customer_metadata

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type PoliciesQuery struct {
	Id           string   `schema:"id,omitempty"`
	Type         string   `schema:"serviceType,omitempty"`
	ServiceType  string   `schema:"service,omitempty"`
	IdentityType string   `schema:"identityType,omitempty"`
	Names        []string `schema:"name,omitempty"`
	ResourceId   string   `schema:"resourceId,omitempty"`
	model.PageQuery
}
