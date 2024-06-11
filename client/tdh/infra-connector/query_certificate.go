package infra_connector

import "github.com/svc-bot-mds/terraform-provider-tdh/client/model"

type CertificateQuery struct {
	Name string `json:"name,omitempty"`
	model.PageQuery
}
