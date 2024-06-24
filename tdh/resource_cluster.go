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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/service_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
	upgrade_service "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/upgrade-service"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/validators"
	"regexp"
	"strconv"
	"strings"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &clusterResource{}
	_ resource.ResourceWithConfigure   = &clusterResource{}
	_ resource.ResourceWithImportState = &clusterResource{}
)

func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

type clusterResource struct {
	client *tdh.Client
}

// clusterResourceModel maps the resource schema data.
type clusterResourceModel struct {
	ID                types.String          `tfsdk:"id"`
	OrgId             types.String          `tfsdk:"org_id"`
	Name              types.String          `tfsdk:"name"`
	ServiceType       types.String          `tfsdk:"service_type"`
	Provider          types.String          `tfsdk:"provider_type"`
	InstanceSize      types.String          `tfsdk:"instance_size"`
	Region            types.String          `tfsdk:"region"`
	Tags              types.Set             `tfsdk:"tags"`
	NetworkPolicyIds  types.Set             `tfsdk:"network_policy_ids"`
	Dedicated         types.Bool            `tfsdk:"dedicated"`
	Shared            types.Bool            `tfsdk:"shared"`
	Status            types.String          `tfsdk:"status"`
	DataPlaneId       types.String          `tfsdk:"data_plane_id"`
	LastUpdated       types.String          `tfsdk:"last_updated"`
	Created           types.String          `tfsdk:"created"`
	Metadata          types.Object          `tfsdk:"metadata"`
	Version           types.String          `tfsdk:"version"`
	StoragePolicyName types.String          `tfsdk:"storage_policy_name"`
	ClusterMetadata   *clusterMetadataModel `tfsdk:"cluster_metadata"`
	Upgrade           *upgradeMetadata      `tfsdk:"upgrade"`
}

// clusterMetadataModel maps order item data.
type clusterMetadataModel struct {
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Database      types.String `tfsdk:"database"`
	Extensions    types.Set    `tfsdk:"extensions"`
	ObjectStoreId types.String `tfsdk:"object_storage_id"`
}

type MetadataModel struct {
	ManagerUri       types.String `tfsdk:"manager_uri"`
	ConnectionUri    types.String `tfsdk:"connection_uri"`
	MetricsEndpoints types.Set    `tfsdk:"metrics_endpoints"`
}

type upgradeMetadata struct {
	OmitBackup types.Bool `tfsdk:"omit_backup"`
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *clusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "INIT__Schema")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Represents a service instance or cluster. Some attributes are used only once for creation, they are: `dedicated`, `network_policy_ids`, `cluster_metadata`.\n" +
			"Changing only `tags` is supported at the moment. If you wish to update network policies associated with it, please refer resource: " +
			"`tdh_cluster_network_policies_association`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Description: "ID of the Org which owns the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
				},
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of TDH Cluster to be created. Supported values: %s.\n"+
					"Default is `POSTGRES`.", supportedServiceTypesMarkdown()),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(service_type.POSTGRES),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provider_type": schema.StringAttribute{
				MarkdownDescription: "Short-code of provider to use for data-plane. Ex: `tkgs`, `tkgm` . Complete list can be seen using datasource `tdh_provider_types`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"instance_size": schema.StringAttribute{
				MarkdownDescription: "Size of instance. Supported values: `XX-SMALL`, `X-SMALL`, `SMALL`, `LARGE`, `XX-LARGE`." +
					"\nPlease make use of datasource `tdh_network_ports` to decide on a size based on resources it requires." +
					"\n`SMALL-LITE` instance size is applicable only for 'POSTGRES' service type",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Region of data plane. Available values can be seen using datasource `tdh_regions`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
				},
			},
			"dedicated": schema.BoolAttribute{
				Description: "If present and set to `true`, the cluster will get deployed on a dedicated data-plane in current Org.",
				Optional:    true,
				Computed:    false,
			},
			"shared": schema.BoolAttribute{
				Description: "If present and set to `true`, the cluster will get deployed on a shared data-plane in current Org.",
				Optional:    true,
				Computed:    false,
			},
			"tags": schema.SetAttribute{
				Description: "Set of tags or labels to categorise the cluster.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"network_policy_ids": schema.SetAttribute{
				Description: "IDs of network policies to attach to the cluster.",
				Required:    true,
				Computed:    false,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						validators.UUIDValidator{},
					),
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the cluster.",
				Computed:    true,
			},
			"data_plane_id": schema.StringAttribute{
				Description: "ID of the data-plane where the cluster is running.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validators.UUIDValidator{},
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Time when the cluster was last modified.",
				Computed:    true,
			},
			"created": schema.StringAttribute{
				Description: "Creation time of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the cluster.\n" +
					"#### Notes:\n" +
					"- Changing version will result in cluster upgrade process. Use datasource `tdh_cluster_target_versions` to get next versions.\n" +
					"- To specify extra options for cluster upgrade, please make use of ['upgrade' attribute](#nestedatt--upgrade).",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"storage_policy_name": schema.StringAttribute{
				Description: "Name of the storage policy for the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validators.EmptyStringValidator{},
				},
			},
			"metadata": schema.SingleNestedAttribute{
				Description: "Additional info of the cluster.",
				CustomType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"cluster_name":   types.StringType,
						"manager_uri":    types.StringType,
						"connection_uri": types.StringType,
						"metrics_endpoints": types.SetType{
							ElemType: types.StringType,
						},
						"object_storage_id": types.StringType,
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"cluster_name": schema.StringAttribute{
						MarkdownDescription: "Name of the cluster.",
						Computed:            true,
					},
					"manager_uri": schema.StringAttribute{
						MarkdownDescription: "URI of the manager.",
						Computed:            true,
					},
					"connection_uri": schema.StringAttribute{
						MarkdownDescription: "Connection URI to the instance.",
						Computed:            true,
					},
					"metrics_endpoints": schema.SetAttribute{
						MarkdownDescription: "List of metrics endpoints exposed on the instance.",
						Computed:            true,
						ElementType:         types.StringType,
					},
					"object_storage_id": schema.StringAttribute{
						MarkdownDescription: "ID of the object storage for backup operations.",
						Computed:            true,
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"cluster_metadata": schema.SingleNestedAttribute{
				MarkdownDescription: fmt.Sprintf("Additional info for the cluster. Required for services: %s.", supportedDataServiceTypesMarkdown()),
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "Username for the cluster.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(3, 32),
							stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z][a-z0-9_]*$`), "must start with an alphabet & may contain only lowercase alphabets, numbers or underscores"),
						},
					},
					"password": schema.StringAttribute{
						Description: "Password for the cluster.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(8, 24),
							validators.PasswordValidator{},
						},
					},
					"database": schema.StringAttribute{
						MarkdownDescription: "Database name in the cluster. **Required for services:** `POSTGRES` & `MYSQL`.",
						Required:            false,
						Optional:            true,
					},
					"extensions": schema.SetAttribute{
						MarkdownDescription: "Set of extensions to be enabled on the cluster *(Specific to service: `POSTGRES`)*. Available values can be fetched using datasource `tdh_service_extensions`.",
						Optional:            true,
						ElementType:         types.StringType,
					},
					"object_storage_id": schema.StringAttribute{
						MarkdownDescription: "ID of the object storage for backup operations. Can be fetched using datasource `tdh_object_storages`.",
						Optional:            true,
						Validators: []validator.String{
							validators.UUIDValidator{},
						},
					},
				},
			},
			"upgrade": schema.SingleNestedAttribute{
				Description: "Use this for specifying extra options for upgrading cluster version.",
				Required:    false,
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"omit_backup": schema.BoolAttribute{
						Description: "Whether to take backup before upgrade process. (default is `false`)",
						Optional:    true,
					},
				},
			},
		},
	}

	tflog.Info(ctx, "END__Schema")
}

// Create a new resource
func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__Create")
	// Retrieve values from plan
	var plan clusterResourceModel

	tflog.Info(ctx, "INIT__Fetching plan")
	diags := req.Plan.Get(ctx, &plan)
	tflog.Info(ctx, "INIT__Fetched plan")

	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	if r.validateInputs(&ctx, &resp.Diagnostics, &plan); resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "INIT__Creating req body")

	// Generate API request body from plan
	clusterRequest := controller.ClusterCreateRequest{
		Name:              plan.Name.ValueString(),
		ServiceType:       plan.ServiceType.ValueString(),
		InstanceSize:      plan.InstanceSize.ValueString(),
		Provider:          plan.Provider.ValueString(),
		Region:            plan.Region.ValueString(),
		Dedicated:         plan.Dedicated.ValueBool(),
		Shared:            plan.Shared.ValueBool(),
		DataPlaneId:       plan.DataPlaneId.ValueString(),
		Version:           plan.Version.ValueString(),
		StoragePolicyName: plan.StoragePolicyName.ValueString(),
	}
	if plan.ServiceType.ValueString() != service_type.RABBITMQ {
		clusterRequest.ClusterMetadata = controller.ClusterMetadata{
			Username:      plan.ClusterMetadata.Username.ValueString(),
			Password:      plan.ClusterMetadata.Password.ValueString(),
			Database:      plan.ClusterMetadata.Database.ValueString(),
			ObjectStoreId: plan.ClusterMetadata.ObjectStoreId.ValueString(),
		}
	}

	plan.ClusterMetadata.Extensions.ElementsAs(ctx, &clusterRequest.ClusterMetadata.Extensions, true)
	tflog.Info(ctx, "INIT__Created req body")
	tflog.Info(ctx, "Creating cluster", map[string]interface{}{
		"cluster_request": clusterRequest,
	})

	plan.Tags.ElementsAs(ctx, &clusterRequest.Tags, true)
	plan.NetworkPolicyIds.ElementsAs(ctx, &clusterRequest.NetworkPolicyIds, true)

	tflog.Info(ctx, "INIT__Submitting request")

	response, err := r.client.Controller.CreateCluster(&clusterRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Submitting request to create cluster",
			"Could not create cluster, unexpected error: "+err.Error(),
		)
		return
	}
	if err = utils.WaitForTask(r.client, response.TaskId); err != nil {
		resp.Diagnostics.AddError("Error in creating cluster",
			"Task responsible for this operation failed, error: "+err.Error(),
		)
		return
	}
	tflog.Info(ctx, "INIT__Fetching clusters")
	clusters, err := r.client.Controller.GetClusters(&controller.ClustersQuery{
		ServiceType:   clusterRequest.ServiceType,
		Name:          clusterRequest.Name,
		FullNameMatch: true,
		PageQuery: model.PageQuery{
			Size: 1,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Fetching clusters",
			"Could not fetch clusters by name, unexpected error: "+err.Error(),
		)
		return
	}

	if len(*clusters.Get()) <= 0 {
		resp.Diagnostics.AddError("Fetching clusters",
			"Unable to fetch the created cluster",
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	createdCluster := &(*clusters.Get())[0]
	tflog.Info(ctx, "INIT__Saving Response")
	if r.saveFromResponse(&ctx, &resp.Diagnostics, &plan, createdCluster) != 0 {
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

// Read resource information
func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "INIT_Read Fetching Cluster from API")
	// Get refreshed cluster value from TDH
	cluster, err := r.client.Controller.GetCluster(state.ID.ValueString())
	tflog.Debug(ctx, "INIT__Read fetched cluster", map[string]interface{}{"dto": cluster})
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading TDH Cluster",
			"Could not read TDH cluster ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "INIT__Read converting response")
	// Overwrite items with refreshed state
	if r.saveFromResponse(&ctx, &resp.Diagnostics, &state, cluster) != 0 {
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

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")

	// Retrieve values from plan
	var plan clusterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve current state
	var state clusterResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Detect version change
	if plan.Version != state.Version {
		tflog.Info(ctx, "Version change detected", map[string]interface{}{
			"old_version": state.Version.ValueString(),
			"new_version": plan.Version.ValueString(),
		})
		// Generate API request to update the version
		versionUpdateRequest := upgrade_service.UpdateClusterVersionRequest{
			Id:            state.ID.ValueString(),
			TargetVersion: plan.Version.ValueString(),
			RequestType:   "SERVICE",
		}

		if plan.Upgrade != nil {
			versionUpdateRequest.Metadata.OmitBackup = strconv.FormatBool(plan.Upgrade.OmitBackup.ValueBool())
		}

		// Call the API to update the version
		response, err := r.client.UpgradeService.UpdateClusterVersion(&versionUpdateRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Updating Cluster Version",
				"Could not update cluster version, unexpected error: "+err.Error(),
			)
			return
		}
		if err = utils.WaitForTask(r.client, response.TaskId); err != nil {
			resp.Diagnostics.AddError("Updating Cluster Version",
				"Operation error: "+err.Error(),
			)
			return
		}
		resp.State.RemoveResource(ctx)
	}

	// Generate API request body from plan
	var updateRequest controller.ClusterUpdateRequest
	plan.Tags.ElementsAs(ctx, &updateRequest.Tags, true)

	// Update existing cluster
	cluster, err := r.client.Controller.UpdateCluster(plan.ID.ValueString(), &updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating TDH Cluster",
			"Could not update cluster, unexpected error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	if r.saveFromResponse(&ctx, &resp.Diagnostics, &plan, cluster) != 0 {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Update")
}

func (r *clusterResource) Delete(ctx context.Context, request resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "INIT__Delete")
	// Get current state
	var state clusterResourceModel
	diags := request.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Submit request to delete TDH Cluster
	response, err := r.client.Controller.DeleteCluster(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Deleting TDH Cluster",
			"Could not delete TDH cluster by ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	if err = utils.WaitForTask(r.client, response.TaskId); err != nil {
		resp.Diagnostics.AddError("Deleting TDH Cluster",
			"Task responsible for this operation failed, error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "END__Delete")
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *clusterResource) saveFromResponse(ctx *context.Context, diagnostics *diag.Diagnostics, state *clusterResourceModel, cluster *model.Cluster) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan")
	state.ID = types.StringValue(cluster.ID)
	state.Name = types.StringValue(cluster.Name)
	state.ServiceType = types.StringValue(cluster.ServiceType)
	state.Provider = types.StringValue(cluster.Provider)
	state.InstanceSize = types.StringValue(cluster.InstanceSize)
	state.Region = types.StringValue(cluster.Region)
	state.Status = types.StringValue(cluster.Status)
	state.OrgId = types.StringValue(cluster.OrgId)
	state.DataPlaneId = types.StringValue(cluster.DataPlaneId)
	state.LastUpdated = types.StringValue(cluster.LastUpdated)
	state.Created = types.StringValue(cluster.Created)
	state.StoragePolicyName = types.StringValue(cluster.StoragePolicyName)
	state.Version = types.StringValue(cluster.Version)
	tflog.Info(*ctx, "trying to save mdsMetadata", map[string]interface{}{
		"obj": cluster.Metadata,
	})

	metadataObject, diags := types.ObjectValueFrom(*ctx, state.Metadata.AttributeTypes(*ctx), cluster.Metadata)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.Metadata = metadataObject
	list, diags := types.SetValueFrom(*ctx, types.StringType, cluster.Tags)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.Tags = list
	return 0
}

func (r *clusterResource) validateInputs(ctx *context.Context, diags *diag.Diagnostics, tfPlan *clusterResourceModel) {
	tflog.Info(*ctx, "validating inputs")
	if tfPlan.ServiceType.ValueString() != service_type.POSTGRES && tfPlan.InstanceSize.ValueString() == "SMALL-LITE" {
		diags.AddAttributeError(path.Root("instance_size"),
			"Invalid input", fmt.Sprintf("Instance Size \"%s\" is not available for Service \"%s\"", tfPlan.InstanceSize.ValueString(), tfPlan.ServiceType.ValueString()))
		return
	}
	if tfPlan.ServiceType.ValueString() == service_type.RABBITMQ {
		return
	}
	if tfPlan.ClusterMetadata == nil {
		diags.AddAttributeError(path.Root("cluster_metadata"),
			"Invalid input", fmt.Sprintf("Service \"%s\" requires this attribute.", tfPlan.ServiceType.ValueString()))
		return
	}

	if tfPlan.ServiceType.ValueString() != service_type.REDIS {
		if tfPlan.ClusterMetadata.Database.IsNull() || tfPlan.ClusterMetadata.Database.ValueString() == "" || strings.TrimSpace(tfPlan.ClusterMetadata.Database.ValueString()) == "" {
			diags.AddAttributeError(path.Root("cluster_metadata").AtName("database"),
				"Invalid input", fmt.Sprintf("Service \"%s\" requires database attribute.", tfPlan.ServiceType.ValueString()))
			return
		}
		const pattern = `^[a-z][a-z0-9_]*$`
		regex := regexp.MustCompile(pattern)

		// Check if the value matches the pattern
		if !regex.MatchString(tfPlan.ClusterMetadata.Database.ValueString()) {
			diags.AddError("Invalid input", "Attribute 'database'  should start with an alphabet & may contain only lowercase alphabets, numbers or underscores")

		}
	}

}
