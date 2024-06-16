package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/policy_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
	customer_metadata "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
	"time"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &clusterNetworkPoliciesAssociationResource{}
	_ resource.ResourceWithConfigure   = &clusterNetworkPoliciesAssociationResource{}
	_ resource.ResourceWithImportState = &clusterNetworkPoliciesAssociationResource{}
)

func NewClusterNetworkPoliciesAssociationResource() resource.Resource {
	return &clusterNetworkPoliciesAssociationResource{}
}

type clusterNetworkPoliciesAssociationResource struct {
	client *tdh.Client
}

type clusterNetworkPoliciesAssociationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	PolicyIds []string     `tfsdk:"policy_ids"`
}

func (r *clusterNetworkPoliciesAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_network_policies_association"
}

func (r *clusterNetworkPoliciesAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *clusterNetworkPoliciesAssociationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "INIT__Schema")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Represents the association between a service instance/cluster and network policies.\nNOTE: Make sure to first import the existing associations that may have been created during cluster creation, since this is an overwrite operation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the cluster.",
				Required:    true,
				Computed:    false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_ids": schema.SetAttribute{
				MarkdownDescription: "IDs of the network policies to associate with the cluster.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}

	tflog.Info(ctx, "END__Schema")
}

func (r *clusterNetworkPoliciesAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__Create")
	// Retrieve values from plan
	var plan clusterNetworkPoliciesAssociationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	updateRequest := controller.ClusterNetworkPoliciesUpdateRequest{
		NetworkPolicyIds: plan.PolicyIds,
	}
	//plan.PolicyIds.ElementsAs(ctx, &updateRequest.NetworkPolicyIds, true)
	response, err := r.client.Controller.UpdateClusterNetworkPolicies(plan.ID.ValueString(), &updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating cluster network policies association",
			"Could not create association, unexpected error: "+err.Error(),
		)
		return
	}
	// this operation usually happens instantly
	time.Sleep(2 * time.Second)
	err = utils.WaitForTask(r.client, response.TaskId)
	if err != nil {
		resp.Diagnostics.AddError("Creating cluster network policies association",
			"Task responsible for this operation failed, error: "+err.Error(),
		)
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

func (r *clusterNetworkPoliciesAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state clusterNetworkPoliciesAssociationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed cluster value from TDH
	policies, err := r.client.CustomerMetadata.GetPolicies(&customer_metadata.PoliciesQuery{
		Type:       policy_type.NETWORK,
		ResourceId: state.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading Cluster Network Policies",
			fmt.Sprintf("Could not read TDH policies for cluster [%s] : %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	state.PolicyIds = []string{}
	for _, item := range *policies.Get() {
		state.PolicyIds = append(state.PolicyIds, item.ID)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Read")
}

func (r *clusterNetworkPoliciesAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")
	// Retrieve values from plan
	var plan clusterNetworkPoliciesAssociationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	updateRequest := controller.ClusterNetworkPoliciesUpdateRequest{
		NetworkPolicyIds: plan.PolicyIds,
	}
	response, err := r.client.Controller.UpdateClusterNetworkPolicies(plan.ID.ValueString(), &updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating cluster network policies association",
			"Could not update association, unexpected error: "+err.Error(),
		)
		return
	}
	// this operation usually happens instantly
	time.Sleep(2 * time.Second)
	err = utils.WaitForTask(r.client, response.TaskId)
	if err != nil {
		resp.Diagnostics.AddError("Updating cluster network policies association",
			"Task responsible for this operation failed, error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "END__Update")
}

func (r *clusterNetworkPoliciesAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "INIT__Delete")
	resp.Diagnostics.AddError(
		"Operation not valid",
		"Cluster must have at least one network policy attached. If you wish to clean it from state, please use \"state rm\" command on this resource.",
	)
	return
}

func (r *clusterNetworkPoliciesAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
