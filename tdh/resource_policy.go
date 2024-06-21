package tdh

import (
	"context"
	"fmt"
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
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
	"sort"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &policyResource{}
	_ resource.ResourceWithConfigure   = &policyResource{}
	_ resource.ResourceWithImportState = &policyResource{}
)

func NewPolicyResource() resource.Resource {
	return &policyResource{}
}

type policyResource struct {
	client *tdh.Client
}

type policyResourceModel struct {
	ID              types.String           `tfsdk:"id"`
	Name            types.String           `tfsdk:"name"`
	Description     types.String           `tfsdk:"description"`
	ServiceType     types.String           `tfsdk:"service_type"`
	PermissionSpecs []*PermissionSpecModel `tfsdk:"permission_specs"`
	ResourceIds     types.Set              `tfsdk:"resource_ids"`
}

type PermissionSpecModel struct {
	Role        types.String `tfsdk:"role"`
	Resource    types.String `tfsdk:"resource"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func (r *policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *policyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *policyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "INIT__Schema")

	resp.Schema = schema.Schema{
		Description: "Represents a policy on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Auto-generated ID of the policy after creation, and can be used to import it from TDH to terraform state.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the policy.",
				Optional:    true,
				Computed:    true,
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of TDH service to managed. Supported values: %s.", supportedServiceTypesMarkdown()),
				Required:            true,
				Validators: []validator.String{
					utils.ServiceTypeValidator,
				},
			},
			"resource_ids": schema.SetAttribute{
				Description: "IDs of service resources/instances being managed by the policy.",
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"permission_specs": schema.SetNestedAttribute{
				MarkdownDescription: "Permissions to enforce on service resources. Only required for policies other than `NETWORK` type.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							MarkdownDescription: "Name of the role, it will vary on service type. Please make use of datasource `tdh_service_roles` to get the relevant values by a service type, `POSTGRES` for example.",
							Required:            true,
						},
						"resource": schema.StringAttribute{
							MarkdownDescription: "Name of the cluster/instance. Please make use of datasource `tdh_clusters` to get the names & `tdh_cluster_metadata` to get cluster service specific resources like databases, schemas, vhosts etc.<br>" +
								"Format of this field is:<br>" +
								"- `cluster:<NAME>[/database:<DB_NAME>[/schema:<SCHEMA>[/table:<TABLE>]]]` for `POSTGRES`" +
								"- `cluster:<NAME>[/database:<DB_NAME>[/table:<TABLE>[/columns:<COMMA_SEPARATED_COLUMNS>]]]` for `MYSQL`" +
								"- `cluster:<NAME>[/vhost:<VHOST>[/queue:<QUEUE>]]` for `RABBITMQ`" +
								"- `cluster:<NAME>` for `REDIS`",
							Required: true,
						},
						"permissions": schema.SetAttribute{
							MarkdownDescription: "Name of the permission, usually same as role name. Please make use of datasource `tdh_service_roles` to get the relevant values.",
							Required:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}

	tflog.Info(ctx, "END__Schema")
}

// Create a new resource
func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__Create")
	// Retrieve values from plan
	var plan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)

	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	rolesReq := make([]customer_metadata.PermissionSpec, len(plan.PermissionSpecs))

	// Generate API request body from plan
	policyRequest := customer_metadata.CreateUpdatePolicyRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		ServiceType: plan.ServiceType.ValueString(),
	}
	for i, roleId := range plan.PermissionSpecs {

		rolesReq[i] = customer_metadata.PermissionSpec{
			Role:     roleId.Role.ValueString(),
			Resource: roleId.Resource.ValueString(),
		}
		roleId.Permissions.ElementsAs(ctx, &rolesReq[i].Permissions, true)
	}
	sort.Slice(rolesReq, func(i, j int) bool { return rolesReq[i].Role < rolesReq[j].Role })
	policyRequest.PermissionsSpec = rolesReq

	tflog.Debug(ctx, "Create Policy DTO", map[string]interface{}{"dto": policyRequest})
	if _, err := r.client.CustomerMetadata.CreatePolicy(&policyRequest); err != nil {

		resp.Diagnostics.AddError(
			"Submitting request to create Policy",
			"Could not create Policy, unexpected error: "+err.Error(),
		)
		return
	}

	policies, err := r.client.CustomerMetadata.GetPolicies(&customer_metadata.PoliciesQuery{
		Type:  plan.ServiceType.ValueString(),
		Names: []string{plan.Name.ValueString()},
	})
	if err != nil {
		resp.Diagnostics.AddError("Fetching Policy",
			"Could not fetch policy, unexpected error: "+err.Error(),
		)
		return
	}

	if len(*policies.Get()) <= 0 {
		resp.Diagnostics.AddError("Fetching Policy",
			"Unable to fetch the created policy",
		)
		return
	}
	createdPolicy := &(*policies.Get())[0]
	tflog.Debug(ctx, "Created Policy DTO", map[string]interface{}{"dto": createdPolicy})
	if saveFromPolicyResponse(&ctx, &resp.Diagnostics, &plan, createdPolicy) != 0 {
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

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")

	// Retrieve values from plan
	var plan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	updateRequest := customer_metadata.CreateUpdatePolicyRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		ServiceType: plan.ServiceType.ValueString(),
	}
	permissionSpecRequest := make([]customer_metadata.PermissionSpec, len(plan.PermissionSpecs))

	for i, roleId := range plan.PermissionSpecs {
		permissionSpecRequest[i] = customer_metadata.PermissionSpec{
			Role:     roleId.Role.ValueString(),
			Resource: roleId.Resource.ValueString(),
		}
		roleId.Permissions.ElementsAs(ctx, &permissionSpecRequest[i].Permissions, true)
	}
	updateRequest.PermissionsSpec = permissionSpecRequest
	tflog.Debug(ctx, "update policy request dto", map[string]interface{}{"dto": updateRequest})

	// Update existing policy
	if err := r.client.CustomerMetadata.UpdatePolicy(plan.ID.ValueString(), &updateRequest); err != nil {
		resp.Diagnostics.AddError(
			"Updating TDH Policy",
			"Could not update Policy, unexpected error: "+err.Error(),
		)
		return
	}

	policy, err := r.client.CustomerMetadata.GetPolicy(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Fetching Policy",
			"Could not fetch policy while updating, unexpected error: "+err.Error(),
		)
		return
	}

	//Update resource state with updated items and timestamp
	if saveFromPolicyResponse(&ctx, &resp.Diagnostics, &plan, policy) != 0 {
		return
	}

	//resp.State.RemoveResource(ctx)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Update")
}

func (r *policyResource) Delete(ctx context.Context, request resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "INIT__Delete")
	// Get current state
	var state policyResourceModel
	diags := request.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Submit request to delete TDH Policy
	err := r.client.CustomerMetadata.DeletePolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Deleting TDH Policy",
			"Could not delete TDH Policy by ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "END__Delete")
}

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state policyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed policy value from TDH
	policy, err := r.client.CustomerMetadata.GetPolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading TDH Policy",
			"Could not read TDH policy ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	if saveFromPolicyResponse(&ctx, &resp.Diagnostics, &state, policy) != 0 {
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

func saveFromPolicyResponse(ctx *context.Context, diagnostics *diag.Diagnostics, state *policyResourceModel, policy *model.Policy) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan", map[string]interface{}{"policy": *policy})
	tfPermissionSpecModel, diags := convertFromPermissionSpecDto(ctx, policy.PermissionsSpec)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}

	sort.Slice(tfPermissionSpecModel, func(i, j int) bool {
		roleI := tfPermissionSpecModel[i].Role.ValueString()
		roleJ := tfPermissionSpecModel[j].Role.ValueString()
		return roleI < roleJ
	})
	state.PermissionSpecs = tfPermissionSpecModel

	state.ID = types.StringValue(policy.ID)
	state.Name = types.StringValue(policy.Name)
	state.ServiceType = types.StringValue(policy.ServiceType)
	state.Description = types.StringValue(policy.Description)
	resourceIds, diags := types.SetValueFrom(*ctx, types.StringType, policy.ResourceIds)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.ResourceIds = resourceIds

	return 0
}

func convertFromPermissionSpecDto(ctx *context.Context, roles []*model.PermissionsSpec) ([]*PermissionSpecModel, diag.Diagnostics) {
	tfPermissionSpecModels := make([]*PermissionSpecModel, len(roles))
	for i, permissionSpec := range roles {
		tfPermissionSpecModels[i] = &PermissionSpecModel{
			Role:     types.StringValue(permissionSpec.Role),
			Resource: types.StringValue(permissionSpec.Resource),
		}
		permissions, _ := types.SetValueFrom(*ctx, types.StringType, []string{permissionSpec.Permissions[0].Name})
		tfPermissionSpecModels[i].Permissions = permissions
	}
	return tfPermissionSpecModels, nil
}
