package tdh

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"time"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &clusterBackupResource{}
	_ resource.ResourceWithConfigure = &clusterBackupResource{}
)

func NewclusterBackupResource() resource.Resource {
	return &clusterBackupResource{}
}

type clusterBackupResource struct {
	client *tdh.Client
}

// clusterBackupResourceModel maps the resource schema data.
type ClusterBackupResourceModel struct {
	ClusterID   types.String `tfsdk:"cluster_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ServiceType types.String `tfsdk:"service_type"`
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
		Description: "Creation of luster Backup",
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Description: "ID of the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the backup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the backup",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description for the create backup",
				Optional:    true,
			},
			"service_type": schema.StringAttribute{
				Description: "Service Type for Backup",
				Optional:    true,
			},
		},
	}

}

// Create a new resource
func (r *clusterBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__CreateBackup")
	// Retrieve values from plan
	var plan ClusterBackupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	query := &controller.ClusterBackupQuery{
		ID: plan.ClusterID.ValueString(),
	}
	fmt.Println(query)
	// Generate API request body from plan
	request := controller.BackupCreateRequest{
		ClusterId:      plan.ClusterID.ValueString(),
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		BackupSchedule: "ON_DEMAND",
		BackupType:     "FULL",
		ServiceType:    plan.ServiceType.ValueString(),
	}

	tflog.Info(ctx, "req param", map[string]interface{}{"create-backup-request": request})

	response, err := r.client.Controller.CreateClusterBackup(&request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Submitting request",
			"Could not create Cluster Backup , unexpected error: "+err.Error(),
		)
		return
	}
	//Check if the response is nil or empty
	if response == nil {
		resp.Diagnostics.AddError(
			"Unexpected response",
			"The API response for creating a cluster backup is empty",
		)
	}

	tflog.Info(ctx, "INIT__Fetching Cluster Backup")
	//var backups model.Paged[model.ClusterBackup]

	backups, err := r.client.Controller.GetBackups(&controller.BackupQuery{
		ServiceType: request.ServiceType,
		Name:        request.Name,
	})
	if err != nil {
		resp.Diagnostics.AddError("Fetching cluster backups",
			"Could not fetch cluster bakups by name, unexpected error: "+err.Error(),
		)
		return
	}

	if backups.Page.TotalElements == 0 {
		resp.Diagnostics.AddError("Fetching Cluster Backup",
			fmt.Sprintf("Could not find any cluster backup by backup name [%s], server error must have occurred while creating user.", plan.Name.ValueString()),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	createdBackup := &(*backups.Get())[0]

	// Map response body to schema and populate Computed attribute values
	if saveBackupFromResponse(&ctx, &plan, createdBackup) != 0 {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__CreateBackup")
}

func (r *clusterBackupResource) Update(ctx context.Context, request resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Delete")
	// Get current state
	var state ClusterBackupResourceModel
	diags := request.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Submit request to delete Backup
	err := r.client.Controller.DeleteClusterBackup(state.ClusterID.ValueString())
	if err != nil {
		apiErr := core.ApiError{}
		errors.As(err, &apiErr)
		if apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting object store",
			"Could not delete object store by ID "+state.ClusterID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "END__Delete")
}

func (r *clusterBackupResource) Delete(ctx context.Context, request resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "INIT__Delete")
	// Get current state
	var state ClusterBackupResourceModel
	diags := request.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Submit request to delete Backup
	err := r.client.Controller.DeleteClusterBackup(state.ID.ValueString())
	if err != nil {
		apiErr := core.ApiError{}
		errors.As(err, &apiErr)
		if apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting object store",
			"Could not delete object store by ID "+state.ClusterID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "END__Delete")
}

func (r *clusterBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state ClusterBackupResourceModel
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
	if saveBackupFromResponse(&ctx, &state, backup) != 0 {
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

func saveBackupFromResponse(ctx *context.Context, state *ClusterBackupResourceModel, response *model.ClusterBackup) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan")
	state.ClusterID = types.StringValue(response.ClusterId)
	state.ID = types.StringValue(response.Id)
	state.Name = types.StringValue(response.Name)
	state.ServiceType = types.StringValue(response.ServiceType)
	return 0
}

func (r *clusterBackupResource) pollTaskStatus(taskId string) error {
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
