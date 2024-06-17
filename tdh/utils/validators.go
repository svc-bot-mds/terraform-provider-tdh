package utils

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var (
	ServiceTypeValidator  = stringvalidator.OneOf("POSTGRES", "MYSQL", "RABBITMQ", "REDIS")
	IdentityTypeValidator = stringvalidator.OneOf("USER_ACCOUNT", "LOCAL_USER_ACCOUNT")
)
