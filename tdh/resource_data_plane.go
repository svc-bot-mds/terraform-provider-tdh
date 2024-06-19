package tdh

import (
	"context"
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
	infra_connector "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
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
	BackupStorageClass    types.String `tfsdk:"backup_storage_class"`
	Shared                types.Bool   `tfsdk:"shared"`
	OrgId                 types.String `tfsdk:"org_id"`
	Tags                  types.Set    `tfsdk:"tags"`
	AutoUpgrade           types.Bool   `tfsdk:"auto_upgrade"`
	Services              types.Set    `tfsdk:"services"`
	CpBootstrappedCluster types.Bool   `tfsdk:"cp_bootstrapped_cluster"`
	ConfigureCoreDns      types.Bool   `tfsdk:"configure_core_dns"`
	Network               types.String `tfsdk:"network"`
	AvailabilityZone      types.String `tfsdk:"az"`
}

func (r *dataPlaneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_plane"
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
		MarkdownDescription: "Represents a TDH Data Plane.\n ## Note: For SRE only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Auto-generated ID of the data plane after creation, and can be used to import it from TDH to terraform state.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account to use for data plane operations. Please use datasource `tdh_cloud_accounts` to get the list of available accounts.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the Data Plane",
				Required:    true,
			},
			"k8s_cluster_name": schema.StringAttribute{
				MarkdownDescription: "Name of Kubernetes Cluster. Please use datasource `tdh_k8s_clusters` to get the list of available clusters from an account.\n Note: This field is non mandatory during the TAS data plane creation . \n It is a mandatory field during Non TAS (i.e tkgm, tkgs, openshift) data plane creation.",
				Required:            false,
				Optional:            true,
			},
			"shared": schema.BoolAttribute{
				MarkdownDescription: "Shared Data Plane.\n Note: This field should be set to true during the TAS data plane creation . \n It can be set to true/false during Non TAS (i.e tkgm, tkgs, openshift) data plane creation.",
				Required:            true,
			},
			"configure_core_dns": schema.BoolAttribute{
				Description: "Turn on publishing the TDH's DNS to core DNS of the K8S cluster. This will enable communication between pods of the control plane and the data plane with control plane pods. Disable this only when the TDH base URL has a forwarder set in corporate DNS or for some specific use case.\nPlease note that TDH needs these communications between the pods to function.",
				Optional:    true,
			},
			"auto_upgrade": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable auto-upgrade on Data plane.\n Note: This field should be set to false for TAS data-plane creation ",
				Optional:            true,
				Required:            false,
			},
			"cp_bootstrapped_cluster": schema.BoolAttribute{
				MarkdownDescription: "Whether to onboard Data Plane on a K8s cluster running TDH Control Plane \n Note: Not a required field during TAS data-plane creation",
				Optional:            true,
			},
			"org_id": schema.StringAttribute{
				Description: "Organization ID. This filed is not required during TAS data-plane creation",
				Optional:    true,
			},
			"provider_name": schema.StringAttribute{
				Description: "Provider name",
				Required:    true,
			},
			"backup_storage_class": schema.StringAttribute{
				MarkdownDescription: "Backup Storage Class will be used to create all backups on this data plane. Please note this cannot be changed in future.\n Note: This field is non mandatory during the TAS data plane creation. \n It is a mandatory field during  Non TAS (i.e tkgm, tkgs, openshift) data plane creation.",
				Required:            false,
				Optional:            true,
			},
			"data_plane_release_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Helm Release. Please use datasource `tdh_data_plane_helm_releases` to get this.\n Note: This field is non mandatory during the TAS data plane creation. \n It is a mandatory field during  Non TAS (i.e tkgm, tkgs, openshift) data plane creation.",
				Required:            false,
				Optional:            true,
			},
			"storage_classes": schema.SetAttribute{
				MarkdownDescription: "Storage Classes on the data plane. \n Note: This field is non mandatory during the TAS data plane creation . \n It is a mandatory field during Non TAS (i.e tkgm, tkgs, openshift) data plane creation.",
				ElementType:         types.StringType,
				Required:            false,
				Optional:            true,
			},
			"tags": schema.SetAttribute{
				Description: "Tags",
				ElementType: types.StringType,
				Optional:    true,
			},
			"services": schema.SetAttribute{
				MarkdownDescription: "Services. \n Note: TAS data-plane creation supports postgres only",
				ElementType:         types.StringType,
				Required:            true,
			},
			"network": schema.StringAttribute{
				Description: "Network Details. It's a mandatory filed during TAS data-plane creation.",
				Optional:    true,
			},
			"az": schema.StringAttribute{
				Description: "Availability Zone. It's a mandatory filed during TAS data-plane creation.",
				Optional:    true,
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

	dataPlaneRequest := infra_connector.DataPlaneCreateRequest{
		AccountId:             plan.AccountId.ValueString(),
		DataplaneName:         plan.Name.ValueString(),
		DataplaneType:         "regular",
		ManagedDns:            true,
		Shared:                true,
		AutoUpgrade:           false,
		CpBootstrappedCluster: false,
		ConfigureCoreDns:      true,
	}

	if !plan.OrgId.IsNull() && plan.ProviderName.ValueString() != "tas" {
		dataPlaneRequest.OrgId = plan.OrgId.ValueString()
	}
	query := &infra_connector.DNSQuery{}
	dnsResponse, err := r.client.InfraConnector.GetDnsconfig(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH DNS Config List",
			err.Error(),
		)
		return
	}
	dnsDto := *dnsResponse.Get()
	dataPlaneRequest.DnsConfigId = dnsDto[0].Id

	cloudAccountResponse, err := r.client.InfraConnector.GetCloudAccount(plan.AccountId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Cloud Account",
			err.Error(),
		)
		return
	}
	certQuery := &infra_connector.CertificatesQuery{
		Provider: cloudAccountResponse.AccountType,
	}
	// since we need only first certificate entry
	certQuery.Size = 1

	tflog.Debug(ctx, "Certificate query", map[string]interface{}{"cert-query": certQuery})
	certificates, err := r.client.InfraConnector.GetCertificates(certQuery)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Certificates",
			err.Error(),
		)
		return
	}

	certDto := *certificates.Get()
	tflog.Debug(ctx, "Certificate page response", map[string]interface{}{"cert-page": certificates})

	if plan.ProviderName.ValueString() != "tas" {
		dataPlaneRequest.K8sClusterName = plan.K8sClusterName.ValueString()
		dataPlaneRequest.BackupStorageClass = plan.BackupStorageClass.ValueString()
		dataPlaneRequest.CertificateId = certDto[0].Id
		dataPlaneRequest.DataPlaneReleaseId = plan.DataPlaneReleaseId.ValueString()
		dataPlaneRequest.Shared = plan.Shared.ValueBool()
		dataPlaneRequest.AutoUpgrade = plan.AutoUpgrade.ValueBool()
		dataPlaneRequest.CpBootstrappedCluster = plan.CpBootstrappedCluster.ValueBool()
		dataPlaneRequest.ConfigureCoreDns = plan.ConfigureCoreDns.ValueBool()
		plan.Services.ElementsAs(ctx, &dataPlaneRequest.Services, true)
		plan.StorageClasses.ElementsAs(ctx, &dataPlaneRequest.StorageClasses, true)
	} else {
		dataPlaneRequest.Services = append(dataPlaneRequest.Services, "postgres")
		dataPlaneRequest.Network = plan.Network.ValueString()
		dataPlaneRequest.AvailabilityZone = plan.AvailabilityZone.ValueString()
	}

	plan.Tags.ElementsAs(ctx, &dataPlaneRequest.Tags, true)

	tflog.Debug(ctx, "Create data-plane DTO", map[string]interface{}{"request-payload": dataPlaneRequest})
	taskResponse, err := r.client.InfraConnector.CreateDataPlane(&dataPlaneRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating data plane",
			"Could not create data plane, unexpected error: "+err.Error(),
		)
		return
	}

	err = utils.WaitForTask(r.client, taskResponse.TaskId)
	if err != nil {
		resp.Diagnostics.AddError("Creating data plane",
			"Task responsible for this operation failed, error: "+err.Error(),
		)
		return
	}

	dataPlanes, err := r.client.InfraConnector.GetDataPlanes(&infra_connector.DataPlanesQuery{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Fetching Data Plane",
			"Could not fetch data plane, unexpected error: "+err.Error(),
		)
		return
	}

	if len(*dataPlanes.Get()) <= 0 {
		resp.Diagnostics.AddError("Fetching Data Plane",
			"Unable to fetch the created data plane",
		)
		return
	}
	createdDataPlane := &(*dataPlanes.Get())[0]
	tflog.Debug(ctx, "Created data plane", map[string]interface{}{"dto": createdDataPlane})
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
			"Updating Data Plane",
			"Could not update data plane, unexpected error: "+err.Error(),
		)
		return
	}

	dataPlane, err := r.client.InfraConnector.GetDataPlaneById(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading Data Plane",
			"Could not read data plane "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	if saveFromDataPlaneResponse(&ctx, &resp.Diagnostics, &state, &dataPlane) != 0 {
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
	taskResponse, err := r.client.InfraConnector.DeleteDataPlane(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Deleting Data Plane",
			"Could not delete data plane "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	err = utils.WaitForTask(r.client, taskResponse.TaskId)
	if err != nil {
		resp.Diagnostics.AddError("Deleting data plane",
			"Task responsible for this operation failed, error: "+err.Error(),
		)
		return
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
			"Reading Data Plane",
			"Could not read data plane "+state.ID.ValueString()+": "+err.Error(),
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

func saveFromDataPlaneResponse(ctx *context.Context, diagnostics *diag.Diagnostics, state *dataPlaneResourceModel, dataPlane *model.DataPlane) int8 {
	tflog.Info(*ctx, "Saving response to resourceModel state/plan", map[string]interface{}{"data-plane": *dataPlane})

	state.ID = types.StringValue(dataPlane.Id)
	state.Name = types.StringValue(dataPlane.DataplaneName)
	state.K8sClusterName = types.StringValue(dataPlane.Name)
	state.ProviderName = types.StringValue(dataPlane.Provider)
	state.DataPlaneReleaseId = types.StringValue(dataPlane.DataPlaneReleaseID)
	state.AutoUpgrade = types.BoolValue(dataPlane.AutoUpgrade)
	state.AccountId = types.StringValue(dataPlane.Account.Id)
	state.CpBootstrappedCluster = types.BoolValue(dataPlane.DataPlaneOnControlPlane)
	list, diags := types.SetValueFrom(*ctx, types.StringType, dataPlane.Tags)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.Tags = list
	return 0
}
