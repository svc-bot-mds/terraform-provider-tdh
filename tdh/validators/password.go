package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
)

var _ validator.String = &PasswordValidator{}

type PasswordValidator struct {
	expressions path.Expressions
}

func (s PasswordValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Must be a valid Password")
}

func (s PasswordValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Must be a valid Password")
}

func (s PasswordValidator) ValidateString(_ context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	password := request.ConfigValue.ValueString()

	pwdPatternUpperCaseRegex := regexp.MustCompile(`[A-Z]+`)
	pwdPatternSplCharRegex := regexp.MustCompile(`[!@#$%^&*()_+=[{|}',./:;<>?\x60\x5D~-]+`)
	pwdPatternNonAsciiRegex := regexp.MustCompile(`[^[:ascii:]]+`)

	// Check if the value matches the pattern
	if !pwdPatternUpperCaseRegex.MatchString(password) {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid Password", "Must contain at least one uppercase character")
		return
	}
	if !pwdPatternSplCharRegex.MatchString(password) {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid Password", "Must contain at least one special character (!@#$%^&*()_+=[-{|}',./:;<>?`~)")
		return
	}
	if pwdPatternNonAsciiRegex.MatchString(password) {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid Password", "should not have high ASCII characters")
		return
	}
}
