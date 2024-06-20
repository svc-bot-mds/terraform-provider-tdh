package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &eligibleDataPlanesDatasource{}
	_ datasource.DataSourceWithConfigure = &eligibleDataPlanesDatasource{}
)

// EligibleDataPlanesDatasourceModel maps the data source schema data.
type EligibleDataPlanesDatasourceModel struct {
	Id       types.String     `tfsdk:"id"`
	Provider types.String     `tfsdk:"provider_name"`
	OrgId    types.String     `tfsdk:"org_id"`
	List     []dataPlaneModel `tfsdk:"list"`
}

type dataPlaneModel struct {
	ID                  types.String `tfsdk:"id"`
	DataPlaneName       types.String `tfsdk:"data_plane_name"`
	StoragePolicies     types.Set    `tfsdk:"storage_policies"`
	BackupStoragePolicy types.String `tfsdk:"backup_storage_policy"`
}

// NewEligibleDataPlanesDatasource is a helper function to simplify the provider implementation.
func NewEligibleDataPlanesDatasource() datasource.DataSource {
	return &eligibleDataPlanesDatasource{}
}

// eligibleDataPlanesDatasource is the data source implementation.
type eligibleDataPlanesDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *eligibleDataPlanesDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eligible_data_planes"
}

// Schema defines the schema for the data source.
func (d *eligibleDataPlanesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all Data planes.\n" +
			"## Note:\n" +
			"- This datasource is using during the service cluster creation",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"provider_name": schema.StringAttribute{
				Description: "Provider Name",
				Required:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Org ID, can be left out to filter shared data planes.",
				Optional:    true,
			},
			"list": schema.ListNestedAttribute{
				Computed: true,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the data plane. This is used during service cluster creation.",
							Computed:    true,
						},
						"data_plane_name": schema.StringAttribute{
							Description: "Name of the data plane",
							Computed:    true,
						},
						"storage_policies": schema.SetAttribute{
							Description: "Storage Policies of the data plane",
							ElementType: types.StringType,
							Computed:    true,
						},
						"backup_storage_policy": schema.StringAttribute{
							Description: "Name of the storage class set for backup purposes.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *eligibleDataPlanesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EligibleDataPlanesDatasourceModel
	var dataPlaneList []dataPlaneModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &infra_connector.EligibleDataPlanesQuery{
		InfraResourceType: "SHARED",
		Provider:          state.Provider.ValueString(),
	}
	if !state.OrgId.IsNull() {
		query.InfraResourceType = "DEDICATED"
		query.OrgId = state.OrgId.ValueString()
	}
	dataPlanes, err := d.client.InfraConnector.GetEligibleDataPlanes(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH eligible data planes for "+state.Provider.ValueString(),
			err.Error(),
		)
		return
	}

	if dataPlanes.Page.TotalPages > 1 {
		for i := 1; i <= dataPlanes.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalCloudAccounts, err := d.client.InfraConnector.GetEligibleDataPlanes(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH eligible data planes for "+state.Provider.ValueString(),
					err.Error(),
				)
				return
			}

			for _, dto := range *totalCloudAccounts.Get() {
				dataPlaneModel, err := d.convertToTfModel(ctx, dto, resp)
				if err {
					return
				}
				dataPlaneList = append(dataPlaneList, dataPlaneModel)
			}
		}

		tflog.Debug(ctx, "dp dto", map[string]interface{}{"dto": dataPlaneList})
	} else {
		for _, dpDto := range *dataPlanes.Get() {
			tflog.Info(ctx, "Converting data plane dto")
			dataPlaneModel, err := d.convertToTfModel(ctx, dpDto, resp)
			if err {
				return
			}
			tflog.Debug(ctx, "converted data plane dto", map[string]interface{}{"dto": dataPlaneModel})
			dataPlaneList = append(dataPlaneList, dataPlaneModel)
		}
	}
	state.List = append(state.List, dataPlaneList...)
	state.Id = types.StringValue(common.DataSource + common.DataplaneId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *eligibleDataPlanesDatasource) convertToTfModel(ctx context.Context, dpDto model.EligibleDataPlane, resp *datasource.ReadResponse) (dataPlaneModel, bool) {
	dataPlane := dataPlaneModel{
		ID:                  types.StringValue(dpDto.Id),
		DataPlaneName:       types.StringValue(dpDto.DataPlaneName),
		BackupStoragePolicy: types.StringValue(dpDto.BackupStoragePolicy),
	}
	list, diags := types.SetValueFrom(ctx, types.StringType, dpDto.StoragePolicies)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return dataPlane, true
	}
	dataPlane.StoragePolicies = list
	return dataPlane, false
}

// Configure adds the provider configured client to the data source.
func (d *eligibleDataPlanesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
