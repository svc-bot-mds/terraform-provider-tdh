package utils

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/identity_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/service_type"
)

var (
	ServiceTypeValidator  = stringvalidator.OneOf(service_type.GetAll()...)
	IdentityTypeValidator = stringvalidator.OneOf(identity_type.GetAll()...)
)
