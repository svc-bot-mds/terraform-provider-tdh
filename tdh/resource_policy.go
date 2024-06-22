package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/service_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
	"reflect"
	"sort"
	"time"
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
	ID              types.String          `tfsdk:"id"`
	Name            types.String          `tfsdk:"name"`
	Description     types.String          `tfsdk:"description"`
	ServiceType     types.String          `tfsdk:"service_type"`
	PermissionSpecs []PermissionSpecModel `tfsdk:"permission_specs"`
	ResourceIds     types.Set             `tfsdk:"resource_ids"`
	Updating        types.Bool            `tfsdk:"updating"`
}

type PermissionSpecModel struct {
	Role       types.String `tfsdk:"role"`
	Resource   types.String `tfsdk:"resource"`
	Permission types.String `tfsdk:"permission"`
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
			"updating": schema.BoolAttribute{
				Description: "Denotes whether there is any task running on policy.",
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
				MarkdownDescription: "Permission to enforce on service resources. Only required for policies other than `NETWORK` type.",
				Required:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					CustomType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"role":       types.StringType,
							"resource":   types.StringType,
							"permission": types.StringType,
						},
					},
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							MarkdownDescription: "Name of the role, it will vary on service type. Please make use of datasource `tdh_service_roles` to get the relevant values by a service type, `POSTGRES` for example.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"resource": schema.StringAttribute{
							MarkdownDescription: "Name of the cluster/instance. Please make use of datasource `tdh_clusters` to get the names & `tdh_cluster_metadata` to get cluster service specific resources like databases, schemas, vhosts etc.\n" +
								"Format of this field is:\n" +
								"- `cluster:<NAME>[/database:<DB_NAME>[/schema:<SCHEMA>[/table:<TABLE>]]]` for `POSTGRES`\n" +
								"- `cluster:<NAME>[/database:<DB_NAME>[/table:<TABLE>[/columns:<COMMA_SEPARATED_COLUMNS>]]]` for `MYSQL`\n" +
								"- `cluster:<NAME>[/vhost:<VHOST>[/queue:<QUEUE>]]` for `RABBITMQ`\n" +
								"- `cluster:<NAME>` for `REDIS`",
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"permission": schema.StringAttribute{
							MarkdownDescription: "Name of the permission, usually same as role name. Please make use of datasource `tdh_service_roles` to get the relevant values.\n" +
								"## Notes\n" +
								"- Optional, must be same as role for services other than `REDIS`\n" +
								"- Required for `REDIS` policy. It has to be extracted from `permission_id` of `tdh_service_roles` datasource.\n" +
								"Ex: If `permission_id` is \"mds:redis:+@read\", fill the value \"+@read\", similarly for other permissions. " +
								"**Note:** When `permission_id` is \"mds:redis:custom\", you can define a custom valid Redis rule.",
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
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

	if err := r.validateSpecs(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid input", err.Error())
		return
	}

	// Generate API request body from plan
	policyRequest := customer_metadata.CreateUpdatePolicyRequest{
		Name:            plan.Name.ValueString(),
		Description:     plan.Description.ValueString(),
		ServiceType:     plan.ServiceType.ValueString(),
		PermissionsSpec: *r.convertFromSpecsTfModel(&plan.PermissionSpecs),
	}

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
	var state, plan, postPlan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	if err := r.validateSpecs(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid input", err.Error())
		return
	}
	// Generate API request body from plan
	updateRequest := customer_metadata.CreateUpdatePolicyRequest{
		Name:            plan.Name.ValueString(),
		Description:     plan.Description.ValueString(),
		ServiceType:     plan.ServiceType.ValueString(),
		PermissionsSpec: *r.convertFromSpecsTfModel(&plan.PermissionSpecs),
	}
	tflog.Debug(ctx, "update policy request dto", map[string]interface{}{"dto": updateRequest})

	// Update existing policy
	var policy *model.Policy
	policy, err := r.client.CustomerMetadata.UpdatePolicy(plan.ID.ValueString(), &updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating TDH Policy",
			"Could not update Policy, unexpected error: "+err.Error(),
		)
		return
	}
	for policy.Updating {
		time.Sleep(5 * time.Second)
		policy, err = r.client.CustomerMetadata.GetPolicy(plan.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Fetching Policy",
				"Could not fetch policy while updating, unexpected error: "+err.Error(),
			)
			return
		}
	}

	//Update resource state with updated items and timestamp
	if saveFromPolicyResponse(&ctx, &resp.Diagnostics, &postPlan, policy) != 0 {
		return
	}
	tflog.Debug(ctx, "checking something", map[string]interface{}{
		"postPlan": *r.convertFromSpecsTfModel(&postPlan.PermissionSpecs),
		"plan":     updateRequest.PermissionsSpec,
	})
	// if both differs even after update, means something failed
	if !reflect.DeepEqual(*r.convertFromSpecsTfModel(&postPlan.PermissionSpecs), updateRequest.PermissionsSpec) {
		resp.Diagnostics.AddError("Updating Policy", "Policy has failed to update. For more info, you can query datasource 'tdh_task' with cluster name(s) used in attribute(s) 'resource' inside 'permission_specs'")
	}

	//resp.State.RemoveResource(ctx)
	diags = resp.State.Set(ctx, postPlan)
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

func (r *policyResource) validateSpecs(p *policyResourceModel) error {
	if p.ServiceType.ValueString() == service_type.REDIS {
		if len(p.PermissionSpecs) > 1 {
			return fmt.Errorf("for service %q, attribute 'permission_specs' must contain only 1 element", p.ServiceType.ValueString())
		}
		for _, specModel := range p.PermissionSpecs {
			if specModel.Permission.IsNull() || specModel.Permission.IsUnknown() {
				return fmt.Errorf("for service %q, attribute 'permission_specs[*].permission' must be specified", p.ServiceType.ValueString())
			}
		}
		return nil
	}
	for i, spec := range p.PermissionSpecs {
		if !(spec.Permission.IsNull() || spec.Permission.IsUnknown()) && spec.Role != spec.Permission {
			return fmt.Errorf("for service %q, attribute 'permission' (when passed) must match with attribute 'role' in 'permission_specs[%d]' ", p.ServiceType.ValueString(), i)
		}
	}
	return nil
}

func (r *policyResource) convertFromSpecsTfModel(specs *[]PermissionSpecModel) *[]customer_metadata.PermissionSpecRequest {
	specModels := make([]customer_metadata.PermissionSpecRequest, len(*specs))

	for i, specModel := range *specs {
		specModels[i] = customer_metadata.PermissionSpecRequest{
			Role:        specModel.Role.ValueString(),
			Resource:    specModel.Resource.ValueString(),
			Permissions: []string{specModel.Role.ValueString()}, // usually same as role, but for REDIS, permission will come due to validation, and will be set below
		}
		if !(specModel.Permission.IsNull() || specModel.Permission.IsUnknown()) {
			specModels[i].Permissions = []string{specModel.Permission.ValueString()}
		}
	}
	sort.Slice(specModels, func(i, j int) bool { return specModels[i].Role < specModels[j].Role })
	return &specModels
}

func saveFromPolicyResponse(ctx *context.Context, diagnostics *diag.Diagnostics, state *policyResourceModel, policy *model.Policy) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan", map[string]interface{}{"policy": *policy})
	tfPermissionSpecModel, diags := convertFromPermissionSpecDto(policy.PermissionsSpec)
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
	state.Updating = types.BoolValue(policy.Updating)
	resourceIds, diags := types.SetValueFrom(*ctx, types.StringType, policy.ResourceIds)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.ResourceIds = resourceIds

	return 0
}

func convertFromPermissionSpecDto(roles []model.PermissionsSpec) ([]PermissionSpecModel, diag.Diagnostics) {
	tfPermissionSpecModels := make([]PermissionSpecModel, len(roles))
	for i, permissionSpec := range roles {
		tfPermissionSpecModels[i] = PermissionSpecModel{
			Role:       types.StringValue(permissionSpec.Role),
			Resource:   types.StringValue(permissionSpec.Resource),
			Permission: types.StringValue(permissionSpec.Permissions[0].Name),
		}
	}
	return tfPermissionSpecModels, nil
}
