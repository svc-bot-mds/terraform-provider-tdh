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
	_ datasource.DataSource              = &dataPlaneDatasource{}
	_ datasource.DataSourceWithConfigure = &dataPlaneDatasource{}
)

// dataPlaneDatasourceModel maps the data source schema data.
type dataPlaneDatasourceModel struct {
	Id   types.String     `tfsdk:"id"`
	List []DataPlaneModel `tfsdk:"list"`
}

type DataPlaneModel struct {
	ID                      types.String     `tfsdk:"id"`
	Provider                types.String     `tfsdk:"provider"`
	Region                  types.String     `tfsdk:"region"`
	Name                    types.String     `tfsdk:"name"`
	DataPlaneName           types.String     `tfsdk:"data_plane_name"`
	Version                 types.String     `tfsdk:"version"`
	Tags                    types.Set        `tfsdk:"tags"`
	Status                  types.String     `tfsdk:"status"`
	Account                 AccountModel     `tfsdk:"account"`
	HelmVersionId           types.String     `tfsdk:"data_plane_release_id"`
	HelmVersion             types.String     `tfsdk:"data_plane_release_name"`
	Shared                  types.Bool       `tfsdk:"shared"`
	AutoUpgrade             types.Bool       `tfsdk:"auto_upgrade"`
	Created                 types.String     `tfsdk:"created"`
	Modified                types.String     `tfsdk:"modified"`
	Certificate             CertificateModel `tfsdk:"certificate"`
	DefaultPolicyName       types.String     `tfsdk:"default_policy_name"`
	StoragePolicies         types.Set        `tfsdk:"storage_policies"`
	BackupStoragePolicy     types.String     `tfsdk:"backup_storage_policy"`
	Services                types.Set        `tfsdk:"services"`
	DataPlaneOnControlPlane types.Bool       `tfsdk:"data_plane_on_control_plane"`
}

type AccountModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type CertificateModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewDataPlaneDatasource is a helper function to simplify the provider implementation.
func NewDataPlaneDatasource() datasource.DataSource {
	return &dataPlaneDatasource{}
}

// dataPlaneDatasource is the data source implementation.
type dataPlaneDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *dataPlaneDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_planes"
}

// Schema defines the schema for the data source.
func (d *dataPlaneDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all Data planes",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"list": schema.ListNestedAttribute{
				Computed: true,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the data plane.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the TKC.",
							Computed:    true,
						},
						"provider": schema.StringAttribute{
							Description: "Provider Name",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Description: "Data plane Region",
							Computed:    true,
						},
						"data_plane_name": schema.StringAttribute{
							Description: "Data plane Name",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Data plane Version",
							Computed:    true,
						},
						"tags": schema.SetAttribute{
							Description: "Tags set on this data plane.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"storage_policies": schema.SetAttribute{
							Description: "Storage Policies of the data plane",
							ElementType: types.StringType,
							Computed:    true,
						},
						"services": schema.SetAttribute{
							Description: "Services available on the data plane",
							ElementType: types.StringType,
							Computed:    true,
						},
						"shared": schema.BoolAttribute{
							Description: "Whether this account is shared between multiple Organisations or not.",
							Computed:    true,
						},
						"auto_upgrade": schema.BoolAttribute{
							Description: "Whether auto-upgrade is enabled on this data plane or not",
							Computed:    true,
						},
						"created": schema.StringAttribute{
							Description: "Creation time of this data plane.",
							Computed:    true,
						},
						"modified": schema.StringAttribute{
							Description: "Modified time of this data plane.",
							Computed:    true,
						},
						"default_policy_name": schema.StringAttribute{
							Description: "Default Policy Name",
							Computed:    true,
						},
						"backup_storage_policy": schema.StringAttribute{
							Description: "Backup Storage Policy",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the data plane",
							Computed:    true,
						},
						"data_plane_release_name": schema.StringAttribute{
							Description: "Helm Version",
							Computed:    true,
						},
						"data_plane_release_id": schema.StringAttribute{
							Description: "Helm Version Id",
							Computed:    true,
						},
						"data_plane_on_control_plane": schema.BoolAttribute{
							Description: "Whether Data plane is deployed on same cluster as TDH Control plane",
							Computed:    true,
						},
						"account": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "ID of the cloud account.",
									Computed:    true,
								},
								"name": schema.StringAttribute{
									Description: "Name of the cloud account.",
									Computed:    true,
								},
							},
						},
						"certificate": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "ID of the certificate.",
									Computed:    true,
								},
								"name": schema.StringAttribute{
									Description: "Name of the certificate.",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dataPlaneDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataPlaneDatasourceModel
	var dataPlaneList []DataPlaneModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &infra_connector.DataPlanesQuery{}

	dataPlanes, err := d.client.InfraConnector.GetDataPlanes(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Data Planes",
			err.Error(),
		)
		return
	}

	if dataPlanes.Page.TotalPages > 1 {
		for i := 1; i <= dataPlanes.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalCloudAccounts, err := d.client.InfraConnector.GetDataPlanes(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH Data Planes",
					err.Error(),
				)
				return
			}

			for _, dataPlaneDto := range *totalCloudAccounts.Get() {
				dataPlane, err := d.convertToTfModel(ctx, dataPlaneDto, resp)
				if err {
					return
				}
				dataPlaneList = append(dataPlaneList, *dataPlane)
			}
		}

		tflog.Debug(ctx, "dp dto", map[string]interface{}{"dto": dataPlaneList})
		state.List = append(state.List, dataPlaneList...)
	} else {
		for _, dpDto := range *dataPlanes.Get() {
			tflog.Info(ctx, "Converting data plane dto")
			dataPlane, err := d.convertToTfModel(ctx, dpDto, resp)
			if err {
				return
			}
			tflog.Debug(ctx, "converted data plane dto", map[string]interface{}{"dto": dataPlane})
			state.List = append(state.List, *dataPlane)
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

func (d *dataPlaneDatasource) convertToTfModel(ctx context.Context, dpDto model.DataPlane, resp *datasource.ReadResponse) (*DataPlaneModel, bool) {
	dataPlane := DataPlaneModel{
		ID:                      types.StringValue(dpDto.Id),
		Name:                    types.StringValue(dpDto.Name),
		Provider:                types.StringValue(dpDto.Provider),
		Region:                  types.StringValue(dpDto.Region),
		DataPlaneName:           types.StringValue(dpDto.DataplaneName),
		Shared:                  types.BoolValue(dpDto.Shared),
		Version:                 types.StringValue(dpDto.Version),
		Status:                  types.StringValue(dpDto.Status),
		Created:                 types.StringValue(dpDto.Created),
		Modified:                types.StringValue(dpDto.Modified),
		HelmVersion:             types.StringValue(dpDto.DataPlaneReleaseName),
		HelmVersionId:           types.StringValue(dpDto.DataPlaneReleaseID),
		DataPlaneOnControlPlane: types.BoolValue(dpDto.DataPlaneOnControlPlane),
		AutoUpgrade:             types.BoolValue(dpDto.AutoUpgrade),
		DefaultPolicyName:       types.StringValue(dpDto.DefaultPolicyName),
		BackupStoragePolicy:     types.StringValue(dpDto.BackupStoragePolicy),
	}

	accountDetails := AccountModel{
		Id:   types.StringValue(dpDto.Account.Id),
		Name: types.StringValue(dpDto.Account.Name),
	}
	dataPlane.Account = accountDetails

	certDetails := CertificateModel{
		Id:   types.StringValue(dpDto.Certificate.Id),
		Name: types.StringValue(dpDto.Certificate.Name),
	}
	dataPlane.Certificate = certDetails
	list, diags := types.SetValueFrom(ctx, types.StringType, dpDto.Tags)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return nil, true
	}
	dataPlane.Tags = list

	storagePolicies, diags := types.SetValueFrom(ctx, types.StringType, dpDto.StoragePolicies)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return nil, true
	}
	dataPlane.StoragePolicies = storagePolicies

	services, diags := types.SetValueFrom(ctx, types.StringType, dpDto.Services)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return nil, true
	}
	dataPlane.Services = services
	return &dataPlane, false
}

// Configure adds the provider configured client to the data source.
func (d *dataPlaneDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
