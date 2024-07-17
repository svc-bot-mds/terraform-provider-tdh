package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/validators"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *tdh.Client
}

type userResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Email         types.String `tfsdk:"email"`
	Status        types.String `tfsdk:"status"`
	Username      types.String `tfsdk:"username"`
	PolicyIds     types.Set    `tfsdk:"policy_ids"`
	RoleIds       []string     `tfsdk:"role_ids"`
	ServiceRoles  types.List   `tfsdk:"service_roles"`
	OrgRoles      types.List   `tfsdk:"org_roles"`
	Tags          types.Set    `tfsdk:"tags"`
	DeleteFromIdp types.Bool   `tfsdk:"delete_from_idp"`
	Organizations types.Set    `tfsdk:"organizations"`
	InviteLink    types.String `tfsdk:"invite_link"`
}

type RolesModel struct {
	ID   types.String `tfsdk:"role_id"`
	Name types.String `tfsdk:"name"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *userResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "INIT__Schema")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Represents an User registered on TDH, can be used to create/update/delete/import an user. Only `SRE` can create another `SRE` user by not passing values to `organizations` field. The only operation allowed for `SRE` is creation of an `Organization/SRE` user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Auto-generated ID after creating an user, and can be passed to import an existing user from TDH to terraform state.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Updating the email results in deletion of existing user and new user with updated email/name is created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delete_from_idp": schema.BoolAttribute{
				MarkdownDescription: "Setting this to `true` will completely delete user from IDP, else only service roles will be removed. By default the value is set to `false` during the user deletion",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				Description: "Active status of user on TDH.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "Short name of user.",
				Computed:    true,
			},
			"policy_ids": schema.SetAttribute{
				Description: "IDs of service policies to be associated with user.",
				Optional:    true,
				Computed:    false,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validators.UUIDValidator{},
					),
				},
			},
			"role_ids": schema.SetAttribute{
				MarkdownDescription: "One or more of (Admin, Developer, Viewer, Operator, Compliance Manager). Please make use of `datasource_roles` to get role_ids. This is a mandatory for the User creation with Non-SRE credentials",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.SetAttribute{
				Description: "Tags or labels to categorise users for ease of finding.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"service_roles": schema.ListNestedAttribute{
				Description: "Roles that determines access level inside services on TDH.",
				Computed:    true,
				Optional:    false,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role_id": schema.StringAttribute{
							Description: "ID of the role.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the role.",
							Computed:    true,
						},
					},
				},
			},
			"org_roles": schema.ListNestedAttribute{
				Description: "Roles that determines access level of the user on TDH.",
				Computed:    true,
				Optional:    false,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role_id": schema.StringAttribute{
							Description: "ID of the role.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the role.",
							Computed:    true,
						},
					},
				},
			},
			"organizations": schema.SetAttribute{
				MarkdownDescription: "Set of Organizations Ids. This field is used only by SRE for a creation of the organization users. Use the organization with 	`sre_org` flag set to false",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"invite_link": schema.StringAttribute{
				Description: "User Invite Link . Only visible for SRE",
				Computed:    true,
			},
		},
	}

	tflog.Info(ctx, "END__Schema")
}

// Create a new resource
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__Create")
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)

	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	rolesReq := make([]customer_metadata.RolesRequest, len(plan.RoleIds))
	for i, roleId := range plan.RoleIds {
		rolesReq[i] = customer_metadata.RolesRequest{
			RoleId: roleId,
		}
	}

	// Generate API request body from plan
	userRequest := customer_metadata.CreateUserRequest{
		Usernames: []string{plan.Email.ValueString()},
	}
	if r.client.Root.IsSre {
		plan.Organizations.ElementsAs(ctx, &userRequest.Organizations, true)
	}
	if (r.client.Root.IsSre && !plan.Organizations.IsNull()) || !r.client.Root.IsSre {
		userRequest.ServiceRoles = rolesReq
		plan.PolicyIds.ElementsAs(ctx, &userRequest.PolicyIds, true)
	}

	plan.Tags.ElementsAs(ctx, &userRequest.Tags, true)

	if r.client.Root.IsSre {
		if err := r.client.CustomerMetadata.CreateOrgOrSREUser(&userRequest); err != nil {
			if plan.Organizations.IsNull() {
				resp.Diagnostics.AddError(
					"Submitting request to create an Organization User",
					"Could not create Organization User, unexpected error: "+err.Error(),
				)
			} else {
				resp.Diagnostics.AddError(
					"Submitting request to create SRE User",
					"Could not create SRE User, unexpected error: "+err.Error(),
				)
			}
			return
		}
	} else {
		if err := r.client.CustomerMetadata.CreateUser(&userRequest); err != nil {
			resp.Diagnostics.AddError(
				"Submitting request to create User",
				"Could not create User, unexpected error: "+err.Error(),
			)
			return
		}
	}

	if r.client.Root.IsSre {

	}

	userQuery := customer_metadata.UsersQuery{}
	if r.client.Root.IsSre {
		userQuery.Email = plan.Email.ValueString()
	} else {
		userQuery.Emails = []string{plan.Email.ValueString()}
	}
	users, err := r.client.CustomerMetadata.GetUsers(&userQuery, r.client.Root.IsSre)
	tflog.Info(ctx, "yo resp: ", map[string]interface{}{
		"roles": users,
	})
	if err != nil {
		resp.Diagnostics.AddError("Fetching user",
			"Could not fetch users, unexpected error: "+err.Error(),
		)
		return
	}
	if users.Page.TotalElements == 0 {
		resp.Diagnostics.AddError("Fetching user",
			fmt.Sprintf("Could not find any user by email [%s], server error must have occurred while creating user.", plan.Email.ValueString()),
		)
		return
	}

	if len(*users.Get()) <= 0 {
		resp.Diagnostics.AddError("Fetching User",
			"Unable to fetch the created user",
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	createdUser := &(*users.Get())[0]

	if saveFromUserResponse(&ctx, &resp.Diagnostics, &plan, createdUser, r.client.Root.IsSre) != 0 {
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

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")

	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client.Root.IsSre {
		resp.Diagnostics.AddError(
			"Updating TDH User",
			"SRE Cannot update the user details",
		)
		return
	}
	// Generate API request body from plan
	updateRequest := customer_metadata.UserUpdateRequest{}
	plan.Tags.ElementsAs(ctx, &updateRequest.Tags, true)
	plan.PolicyIds.ElementsAs(ctx, &updateRequest.PolicyIds, true)
	if plan.Status.ValueString() != "INVITED" {
		rolesReq := make([]customer_metadata.RolesRequest, len(plan.RoleIds))
		for i, roleId := range plan.RoleIds {
			rolesReq[i] = customer_metadata.RolesRequest{
				RoleId: roleId,
			}
		}
		tflog.Info(ctx, "Setting serviceRoles for update: ", map[string]interface{}{
			"roles": rolesReq,
		})
		if len(rolesReq) > 0 {
			updateRequest.ServiceRoles = rolesReq
		}
	}

	// Update existing user
	if err := r.client.CustomerMetadata.UpdateUser(plan.ID.ValueString(), &updateRequest); err != nil {
		resp.Diagnostics.AddError(
			"Updating TDH User",
			"Could not update user, unexpected error: "+err.Error(),
		)
		return
	}

	user, err := r.client.CustomerMetadata.GetUser(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Fetching User",
			"Could not fetch users while updating, unexpected error: "+err.Error(),
		)
		return
	}

	//Update resource state with updated items and timestamp
	if saveFromUserResponse(&ctx, &resp.Diagnostics, &plan, user, false) != 0 {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Update")
}

func (r *userResource) Delete(ctx context.Context, request resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "INIT__Delete")
	// Get current state
	var state userResourceModel
	diags := request.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client.Root.IsSre {
		resp.Diagnostics.AddError(
			"Deleting TDH User",
			"SRE Cannot delete the user",
		)
		return
	}

	// Submit request to delete TDH Cluster
	err := r.client.CustomerMetadata.DeleteUser(state.ID.ValueString(), &customer_metadata.DeleteUserQuery{
		DeleteFromIdp: state.DeleteFromIdp.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Deleting TDH User",
			"Could not delete TDH User by ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "END__Delete")
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := &customer_metadata.UsersQuery{}
	if r.client.Root.IsSre {
		user, err := r.client.CustomerMetadata.GetUsers(query, r.client.Root.IsSre)
		if err != nil {
			resp.Diagnostics.AddError(
				"Reading TDH user",
				"Could not read TDH user ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		if user.Page.TotalPages > 1 {
			for i := 1; i <= user.Page.TotalPages; i++ {
				query.PageQuery.Index = i - 1
				page, err := r.client.CustomerMetadata.GetUsers(query, r.client.Root.IsSre)
				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to Read User",
						err.Error(),
					)
					return
				}

				for _, dto := range *page.Get() {
					if dto.Id == state.ID.ValueString() {
						if saveFromUserResponse(&ctx, &resp.Diagnostics, &state, &dto, false) != 0 {
							return
						}
					}
				}
			}

		} else {
			for _, dto := range *user.Get() {
				if dto.Id == state.ID.ValueString() {
					if saveFromUserResponse(&ctx, &resp.Diagnostics, &state, &dto, false) != 0 {
						return
					}
				}
			}
		}
	} else {
		// Get refreshed cluster value from TDH
		user, err := r.client.CustomerMetadata.GetUser(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Reading TDH user",
				"Could not read TDH user ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		// Overwrite items with refreshed state
		if saveFromUserResponse(&ctx, &resp.Diagnostics, &state, user, false) != 0 {
			return
		}
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Read")
}

func saveFromUserResponse(ctx *context.Context, diagnostics *diag.Diagnostics, state *userResourceModel, user *model.User, isSRE bool) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan", map[string]interface{}{"user": *user})

	roles, diags := convertFromRolesDto(ctx, &user.ServiceRoles)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.ServiceRoles = roles

	roles, diags = convertFromRolesDto(ctx, &user.OrgRoles)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.OrgRoles = roles
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}

	state.ID = types.StringValue(user.Id)
	state.Email = types.StringValue(user.Email)
	state.Status = types.StringValue(user.Status)
	tags, diags := types.SetValueFrom(*ctx, types.StringType, user.Tags)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.Tags = tags
	state.Username = types.StringValue(user.Name)
	if isSRE {
		state.InviteLink = types.StringValue(user.InviteLink)
	}
	return 0
}

func convertFromRolesDto(ctx *context.Context, roles *[]model.RoleMini) (types.List, diag.Diagnostics) {
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
