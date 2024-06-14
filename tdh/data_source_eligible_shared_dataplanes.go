package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	infra_connector "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &eligibleSharedDataplanesDatasource{}
	_ datasource.DataSourceWithConfigure = &eligibleSharedDataplanesDatasource{}
)

// EligibleDataplanesDatasourceModel maps the data source schema data.
type EligibleSharedDataplanesDatasourceModel struct {
	Id             types.String           `tfsdk:"id"`
	Provider       types.String           `tfsdk:"provider_name"`
	DataplaneModel []SharedDataplaneModel `tfsdk:"dataplanes"`
}

type SharedDataplaneModel struct {
	ID                  types.String `tfsdk:"id"`
	DataplaneName       types.String `tfsdk:"dataplane_name"`
	StoragePolicies     types.Set    `tfsdk:"storage_policies"`
	BackupStoragePolicy types.String `tfsdk:"backup_storage_policy"`
}

// NewDataplaneDatasource is a helper function to simplify the provider implementation.
func NewEligibleSharedDataplanesDatasource() datasource.DataSource {
	return &eligibleSharedDataplanesDatasource{}
}

// eligibleSharedDataplanesDatasource is the data source implementation.
type eligibleSharedDataplanesDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *eligibleSharedDataplanesDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eligible_shared_dataplanes"
}

// Schema defines the schema for the data source.
func (d *eligibleSharedDataplanesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all Shared Dataplanes by passing the Provider input .The response will have Storage Policies which can be used while creating cluster",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},

			"provider_name": schema.StringAttribute{
				Description: "Provider Name",
				Required:    true,
			},
			"dataplanes": schema.ListNestedAttribute{
				Computed: true,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the dataplane.",
							Computed:    true,
						},
						"dataplane_name": schema.StringAttribute{
							Description: "Dataplane Name",
							Computed:    true,
						},
						"storage_policies": schema.SetAttribute{
							Description: "Storage Policies of the dataplane",
							ElementType: types.StringType,
							Computed:    true,
						},
						"backup_storage_policy": schema.StringAttribute{
							Description: "Backup storage Class",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *eligibleSharedDataplanesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EligibleSharedDataplanesDatasourceModel
	var dataplaneList []SharedDataplaneModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &infra_connector.EligibleSharedDataPlaneQuery{
		InfraResourceType: "SHARED",
		Provider:          state.Provider.ValueString(),
	}

	dataplanes, err := d.client.InfraConnector.GetEligibleSharedDataPlanes(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Eligible Shared Dataplanes for "+state.Provider.ValueString(),
			err.Error(),
		)
		return
	}

	if dataplanes.Page.TotalPages > 1 {
		for i := 1; i <= dataplanes.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalCloudAccounts, err := d.client.InfraConnector.GetEligibleSharedDataPlanes(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH Eligible Shared Dataplanes for "+state.Provider.ValueString(),
					err.Error(),
				)
				return
			}

			for _, dataplaneDto := range *totalCloudAccounts.Get() {
				dataplane, err := d.convertToTfModel(ctx, dataplaneDto, resp)
				if err {
					return
				}
				dataplaneList = append(dataplaneList, dataplane)
			}
		}

		tflog.Debug(ctx, "dp dto", map[string]interface{}{"dto": dataplaneList})
		state.DataplaneModel = append(state.DataplaneModel, dataplaneList...)
	} else {
		for _, dpDto := range *dataplanes.Get() {
			tflog.Info(ctx, "Converting dataplane dto")
			dataplane, err := d.convertToTfModel(ctx, dpDto, resp)
			if err {
				return
			}
			tflog.Debug(ctx, "converted dataplane dto", map[string]interface{}{"dto": dataplane})
			state.DataplaneModel = append(state.DataplaneModel, dataplane)
		}
	}
	state.Id = types.StringValue(common.DataSource + common.DataplaneId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *eligibleSharedDataplanesDatasource) convertToTfModel(ctx context.Context, dpDto model.EligibleSharedDataPlane, resp *datasource.ReadResponse) (SharedDataplaneModel, bool) {
	dataplane := SharedDataplaneModel{
		ID:                  types.StringValue(dpDto.Id),
		DataplaneName:       types.StringValue(dpDto.DataplaneName),
		BackupStoragePolicy: types.StringValue(dpDto.BackupStoragePolicy),
	}
	list, diags := types.SetValueFrom(ctx, types.StringType, dpDto.StoragePolicies)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return SharedDataplaneModel{}, true
	}
	dataplane.StoragePolicies = list
	return dataplane, false
}

// Configure adds the provider configured client to the data source.
func (d *eligibleSharedDataplanesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
