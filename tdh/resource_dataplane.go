package tdh

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	infra_connector "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"net/http"
	"time"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dataPlaneResource{}
	_ resource.ResourceWithConfigure   = &dataPlaneResource{}
	_ resource.ResourceWithImportState = &dataPlaneResource{}
)

func NewDataPlaneResource() resource.Resource {
	return &dataPlaneResource{}
}

type dataPlaneResource struct {
	client *tdh.Client
}

type dataPlaneResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	AccountId             types.String `tfsdk:"account_id"`
	ProviderName          types.String `tfsdk:"provider_name"`
	DataPlaneReleaseId    types.String `tfsdk:"data_plane_release_id"`
	K8sClusterName        types.String `tfsdk:"k8s_cluster_name"`
	StorageClasses        types.Set    `tfsdk:"storage_classes"`
	BackupStorageCLass    types.String `tfsdk:"backup_storage_class"`
	Shared                types.Bool   `tfsdk:"shared"`
	OrgId                 types.String `tfsdk:"org_id"`
	Tags                  types.Set    `tfsdk:"tags"`
	AutoUpgrade           types.Bool   `tfsdk:"auto_upgrade"`
	Services              types.Set    `tfsdk:"services"`
	CpBootstrappedCluster types.Bool   `tfsdk:"cp_bootstrapped_cluster"`
	ConfigureCoreDns      types.Bool   `tfsdk:"configure_core_dns"`
}

func (r *dataPlaneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataplane"
}

func (r *dataPlaneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *dataPlaneResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "INIT__Schema")
	resp.Schema = schema.Schema{
		Description: "Represents a TDH Dataplane . Supported actions are Add and Delete ",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Auto-generated ID of the dataplane after creation, and can be used to import it from TDH to terraform state.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "Id of the selected Cloud Account",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the DataPlane",
				Required:    true,
			},
			"k8s_cluster_name": schema.StringAttribute{
				Description: "Name of the TKC",
				Required:    true,
			},
			"shared": schema.BoolAttribute{
				Description: "Shared Dataplane",
				Required:    true,
			},
			"configure_core_dns": schema.BoolAttribute{
				Description: "Turn on publishing the TDH's DNS to core DNS of the K8S cluster.This will enable communication between pods of the control plane and the data plane with control plane pods. Disable this only when the TDH base URL has a forwarder set in corporate DNS or for some specific use case.\nPlease note that TDH needs these communications between the pods to function.",
				Required:    true,
			},
			"auto_upgrade": schema.BoolAttribute{
				Description: "Auto Upgrade Dataplane",
				Required:    true,
			},
			"cp_bootstrapped_cluster": schema.BoolAttribute{
				Description: "Onboard Data Plane on TDH Control Plane",
				Required:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Organization Id",
				Required:    true,
				Default:     nil,
			},
			"provider_name": schema.StringAttribute{
				Description: "Provider name",
				Required:    true,
			},
			"backup_storage_class": schema.StringAttribute{
				Description: "Backup Storage Class will be used to create all backups on this data plane. Please note this cannot be changed in future.",
				Required:    true,
			},
			"data_plane_release_id": schema.StringAttribute{
				Description: "Helm Release Id",
				Required:    true,
			},
			"storage_classes": schema.SetAttribute{
				Description: "Storage Classes on the dataplane",
				ElementType: types.StringType,
				Required:    true,
			},
			"tags": schema.SetAttribute{
				Description: "Tags",
				ElementType: types.StringType,
				Required:    true,
			},
			"services": schema.SetAttribute{
				Description: "Services",
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}

	tflog.Info(ctx, "END__Schema")
}

// Create a new resource
func (r *dataPlaneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__Create")
	// Retrieve values from plan
	var plan dataPlaneResourceModel
	diags := req.Plan.Get(ctx, &plan)

	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan

	dataplaneRequest := infra_connector.DataPlaneCreateRequest{
		AccountId:             plan.AccountId.ValueString(),
		DataplaneName:         plan.Name.ValueString(),
		K8sClusterName:        plan.K8sClusterName.ValueString(),
		DataplaneType:         "regular",
		BackupStorageCLass:    plan.BackupStorageCLass.ValueString(),
		ManagedDns:            true,
		DataPlaneReleaseId:    plan.DataPlaneReleaseId.ValueString(),
		Shared:                plan.Shared.ValueBool(),
		AutoUpgrade:           plan.AutoUpgrade.ValueBool(),
		CpBootstrappedCluster: plan.CpBootstrappedCluster.ValueBool(),
		ConfigureCoreDns:      plan.ConfigureCoreDns.ValueBool(),
	}

	if plan.OrgId.ValueString() != "null" {
		dataplaneRequest.OrgId = plan.OrgId.ValueString()
	}
	query := &infra_connector.DNSQuery{}
	dnsResponse, err := r.client.InfraConnector.GetDnsconfig(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Dns Config List",
			err.Error(),
		)
		return
	}
	dnsDto := *dnsResponse.Get()
	dataplaneRequest.DnsConfigId = dnsDto[0].Id

	certQuery := &infra_connector.CertificateQuery{
		Provider: plan.ProviderName.ValueString(),
	}
	certQuery.Size = 100

	tflog.Debug(ctx, "certqrtrte DTO", map[string]interface{}{"dto": certQuery})
	certificates, err := r.client.InfraConnector.GetCertificates(certQuery)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Certificates",
			err.Error(),
		)
		return
	}

	certDto := *certificates.Get()
	tflog.Debug(ctx, "Creawmejbjwhekgerwfjkg DTO", map[string]interface{}{"dto": certificates})
	dataplaneRequest.CertificateId = certDto[0].Id
	plan.Tags.ElementsAs(ctx, &dataplaneRequest.Tags, true)
	plan.StorageClasses.ElementsAs(ctx, &dataplaneRequest.StorageClasses, true)
	plan.Services.ElementsAs(ctx, &dataplaneRequest.Services, true)
	tflog.Debug(ctx, "Create dataplane DTO", map[string]interface{}{"dto": dataplaneRequest})
	if _, err := r.client.InfraConnector.CreateDataPlane(&dataplaneRequest); err != nil {

		resp.Diagnostics.AddError(
			"Submitting request to create dataplane",
			"Could not create dataplane, unexpected error: "+err.Error(),
		)
		return
	}

	dataplanes, err := r.client.InfraConnector.GetDataPlanes(&infra_connector.DataPlaneQuery{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Fetching DataPlane",
			"Could not fetch data plane, unexpected error: "+err.Error(),
		)
		return
	}

	if len(*dataplanes.Get()) <= 0 {
		resp.Diagnostics.AddError("Fetching dataplane",
			"Unable to fetch the created dataplane",
		)
		return
	}
	createdDataPlane := &(*dataplanes.Get())[0]
	tflog.Debug(ctx, "Created dataplane DTO", map[string]interface{}{"dto": createdDataPlane})
	if saveFromDataPlaneResponse(&ctx, &resp.Diagnostics, &plan, createdDataPlane) != 0 {
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

func (r *dataPlaneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")

	// Retrieve values from plan
	var plan dataPlaneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve current state
	var state dataPlaneResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var updateRequest infra_connector.DataPlaneUpdateRequest
	plan.Tags.ElementsAs(ctx, &updateRequest.Tags, true)
	updateRequest.DataplaneName = plan.Name.ValueString()
	updateRequest.AutoUpgrade = plan.AutoUpgrade.ValueBool()

	// Update existing cluster
	err := r.client.InfraConnector.UpdateDataPlane(plan.ID.ValueString(), &updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating TDH Dataplane",
			"Could not update dataplane, unexpected error: "+err.Error(),
		)
		return
	}

	dataplane, err := r.client.InfraConnector.GetDataPlaneById(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading Dataplane",
			"Could not read dataplane "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	if saveFromDataPlaneResponse(&ctx, &resp.Diagnostics, &state, &dataplane) != 0 {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Update")
}

func (r *dataPlaneResource) Delete(ctx context.Context, request resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "INIT__Delete")
	// Get current state
	var state dataPlaneResourceModel
	diags := request.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Submit request to delete  DataPlane
	err := r.client.InfraConnector.DeleteDataPlane(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Deleting DataPlane",
			"Could not delete dataplane "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	for {
		time.Sleep(10 * time.Second)
		if _, err := r.client.InfraConnector.GetDataPlaneById(state.ID.ValueString()); err != nil {
			if err != nil {
				var apiError core.ApiError
				if errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound {
					break
				}
				resp.Diagnostics.AddError("Fetching cluster",
					fmt.Sprintf("Could not dataplane  by id [%v], unexpected error: %s", state.ID, err.Error()),
				)
				return
			}
		}
	}
	tflog.Info(ctx, "END__Delete")
}

func (r *dataPlaneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *dataPlaneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state dataPlaneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed dataplane value
	dataplane, err := r.client.InfraConnector.GetDataPlaneById(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading Dataplane",
			"Could not read dataplane "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	if saveFromDataPlaneResponse(&ctx, &resp.Diagnostics, &state, &dataplane) != 0 {
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

func saveFromDataPlaneResponse(ctx *context.Context, diagnostics *diag.Diagnostics, state *dataPlaneResourceModel, dataplane *model.DataPlane) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan", map[string]interface{}{"dataplane": *dataplane})

	state.ID = types.StringValue(dataplane.Id)
	state.Name = types.StringValue(dataplane.DataplaneName)
	state.K8sClusterName = types.StringValue(dataplane.Name)
	state.ProviderName = types.StringValue(dataplane.Provider)
	state.DataPlaneReleaseId = types.StringValue(dataplane.DataPlaneReleaseID)
	state.AutoUpgrade = types.BoolValue(dataplane.AutoUpgrade)
	state.AccountId = types.StringValue(dataplane.Account.Id)
	state.CpBootstrappedCluster = types.BoolValue(dataplane.DataPlaneOnControlPlane)
	list, diags := types.SetValueFrom(*ctx, types.StringType, dataplane.Tags)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.Tags = list
	return 0
}
