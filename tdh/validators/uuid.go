package validators

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = &UUIDValidator{}

type UUIDValidator struct {
	expressions path.Expressions
}

func (s UUIDValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Must be a valid UUID")
}

func (s UUIDValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Must be a valid UUID")
}

func (s UUIDValidator) ValidateString(_ context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	id := request.ConfigValue.ValueString()
	_, err := uuid.Parse(id)
	if err != nil {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid ID", err.Error())
		return
	}
}
