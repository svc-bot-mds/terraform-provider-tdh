package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/validators"
	"slices"
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
	Status                types.String `tfsdk:"status"`
	AutoUpgrade           types.Bool   `tfsdk:"auto_upgrade"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	Sync                  types.Bool   `tfsdk:"sync"`
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
		MarkdownDescription: "Represents a TDH Data Plane.\n" +
			"**Note:** For SRE only.",
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
				Validators: []validator.String{
					validators.UUIDValidator{},
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Data Plane",
				Required:    true,
			},
			"k8s_cluster_name": schema.StringAttribute{
				MarkdownDescription: "Name of Kubernetes Cluster. Please use datasource `tdh_k8s_clusters` to get the list of available clusters from an account.\n" +
					"## Notes\n" +
					"- This field is non-mandatory during the TAS data plane creation.\n" +
					"- It is a mandatory field during Non TAS (i.e `tkgm`, `tkgs`, `openshift`)	data plane creation.",
				Required: true,
				Validators: []validator.String{
					validators.EmptyStringValidator{},
				},
			},
			"shared": schema.BoolAttribute{
				MarkdownDescription: "Shared Data Plane.\n" +
					"## Notes\n" +
					"- This field should be set to true during the TAS data plane creation.\n" +
					"- It can be set to true/false during Non TAS (i.e `tkgm`, `tkgs`, `openshift`) data plane creation.",
				Required: true,
			},
			"configure_core_dns": schema.BoolAttribute{
				Description: "Turn on publishing the TDH's DNS to core DNS of the K8S cluster. This will enable communication between pods of the control plane and the data plane with control plane pods. Disable this only when the TDH base URL has a forwarder set in corporate DNS or for some specific use case.\n" +
					"Please note that TDH needs these communications between the pods to function.",
				Optional: true,
			},
			"auto_upgrade": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable auto-upgrade on Data plane.\n" +
					"**Note:** This field should be set to false for TAS data-plane creation.",
				Optional: true,
				Required: false,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable the Data plane.\n" +
					"**Note:** This field should be omitted or set to true for TAS data-plane creation.",
				Optional: true,
				Required: false,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"sync": schema.BoolAttribute{
				MarkdownDescription: "Set this to `true` whenever syncing is required.",
				Optional:            true,
				Required:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cp_bootstrapped_cluster": schema.BoolAttribute{
				MarkdownDescription: "Whether to onboard Data Plane on a K8s cluster running TDH Control Plane.\n" +
					"**Note:** Not a required field during TAS data-plane creation.",
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Description: "Organization ID. This filed is not required during TAS data-plane creation",
				Optional:    true,
				Validators: []validator.String{
					validators.UUIDValidator{},
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the data plane",
				Computed:    true,
			},
			"provider_name": schema.StringAttribute{
				Description: "Provider name",
				Required:    true,
			},
			"backup_storage_class": schema.StringAttribute{
				MarkdownDescription: "Backup Storage Class will be used to create all backups on this data plane. Please note this cannot be changed in future.\n" +
					"## Notes\n" +
					"- This field is non-mandatory during the TAS data plane creation.\n" +
					"- It is a mandatory field during Non TAS (i.e `tkgm`, `tkgs`, `openshift`)	data plane creation.",
				Required: false,
				Optional: true,
			},
			"data_plane_release_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Helm Release. Please use datasource `tdh_data_plane_helm_releases` to get this.\n" +
					"## Notes\n" +
					"- This field is non-mandatory during the TAS data plane creation.\n" +
					"- It is a mandatory field during Non TAS (i.e `tkgm`, `tkgs`, `openshift`)	data plane creation.",
				Required: false,
				Optional: true,
				Validators: []validator.String{
					validators.UUIDValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_classes": schema.SetAttribute{
				MarkdownDescription: "Storage Classes on the data plane.\n" +
					"## Notes\n" +
					"- This field is non-mandatory during the TAS data plane creation.\n" +
					"- It is a mandatory field during Non TAS (i.e `tkgm`, `tkgs`, `openshift`)	data plane creation.",
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.SetAttribute{
				Description: "Tags",
				ElementType: types.StringType,
				Optional:    true,
			},
			"services": schema.SetAttribute{
				MarkdownDescription: "Services to support on this data plane. Please use datasource `tdh_data_plane_helm_releases` to get the list of available services in a release.\n**Note:** TAS data-plane creation supports `postgres` only.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"network": schema.StringAttribute{
				Description: "Network Details. It's a mandatory field during TAS data-plane creation.",
				Optional:    true,
			},
			"az": schema.StringAttribute{
				Description: "Availability Zone. It's a mandatory field during TAS data-plane creation.",
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

	if r.validateDpCreateInputs(plan, &resp.Diagnostics).HasError() {
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
		dataPlaneRequest.OrgId = plan.OrgId.ValueString()
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

	var state, plan dataPlaneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() { // Retrieve values from plan
		return
	}
	diags = req.State.Get(ctx, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() { // Retrieve current state
		return
	}

	if !state.Services.Equal(plan.Services) {
		tflog.Debug(ctx, "services are changed in plan, validating...")
		for _, currentService := range state.Services.Elements() {
			if !slices.Contains(plan.Services.Elements(), currentService) {
				resp.Diagnostics.AddError(
					"Updating Data Plane",
					"Removing a service is not allowed",
				)
				return
			}
		}

		tflog.Debug(ctx, "service change is valid, proceeding...")
		req := infra_connector.DataPlaneUpdateServicesRequest{
			DataPlaneId: state.ID.ValueString(),
		}
		if plan.Services.ElementsAs(ctx, &req.Services, true).HasError() {
			return
		}
		taskResponse, err := r.client.InfraConnector.UpdateDataPlaneServices(&req)
		if err != nil {
			resp.Diagnostics.AddError(
				"Updating Data Plane",
				"Could not update data plane, unexpected error: "+err.Error(),
			)
			return
		}
		if err = utils.WaitForTask(r.client, taskResponse.TaskId); err != nil {
			resp.Diagnostics.AddError("Updating data plane",
				"Operation error: "+err.Error())
			return
		}
	}

	// Generate API request body from plan
	var updateRequest infra_connector.DataPlaneUpdateRequest
	plan.Tags.ElementsAs(ctx, &updateRequest.Tags, true)
	updateRequest.DataplaneName = plan.Name.ValueString()
	updateRequest.AutoUpgrade = plan.AutoUpgrade.ValueBool()
	updateRequest.Enable = plan.Enabled.ValueBool()

	tflog.Debug(ctx, "hitting update request", map[string]interface{}{
		"req": updateRequest,
	})
	// Update existing cluster
	err := r.client.InfraConnector.UpdateDataPlane(plan.ID.ValueString(), &updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating Data Plane",
			"Could not update data plane, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.Sync.ValueBool() {
		tflog.Debug(ctx, "Triggering Sync request")
		taskResponse, err := r.client.InfraConnector.SyncDataPlane(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Updating Data Plane",
				"Could not trigger data plane sync, error: "+err.Error(),
			)
			return
		}
		if err = utils.WaitForTask(r.client, taskResponse.TaskId); err != nil {
			resp.Diagnostics.AddError("Updating data plane",
				"Sync operation error: "+err.Error())
			return
		}
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
	if saveFromDataPlaneResponse(&ctx, &resp.Diagnostics, &plan, &dataPlane) != 0 {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
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

	state.Sync = types.BoolValue(false)
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
	state.Shared = types.BoolValue(dataPlane.Shared)
	state.Status = types.StringValue(dataPlane.Status)
	state.Enabled = types.BoolValue(dataPlane.Enabled)
	list, diags := types.SetValueFrom(*ctx, types.StringType, dataPlane.Services)
	if diagnostics.Append(diags...); diags.HasError() {
		return 1
	}
	state.Services = list
	list, diags = types.SetValueFrom(*ctx, types.StringType, dataPlane.StoragePolicies)
	if diagnostics.Append(diags...); diags.HasError() {
		return 1
	}
	state.StorageClasses = list

	list, diags = types.SetValueFrom(*ctx, types.StringType, dataPlane.Tags)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return 1
	}
	state.Tags = list
	return 0
}

func (r *dataPlaneResource) validateDpCreateInputs(plan dataPlaneResourceModel, diag *diag.Diagnostics) *diag.Diagnostics {
	if plan.ProviderName.ValueString() == "tas" {
		if plan.Network.IsNull() {
			diag.AddError("Invalid inputs", "'network' is required during tas data plane create operation.")
			return diag
		}
		if plan.ProviderName.ValueString() == "tas" && plan.AvailabilityZone.IsNull() {
			diag.AddError("Invalid inputs", "'az' is required during tas data plane create operation.")
			return diag
		}
		if plan.Shared.ValueBool() == false || (!plan.OrgId.IsNull() && plan.Shared.ValueBool() == true) {
			diag.AddError("Invalid inputs", "Cannot create dedicated data plane for the tas provider. Please Set the value of 'shared' to true and 'org_id' to null")
			return diag
		}
	}

	if plan.ProviderName.ValueString() != "tas" {
		if plan.Services.IsNull() {
			diag.AddError("Invalid inputs", "'services' is required during data plane create operation of 'tkgm','tkgs', 'openshift'")
			return diag
		}
		if plan.StorageClasses.IsNull() {
			diag.AddError("Invalid inputs", "'storage_classes' is required during data plane create operation of 'tkgm','tkgs', 'openshift'")
			return diag
		}
		if plan.DataPlaneReleaseId.IsNull() {
			diag.AddError("Invalid inputs", "'data_plane_release_id' is required during data plane create operation of 'tkgm','tkgs', 'openshift'")
			return diag
		}
		if plan.BackupStorageClass.IsNull() {
			diag.AddError("Invalid inputs", "'backup_storage_class' is required during data plane create operation of 'tkgm','tkgs', 'openshift'")
			return diag
		}
		if plan.CpBootstrappedCluster.IsNull() {
			diag.AddError("Invalid inputs", "'cp_bootstrapped_cluster' is required during data plane create operation of 'tkgm','tkgs', 'openshift'")
			return diag
		}
		if plan.ConfigureCoreDns.IsNull() {
			diag.AddError("Invalid inputs", "'configure_core_dns' is required during data plane create operation of 'tkgm','tkgs', 'openshift'")
			return diag
		}
		if plan.AutoUpgrade.IsNull() {
			diag.AddError("Invalid inputs", "'auto_upgrade' is required during data plane create operation of 'tkgm','tkgs', 'openshift'")
			return diag
		}
		if plan.K8sClusterName.IsNull() {
			diag.AddError("Invalid inputs", "'k8s_cluster_name' is required during data plane create operation of 'tkgm','tkgs', 'openshift'")
			return diag
		}

		if plan.Shared.ValueBool() == false && plan.OrgId.IsNull() {
			diag.AddError("Invalid inputs", "Cannot create dedicated data plane for the provider. Please Set the value of 'org_id' to valid organization id")
			return diag
		} else if !plan.OrgId.IsNull() && plan.Shared.ValueBool() == true {
			diag.AddError("Invalid inputs", "Cannot create shared data plane for the provider. Please Set the value of 'org_id' to null")
			return diag
		}

	}
	return diag
}
