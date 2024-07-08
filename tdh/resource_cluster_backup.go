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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/service_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/validators"
	"time"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &clusterBackupResource{}
	_ resource.ResourceWithConfigure   = &clusterBackupResource{}
	_ resource.ResourceWithImportState = &clusterBackupResource{}
)

func NewClusterBackupResource() resource.Resource {
	return &clusterBackupResource{}
}

type clusterBackupResource struct {
	client *tdh.Client
}

// clusterBackupResourceModel maps the resource schema data.
type clusterBackupResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	ClusterName       types.String `tfsdk:"cluster_name"`
	Name              types.String `tfsdk:"name"`
	GeneratedName     types.String `tfsdk:"generated_name"`
	Description       types.String `tfsdk:"description"`
	ServiceType       types.String `tfsdk:"service_type"`
	BackupTriggerType types.String `tfsdk:"backup_trigger_type"`
	DataPlaneId       types.String `tfsdk:"data_plane_id"`
	ClusterVersion    types.String `tfsdk:"cluster_version"`
	OrgId             types.String `tfsdk:"org_id"`
	Provider          types.String `tfsdk:"provider_name"`
	Region            types.String `tfsdk:"region"`
	Size              types.String `tfsdk:"size"`
	Status            types.String `tfsdk:"status"`
	TimeStarted       types.String `tfsdk:"time_started"`
	TimeCompleted     types.String `tfsdk:"time_completed"`
	Restore           types.Object `tfsdk:"restore"`
	Metadata          types.Object `tfsdk:"metadata"`
}

type RestoreInfoModel struct {
	ClusterName      types.String `tfsdk:"cluster_name"`
	StoragePolicy    types.String `tfsdk:"storage_policy"`
	NetworkPolicyIds []string     `tfsdk:"network_policy_ids"`
	Tags             []string     `tfsdk:"tags"`
}

func (r *clusterBackupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_backup"
}

func (r *clusterBackupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *clusterBackupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "INIT__Schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "This is used to create backup (and restore a backup) of a database service cluster like `POSTGRES`, `MYSQL`, `REDIS`.\n" +
			"**Note:** To restore a backup, either create a backup or import by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the backup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "ID of the cluster to take backup of.",
				Required:    true,
				Validators: []validator.String{
					validators.UUIDValidator{},
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the backup.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description for the backup.",
				Optional:    true,
			},
			"cluster_name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_version": schema.StringAttribute{
				Description: "Version of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_type": schema.StringAttribute{
				Description: "Service Type of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"generated_name": schema.StringAttribute{
				Description: "Name that was automatically generated for this backup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time_started": schema.StringAttribute{
				Description: "Time when the backup was initiated.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time_completed": schema.StringAttribute{
				Description: "Time when the backup was completed.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the backup.",
				Computed:    true,
			},
			"size": schema.StringAttribute{
				Description: "Size of the cluster.",
				Computed:    true,
			},
			"backup_trigger_type": schema.StringAttribute{
				Description: "The type of trigger for the backup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provider_name": schema.StringAttribute{
				Description: "The provider of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Description: "ID of the org this backup belongs to.",
				Computed:    true,
				Validators: []validator.String{
					validators.UUIDValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Description: "The region of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_plane_id": schema.StringAttribute{
				Description: "The ID of the data plane.",
				Computed:    true,
				Validators: []validator.String{
					validators.UUIDValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.SingleNestedAttribute{
				Description: "The metadata of the backup.",
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				CustomType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"cluster_name":    types.StringType,
						"cluster_size":    types.StringType,
						"backup_location": types.StringType,
						"databases": types.SetType{
							ElemType: types.StringType,
						},
						"extensions": types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":    types.StringType,
									"version": types.StringType,
								},
							},
						},
					},
				},
				Attributes: map[string]schema.Attribute{
					"cluster_name": schema.StringAttribute{
						Description: "Name of the cluster.",
						Computed:    true,
					},
					"cluster_size": schema.StringAttribute{
						Description: "Size of the Instance Type.",
						Computed:    true,
					},
					"backup_location": schema.StringAttribute{
						Description: "Backup Location",
						Computed:    true,
					},
					"databases": schema.SetAttribute{
						Description: "List of databases part of backup.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"extensions": schema.SetNestedAttribute{
						MarkdownDescription: "List of extensions part of backup. *(Specific to service `POSTGRES`)*",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Name of the extension.",
									Computed:    true,
								},
								"version": schema.StringAttribute{
									Description: "Version of the extension.",
									Computed:    true,
								},
							},
						},
					},
				},
			},
			"restore": schema.SingleNestedAttribute{
				MarkdownDescription: "Define this block to restore a pre-created backup.\n" +
					"## Notes\n" +
					"- Just declare it as empty block in case of `REDIS` cluster backup since in case of Redis, restore happens on same cluster i.e. the cluster has to be present and there will be some downtime.\n" +
					"- Backup creation and restore won't happen in same operation, so first backup has to be created or imported, then next 'apply' will trigger restore.",
				Required: false,
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				CustomType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"cluster_name":   types.StringType,
						"storage_policy": types.StringType,
						"network_policy_ids": types.SetType{
							ElemType: types.StringType,
						},
						"tags": types.SetType{
							ElemType: types.StringType,
						},
					},
				},
				Attributes: map[string]schema.Attribute{
					"cluster_name": schema.StringAttribute{
						Description: "Name of the target instance.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
							stringvalidator.AlsoRequires(path.Expressions{
								path.Root("restore").AtName("storage_policy").Expression(),
								path.Root("restore").AtName("network_policy_ids").Expression(),
							}...),
						},
					},
					"storage_policy": schema.StringAttribute{
						Description: "Name of the storage policy.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
							stringvalidator.AlsoRequires(path.Expressions{
								path.Root("restore").AtName("cluster_name").Expression(),
								path.Root("restore").AtName("network_policy_ids").Expression(),
							}...),
						},
					},
					"network_policy_ids": schema.SetAttribute{
						Description: "List of network policy IDs for network configuration on the instance.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.AlsoRequires(path.Expressions{
								path.Root("restore").AtName("cluster_name").Expression(),
								path.Root("restore").AtName("storage_policy").Expression(),
							}...),
						},
					},
					"tags": schema.SetAttribute{
						Description: "List of tags to set on the instance.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.Set{
							setvalidator.AlsoRequires(path.Expressions{
								path.Root("restore").AtName("cluster_name").Expression(),
								path.Root("restore").AtName("storage_policy").Expression(),
								path.Root("restore").AtName("network_policy_ids").Expression(),
							}...),
						},
					},
				},
			},
		},
	}

}

// Create a new resource
func (r *clusterBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__CreateBackup")
	// Retrieve values from plan
	var plan clusterBackupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	if !(plan.Restore.IsNull() || plan.Restore.IsUnknown()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("restore"),
			"Invalid input",
			"attribute 'restore' is only required after creating/importing a backup",
		)
		return
	}
	// Generate API request body from plan
	request := controller.BackupCreateRequest{
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		BackupSchedule: "ON_DEMAND",
		BackupType:     "FULL",
	}

	tflog.Info(ctx, "req body", map[string]interface{}{"create-backup-request": request})

	response, err := r.client.Controller.CreateClusterBackup(plan.ClusterID.ValueString(), &request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating cluster backup",
			"Could not create Cluster Backup, unexpected error: "+err.Error(),
		)
		return
	}
	err = utils.WaitForTask(r.client, response.TaskId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating cluster backup",
			"Task responsible for this operation failed, error: "+err.Error(),
		)
	}

	tflog.Info(ctx, "INIT__Fetching Cluster Backup")

	backups, err := r.client.Controller.GetClusterBackups(&controller.BackupsQuery{
		Name:      request.Name,
		ClusterId: plan.ClusterID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Fetching cluster backups",
			"Could not fetch cluster backups by name, unexpected error: "+err.Error(),
		)
		return
	}

	if backups.Page.TotalElements == 0 {
		resp.Diagnostics.AddError("Fetching Cluster Backup",
			fmt.Sprintf("Could not find any cluster backup by backup name [%s], server error must have occurred while creating user.", plan.Name.ValueString()),
		)
		return
	}
	createdBackup := &(*backups.Get())[0]
	if err = r.waitForBackupStatus(ctx, createdBackup); err != nil {
		resp.Diagnostics.AddError("Verifying cluster backup",
			"Could not verify backup status, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	if r.saveBackupFromResponse(&ctx, &resp.Diagnostics, &plan, createdBackup) != 0 {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Create")
}

func (r *clusterBackupResource) waitForBackupStatus(ctx context.Context, createdBackup *model.ClusterBackup) error {
	bkpChan := make(chan error)
	go func(ch chan error) {
		triesLeft := 3
		ticker := time.NewTicker(15 * time.Second)
		for triesLeft != 0 {
			select {
			case <-ctx.Done():
				ch <- fmt.Errorf("polling task has been cancelled")
				break
			case <-ticker.C:
				bkp, err := r.client.Controller.GetBackup(createdBackup.Id)
				if err != nil {
					ch <- fmt.Errorf("backup progress could not be checked due to error: %v", err.Error())
					break
				}
				if bkp.Status == "Completed" || bkp.Status == "Succeeded" {
					ch <- nil
					break
				}
				triesLeft--
			}
		}
	}(bkpChan)

	return <-bkpChan
}

func (r *clusterBackupResource) Update(ctx context.Context, request resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")
	// Get current plan
	var state, plan clusterBackupResourceModel
	diags := request.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	diags = request.State.Get(ctx, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	if !plan.Restore.IsNull() {
		tflog.Info(ctx, "Considering it as restore action")
		if r.validateRestoreInputs(&ctx, &resp.Diagnostics, &state, &plan); resp.Diagnostics.HasError() {
			return
		}
		r.triggerRestoreAndWait(&ctx, &resp.Diagnostics, &state, &plan)
		state.Restore, diags = types.ObjectValueFrom(ctx, state.Restore.AttributeTypes(ctx), plan.Restore)
		diags = resp.State.Set(ctx, &state)
	}
	tflog.Info(ctx, "END__Update")
}

func (r *clusterBackupResource) Delete(ctx context.Context, request resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "INIT__Delete")
	// Get current state
	var state clusterBackupResourceModel
	diags := request.State.Get(ctx, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	// Submit request to delete Backup
	response, err := r.client.Controller.DeleteClusterBackup(state.ID.ValueString())
	if err != nil {
		apiErr := core.ApiError{}
		errors.As(err, &apiErr)
		if apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting cluster backup",
			"Could not delete cluster backup by ID "+state.ClusterID.ValueString()+": "+err.Error(),
		)
		return
	}
	err = utils.WaitForTask(r.client, response.TaskId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Deleting cluster backup",
			"Task responsible for this operation failed, error: "+err.Error(),
		)
	}

	tflog.Info(ctx, "END__Delete")
}

func (r *clusterBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state clusterBackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the backup from the API
	backup, err := r.client.Controller.GetBackup(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading TDH cluster backup",
			"Could not read TDH cluster backup "+state.ID.ValueString()+": "+err.Error(),
		)
		resp.Diagnostics.AddError("Reading TDH cluster backup", err.Error())
		return
	}

	// Map response to the state
	if r.saveBackupFromResponse(&ctx, &resp.Diagnostics, &state, backup) != 0 {
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

func (r *clusterBackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *clusterBackupResource) saveBackupFromResponse(ctx *context.Context, diags *diag.Diagnostics, state *clusterBackupResourceModel, response *model.ClusterBackup) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan")
	if !(state.ID.IsNull() || state.ID.IsUnknown()) {
		state.Restore = types.ObjectNull(state.Restore.AttributeTypes(*ctx))
	}
	state.ID = types.StringValue(response.Id)
	state.ClusterID = types.StringValue(response.ClusterId)
	state.Name = types.StringValue(response.Name)
	state.GeneratedName = types.StringValue(response.GeneratedName)
	state.ServiceType = types.StringValue(response.ServiceType)
	state.ClusterName = types.StringValue(response.ClusterName)
	state.ClusterVersion = types.StringValue(response.ClusterVersion)
	state.BackupTriggerType = types.StringValue(response.BackupTriggerType)
	state.DataPlaneId = types.StringValue(response.DataPlaneId)
	state.OrgId = types.StringValue(response.OrgId)
	state.Region = types.StringValue(response.Region)
	state.Status = types.StringValue(response.Status)
	state.Size = types.StringValue(response.Size)
	state.Provider = types.StringValue(response.Provider)
	state.TimeStarted = types.StringValue(response.TimeStarted)
	state.TimeCompleted = types.StringValue(response.TimeCompleted)
	metadataModel := BackupMetadata{
		ClusterName:    types.StringValue(response.Metadata.ClusterName),
		ClusterSize:    types.StringValue(response.Metadata.ClusterSize),
		BackupLocation: types.StringValue(response.Metadata.BackupLocation),
		Databases:      response.Metadata.Databases,
	}
	for _, ext := range response.Metadata.PostgresExtensions {
		extModel := BackupMetadataExtension{
			Name:    types.StringValue(ext.Name),
			Version: types.StringValue(ext.Version),
		}
		metadataModel.Extensions = append(metadataModel.Extensions, extModel)
	}
	metadataObject, dgs := types.ObjectValueFrom(*ctx, state.Metadata.AttributeTypes(*ctx), metadataModel)
	if diags.Append(dgs...); diags.HasError() {
		return 1
	}
	state.Metadata = metadataObject
	return 0
}

func (r *clusterBackupResource) validateRestoreInputs(ctx *context.Context, diags *diag.Diagnostics, state *clusterBackupResourceModel, plan *clusterBackupResourceModel) {
	tflog.Info(*ctx, "validating restore inputs", map[string]interface{}{
		"plan": plan,
	})
	defer func() {
		tflog.Info(*ctx, "validated restore inputs")
	}()
	if state.ServiceType.ValueString() == service_type.REDIS {
		return
	}
	var restoreInfo RestoreInfoModel
	plan.Restore.As(*ctx, &restoreInfo, basetypes.ObjectAsOptions{})
	if restoreInfo.ClusterName.IsNull() {
		diags.AddError("Invalid input", "Value for 'cluster_name' is required for restore.")
	}
	if restoreInfo.StoragePolicy.IsNull() {
		diags.AddError("Invalid input", "Value for 'storage_policy' is required for restore.")
	}
	if restoreInfo.NetworkPolicyIds == nil {
		diags.AddError("Invalid input", "Value for 'network_policy_ids' is required for restore.")
	}
}

func (r *clusterBackupResource) triggerRestoreAndWait(ctx *context.Context, diags *diag.Diagnostics, state *clusterBackupResourceModel, plan *clusterBackupResourceModel) {
	tflog.Info(*ctx, "initiating restore")
	defer func() {
		tflog.Info(*ctx, "exiting restore")
	}()
	var restoreInfo RestoreInfoModel
	plan.Restore.As(*ctx, &restoreInfo, basetypes.ObjectAsOptions{})
	var metadataModel BackupMetadata
	state.Metadata.As(*ctx, &metadataModel, basetypes.ObjectAsOptions{})
	request := controller.ClusterCreateRequest{
		Name:              restoreInfo.ClusterName.ValueString(),
		StoragePolicyName: restoreInfo.StoragePolicy.ValueString(),
		NetworkPolicyIds:  restoreInfo.NetworkPolicyIds,
		Region:            state.Region.ValueString(),
		Provider:          state.Provider.ValueString(),
		Version:           state.ClusterVersion.ValueString(),
		InstanceSize:      metadataModel.ClusterSize.ValueString(),
		ServiceType:       state.ServiceType.ValueString(),
		Tags:              restoreInfo.Tags,
	}
	if state.ServiceType.ValueString() == service_type.REDIS {
		request.Name = state.ClusterName.ValueString()
		request.StoragePolicyName = "dummy"
		request.NetworkPolicyIds = append(request.NetworkPolicyIds, "dummy")
	}
	tflog.Debug(*ctx, "initial req", map[string]interface{}{"req": request})
	var extNames []string
	for _, extension := range metadataModel.Extensions {
		extNames = append(extNames, extension.Name.ValueString())
	}
	clusterMetadata := controller.ClusterMetadata{
		Username:    "tdh_internal_user",
		Password:    "********",
		RestoreFrom: state.ID.ValueString(),
		Extensions:  []string{}, // sending exact extensions fails for some reason
	}
	if len(metadataModel.Databases) > 0 {
		clusterMetadata.Database = metadataModel.Databases[0]
	}
	request.ClusterMetadata = clusterMetadata
	tflog.Debug(*ctx, "req with metadata", map[string]interface{}{"req": request})
	dataPlane, err := r.client.InfraConnector.GetDataPlaneById(state.DataPlaneId.ValueString())
	if err != nil {
		diags.AddError("Restoring cluster backup", "Could not fetch required details: "+err.Error())
		return
	}
	tflog.Debug(*ctx, "fetched data plane", map[string]interface{}{"dp": dataPlane})
	request.Shared = dataPlane.Shared
	request.Dedicated = len(dataPlane.OrgId) != 0
	tflog.Debug(*ctx, "req after dp details", map[string]interface{}{"req": request})
	var api func(request *controller.ClusterCreateRequest) (*model.TaskResponse, error)
	if request.ServiceType == service_type.POSTGRES {
		api = r.client.Controller.CreateCluster
	} else {
		api = r.client.Controller.RestoreClusterBackup
	}
	response, err := api(&request)
	if err != nil {
		diags.AddError("Restoring cluster backup", "Got error while submitting request: "+err.Error())
		return
	}
	err = utils.WaitForTask(r.client, response.TaskId)
	if err != nil {
		diags.AddError("Restoring cluster backup",
			"Task responsible for this operation failed, error: "+err.Error())
		return
	}
	tflog.Debug(*ctx, "task successfully completed")
}
