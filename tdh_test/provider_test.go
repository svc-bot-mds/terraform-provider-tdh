package tdh_test

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the tdh client is properly configured.
	providerConfig = `
provider "tdh" {
   host     = "TDH_HOST_URL"
   username = "TDH_USERNAME"
   password = "TDH_PASSWORD"
   org_id   = "TDH_ORG_ID"
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"tdh": providerserver.NewProtocol6WithError(tdh.New()),
	}
)
