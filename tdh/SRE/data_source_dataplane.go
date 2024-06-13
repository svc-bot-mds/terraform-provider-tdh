package SRE

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
	_ datasource.DataSource              = &dataplaneDatasource{}
	_ datasource.DataSourceWithConfigure = &dataplaneDatasource{}
)

// instanceTypesDataSourceModel maps the data source schema data.
type dataplaneDatasourceModel struct {
	Id             types.String     `tfsdk:"id"`
	DataplaneModel []dataplaneModel `tfsdk:"dataplane"`
}

type dataplaneModel struct {
	ID                      types.String     `tfsdk:"id"`
	Provider                types.String     `tfsdk:"provider"`
	Region                  types.String     `tfsdk:"region"`
	Name                    types.String     `tfsdk:"name"`
	DataplaneName           types.String     `tfsdk:"dataplane_name"`
	Version                 types.String     `tfsdk:"version"`
	Tags                    types.Set        `tfsdk:"tags"`
	Status                  types.String     `tfsdk:"status"`
	Account                 AccountModel     `tfsdk:"account"`
	HelmVersionId           types.String     `tfsdk:"dataplane_release_id"`
	HelmVersion             types.String     `tfsdk:"dataplane_release_name"`
	Shared                  types.Bool       `tfsdk:"shared"`
	AutoUpgrade             types.Bool       `tfsdk:"auto_upgrade"`
	Created                 types.String     `tfsdk:"created"`
	Modified                types.String     `tfsdk:"modified"`
	Certificate             CertificateModel `tfsdk:"certificate"`
	DefaultPolicyName       types.String     `tfsdk:"default_policy_name"`
	StoragePolicies         types.Set        `tfsdk:"storage_policies"`
	BackupStoragePolicy     types.String     `tfsdk:"backup_storage_policy"`
	Services                types.Set        `tfsdk:"services"`
	DataPlaneOnControlPlane types.Bool       `tfsdk:"data_plane_on_control_plane"
"`
}

type AccountModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type CertificateModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewDataplaneDatasource is a helper function to simplify the provider implementation.
func NewDataplaneDatasource() datasource.DataSource {
	return &dataplaneDatasource{}
}

// dataplaneDatasource is the data source implementation.
type dataplaneDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *dataplaneDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataplanes"
}

// Schema defines the schema for the data source.
func (d *dataplaneDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all Dataplanes",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"dataplane": schema.ListNestedAttribute{
				Computed: true,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the dataplane.",
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
							Description: "Dataplane Region",
							Computed:    true,
						},
						"dataplane_name": schema.StringAttribute{
							Description: "Dataplane Name",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Dataplane Version",
							Computed:    true,
						},
						"tags": schema.SetAttribute{
							Description: "Tags set on this dataplane.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"storage_policies": schema.SetAttribute{
							Description: "Storage Policies of the dataplane",
							ElementType: types.StringType,
							Computed:    true,
						},
						"services": schema.SetAttribute{
							Description: "Services available on the dataplane",
							ElementType: types.StringType,
							Computed:    true,
						},
						"shared": schema.BoolAttribute{
							Description: "Whether this account is shared between multiple Organisations or not.",
							Computed:    true,
						},
						"auto_upgrade": schema.BoolAttribute{
							Description: "Flag to set dataplane autoupgradable",
							Computed:    true,
						},
						"created": schema.StringAttribute{
							Description: "Creation time of this dataplane.",
							Computed:    true,
						},
						"modified": schema.StringAttribute{
							Description: "Modified time of this dataplane.",
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
							Description: "Status of the Dataplane",
							Computed:    true,
						},
						"dataplane_release_name": schema.StringAttribute{
							Description: "Helm Version",
							Computed:    true,
						},
						"dataplane_release_id": schema.StringAttribute{
							Description: "Helm Version Id",
							Computed:    true,
						},
						"data_plane_on_control_plane": schema.BoolAttribute{
							Description: "Dataplane on Controlplane",
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
func (d *dataplaneDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataplaneDatasourceModel
	var dataplaneList []dataplaneModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &infra_connector.DataPlaneQuery{}

	dataplanes, err := d.client.InfraConnector.GetDataPlanes(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Dataplanes",
			err.Error(),
		)
		return
	}

	if dataplanes.Page.TotalPages > 1 {
		for i := 1; i <= dataplanes.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalCloudAccounts, err := d.client.InfraConnector.GetDataPlanes(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH dataplanes",
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

func (d *dataplaneDatasource) convertToTfModel(ctx context.Context, dpDto model.DataPlane, resp *datasource.ReadResponse) (dataplaneModel, bool) {
	dataplane := dataplaneModel{
		ID:                      types.StringValue(dpDto.Id),
		Name:                    types.StringValue(dpDto.Name),
		Provider:                types.StringValue(dpDto.Provider),
		Region:                  types.StringValue(dpDto.Region),
		DataplaneName:           types.StringValue(dpDto.DataplaneName),
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
	dataplane.Account = accountDetails

	certDetails := CertificateModel{
		Id:   types.StringValue(dpDto.Certificate.Id),
		Name: types.StringValue(dpDto.Certificate.Name),
	}
	dataplane.Certificate = certDetails
	list, diags := types.SetValueFrom(ctx, types.StringType, dpDto.Tags)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return dataplaneModel{}, true
	}
	dataplane.Tags = list

	storagePolicies, diags := types.SetValueFrom(ctx, types.StringType, dpDto.StoragePolicies)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return dataplaneModel{}, true
	}
	dataplane.StoragePolicies = storagePolicies

	services, diags := types.SetValueFrom(ctx, types.StringType, dpDto.Services)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return dataplaneModel{}, true
	}
	dataplane.Services = services
	return dataplane, false
}

// Configure adds the provider configured client to the data source.
func (d *dataplaneDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
