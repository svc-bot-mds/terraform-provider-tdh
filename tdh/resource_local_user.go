package tdh

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	customer_metadata "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"time"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &localUserResource{}
	_ resource.ResourceWithConfigure   = &localUserResource{}
	_ resource.ResourceWithImportState = &localUserResource{}
)

func NewLocalUserResource() resource.Resource {
	return &localUserResource{}
}

type localUserResource struct {
	client *tdh.Client
}

type localUserResourceModel struct {
	ID        types.String       `tfsdk:"id"`
	Username  types.String       `tfsdk:"username"`
	PolicyIds types.Set          `tfsdk:"policy_ids"`
	Password  *localUserPassword `tfsdk:"password"`
}

type localUserPassword struct {
	Current types.String `tfsdk:"current"`
	New     types.String `tfsdk:"new"`
	Confirm types.String `tfsdk:"confirm"`
}

func (r *localUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_user"
}

func (r *localUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tdh.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tdh.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *localUserResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "INIT__Schema")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Represents an Local User registered on TDH, can be used to create/update/delete/import a local user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Auto-generated ID after creating a local user, and can be passed to import an existing local user from TDH to terraform state.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Updating the username results in deletion of existing local user and new user with updated name is created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_ids": schema.SetAttribute{
				Description: "IDs of service policies to be associated with local user.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"password": schema.SingleNestedAttribute{
				MarkdownDescription: "Used to create or update password. During creation of resource, only `new` and `confirm` are required. All will be required for password reset.",
				Required:            false,
				Optional:            true,
				Sensitive:           true,
				Attributes: map[string]schema.Attribute{
					"current": schema.StringAttribute{
						Description: "Current password of this local user. (Required for changing password)",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.Expressions{
								path.Root("password").AtName("new").Expression(),
								path.Root("password").AtName("confirm").Expression(),
							}...),
						},
					},
					"new": schema.StringAttribute{
						Description: "Password to set for this local user. (Required for creation)",
						Required:    true,
					},
					"confirm": schema.StringAttribute{
						MarkdownDescription: "Confirm the password to match the `new`. (Required for creation & password reset)",
						Required:            true,
					},
				},
			},
		},
	}

	tflog.Info(ctx, "END__Schema")
}

// Create a new resource
func (r *localUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__Create")
	// Retrieve values from plan
	var plan localUserResourceModel
	diags := req.Plan.Get(ctx, &plan)

	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	if r.validateCreateInputs(plan.Password, &resp.Diagnostics).HasError() {
		return
	}
	// Generate API request body from plan
	request := customer_metadata.CreateLocalUserRequest{
		Usernames:       []string{plan.Username.ValueString()},
		Password:        plan.Password.New.ValueString(),
		ConfirmPassword: plan.Password.Confirm.ValueString(),
	}
	plan.PolicyIds.ElementsAs(ctx, &request.PolicyIds, true)
	response, err := r.client.CustomerMetadata.CreateLocalUser(&request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Submitting request",
			"Could not create Local User, unexpected error: "+err.Error(),
		)
		return
	}

	// local user operation usually happens instantly
	time.Sleep(5 * time.Second)
	err = r.pollTaskStatus((*response)[0].TaskId)
	if err != nil {
		resp.Diagnostics.AddError("Creating local user",
			"Could not create local user, error: "+err.Error(),
		)
		return
	}
	users, err := r.client.CustomerMetadata.GetLocalUsers(&customer_metadata.LocalUsersQuery{
		Username: plan.Username.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Fetching local user",
			"Could not fetch local user, unexpected error: "+err.Error(),
		)
		return
	}
	if users.Page.TotalElements == 0 {
		resp.Diagnostics.AddError("Fetching user",
			fmt.Sprintf("Could not find any local user by username [%s], server error must have occurred while creating user.", plan.Username.ValueString()),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	createdUser := &(*users.Get())[0]

	if r.saveFromResponse(&ctx, &resp.Diagnostics, &plan, createdUser) != 0 {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Create")
}

func (r *localUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")

	// Retrieve values from plan
	var plan, state localUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	if r.validateUpdateInputs(state.Password, plan.Password, &resp.Diagnostics).HasError() {
		return
	}
	// Generate API request body from plan
	updateRequest := customer_metadata.LocalUserUpdateRequest{}
	var taskRequired = true
	plan.PolicyIds.ElementsAs(ctx, &updateRequest.PolicyIds, true)
	if plan.Password != nil {
		updateRequest.CurrentPassword = plan.Password.Current.ValueString()
		updateRequest.NewPassword = plan.Password.New.ValueString()
		updateRequest.ConfirmNewPassword = plan.Password.Confirm.ValueString()
		taskRequired = false
	}

	response, err := r.client.CustomerMetadata.UpdateLocalUser(plan.ID.ValueString(), &updateRequest)
	// Update existing user
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating Local User",
			"Could not update local user, unexpected error: "+err.Error(),
		)
		return
	}
	if taskRequired {
		// local user operation usually happens instantly
		time.Sleep(5 * time.Second)
		err = r.pollTaskStatus((*response)[0].TaskId)
		if err != nil {
			resp.Diagnostics.AddError("Updating local user",
				"Could not update local user, error: "+err.Error(),
			)
			return
		}
	}

	user, err := r.client.CustomerMetadata.GetLocalUser(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Fetching Local User",
			"Could not fetch local users while updating, unexpected error: "+err.Error(),
		)
		return
	}

	//Update resource state with updated items and timestamp
	if r.saveFromResponse(&ctx, &resp.Diagnostics, &plan, user) != 0 {
		return
	}

	diags = resp.State.Set(ctx, plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Update")
}

func (r *localUserResource) Delete(ctx context.Context, request resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "INIT__Delete")
	// Get current state
	var state localUserResourceModel
	diags := request.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Submit request to delete TDH Cluster
	response, err := r.client.CustomerMetadata.DeleteLocalUser(state.ID.ValueString())
	if err != nil {
		apiErr := core.ApiError{}
		errors.As(err, &apiErr)
		if apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting Local User",
			"Could not delete Local User by ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// local user operation usually happens instantly
	time.Sleep(5 * time.Second)
	err = r.pollTaskStatus((*response)[0].TaskId)
	if err != nil {
		resp.Diagnostics.AddError("Deleting local user",
			"Could not delete local user, error: "+err.Error(),
		)
		return
	}
	tflog.Info(ctx, "END__Delete")
}

func (r *localUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *localUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state localUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed cluster value from TDH
	user, err := r.client.CustomerMetadata.GetLocalUser(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading Local user",
			"Could not read local user ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	if r.saveFromResponse(&ctx, &resp.Diagnostics, &state, user) != 0 {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Read")
}

func (r *localUserResource) saveFromResponse(ctx *context.Context, diagnostics *diag.Diagnostics, state *localUserResourceModel, user *model.LocalUser) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan", map[string]interface{}{"local-user": *user})

	state.ID = types.StringValue(user.Id)
	state.Username = types.StringValue(user.Username)

	policyIds, diags := types.SetValueFrom(*ctx, types.StringType, user.PolicyIds)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.PolicyIds = policyIds

	return 0
}

func (r *localUserResource) convertFromRolesDto(ctx *context.Context, roles *[]model.RoleMini) (types.List, diag.Diagnostics) {
	tfRoleModels := make([]RolesModel, len(*roles))
	for i, role := range *roles {
		tfRoleModels[i] = RolesModel{
			ID:   types.StringValue(role.RoleID),
			Name: types.StringValue(role.Name),
		}
	}
	return types.ListValueFrom(*ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":    types.StringType,
		"role_id": types.StringType,
	}}, tfRoleModels)
}

func (r *localUserResource) pollTaskStatus(taskId string) error {
	for true {
		taskResponse, err := r.client.TaskService.GetTask(taskId)
		if err != nil {
			return err
		}
		if taskResponse.Status == "SUCCESS" {
			return nil
		} else if taskResponse.Status == "FAILED" {
			return err
		}
		time.Sleep(time.Second * 10)
	}
	return nil
}

func (r *localUserResource) validateCreateInputs(password *localUserPassword, diag *diag.Diagnostics) *diag.Diagnostics {
	if password == nil {
		diag.AddError("Invalid inputs", "'password' is required during create operation.")
		return diag
	}
	if password.New.IsNull() {
		diag.AddError("Invalid inputs", "'new' is required during create operation.")
		return diag
	}
	if password.Confirm.IsNull() {
		diag.AddError("Invalid inputs", "'confirm' is required during create operation.")
		return diag
	}
	if password.New.ValueString() != password.Confirm.ValueString() {
		diag.AddError("Invalid inputs", "'new' must match with 'confirm'.")
		return diag
	}
	return diag
}

func (r *localUserResource) validateUpdateInputs(statePassword *localUserPassword, password *localUserPassword, diag *diag.Diagnostics) *diag.Diagnostics {
	if password == nil {
		return diag
	}
	if password.Current.IsNull() && (statePassword == nil || statePassword.New != password.New) {
		diag.AddError("Invalid inputs", "Password reset requires all of 'current', 'new' & 'confirm'.")
		return diag
	}
	return diag
}
