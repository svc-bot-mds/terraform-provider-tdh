package main

import (
	"context"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name tdh

func main() {
	err := providerserver.Serve(context.Background(), tdh.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/svc-bot-mds/tdh",
	})
	if err != nil {
		return
	}
}
