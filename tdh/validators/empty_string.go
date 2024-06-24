package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"strings"
)

var _ validator.String = &EmptyStringValidator{}

type EmptyStringValidator struct {
	expressions path.Expressions
}

func (s EmptyStringValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Must not contain any whitespace characters")
}

func (s EmptyStringValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Must not contain any whitespace characters")
}

func (s EmptyStringValidator) ValidateString(_ context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	stringValue := request.ConfigValue.ValueString()
	if strings.TrimSpace(stringValue) == "" {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid String", "Enter a valid string. Must Must not contain any whitespace characters")
		return
	}
}
