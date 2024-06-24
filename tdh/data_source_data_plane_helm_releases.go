package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &dataPlaneHelmReleaseDatasource{}
	_ datasource.DataSourceWithConfigure = &dataPlaneHelmReleaseDatasource{}
)

// dataPlaneHelmReleasesDataSourceModel maps the data source schema data.
type dataPlaneHelmReleasesDataSourceModel struct {
	Id   types.String       `tfsdk:"id"`
	List []helmReleaseModel `tfsdk:"list"`
}
type helmReleaseModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Services types.Set    `tfsdk:"services"`
}

// NewDataPlaneHelmReleaseDatasource is a helper function to simplify the provider implementation.
func NewDataPlaneHelmReleaseDatasource() datasource.DataSource {
	return &dataPlaneHelmReleaseDatasource{}
}

// rolesDatasource is the data source implementation.
type dataPlaneHelmReleaseDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *dataPlaneHelmReleaseDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_plane_helm_releases"

}

// Schema defines the schema for the data source.
func (d *dataPlaneHelmReleaseDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch all Helm Releases for the data plane.\n" +
			"**Note:** For SRE only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"list": schema.ListNestedAttribute{
				MarkdownDescription: "List of Helm Releases / Data plane Versions. Please use the list item which has `enabled` set to true.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the Helm Release. Can be used while onboarding data plane.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the Helm Release.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Denotes if the helm release is enabled/not. Use the helm release with the flag set to true while onboarding the data plane.",
						},
						"services": schema.SetAttribute{
							Computed:    true,
							Description: "Services available for the Helm Release.",
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dataPlaneHelmReleaseDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataPlaneHelmReleasesDataSourceModel
	var helmReleaseList []helmReleaseModel
	tflog.Info(ctx, "INIT -- READ service roles")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &infra_connector.DNSQuery{}
	helmVersions, err := d.client.InfraConnector.GetHelmRelease(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Helm Versions",
			err.Error(),
		)
		return
	}
	if helmVersions.Page.TotalPages > 1 {
		for i := 1; i <= helmVersions.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalPolicies, err := d.client.InfraConnector.GetHelmRelease(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH Helm Versions",
					err.Error(),
				)
				return
			}

			for _, helmDto := range *totalPolicies.Get() {
				helmRelease := helmReleaseModel{
					Name:    types.StringValue(helmDto.Name),
					Id:      types.StringValue(helmDto.Id),
					Enabled: types.BoolValue(helmDto.IsEnabled),
				}
				services, diags := types.SetValueFrom(ctx, types.StringType, helmDto.Services)
				resp.Diagnostics.Append(diags...)
				helmRelease.Services = services
				helmReleaseList = append(helmReleaseList, helmRelease)
			}

		}

		state.List = append(state.List, helmReleaseList...)
	} else {
		for _, helmDto := range *helmVersions.Get() {
			helmRelease := helmReleaseModel{
				Name:    types.StringValue(helmDto.Name),
				Id:      types.StringValue(helmDto.Id),
				Enabled: types.BoolValue(helmDto.IsEnabled),
			}

			services, diags := types.SetValueFrom(ctx, types.StringType, helmDto.Services)
			resp.Diagnostics.Append(diags...)
			helmRelease.Services = services
			helmReleaseList = append(helmReleaseList, helmRelease)
		}
		state.List = append(state.List, helmReleaseList...)
	}
	state.Id = types.StringValue(common.DataSource + common.HelmReleaseList)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataPlaneHelmReleaseDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
