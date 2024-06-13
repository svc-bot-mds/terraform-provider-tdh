package infra_connector

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type CertificateQuery struct {
	Name     string `schema:"name,omitempty"`
	Provider string `schema:"provider,omitempty"`
	model.PageQuery
}
