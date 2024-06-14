package SRE

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	infra_connector "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &dataplaneHelmReleaseDatasource{}
	_ datasource.DataSourceWithConfigure = &dataplaneHelmReleaseDatasource{}
)

// DnsDataSourceModel maps the data source schema data.
type DataplaneHelmReleaseDataSourceModel struct {
	Id              types.String      `tfsdk:"id"`
	HelmReleaseList []HelmReleaseList `tfsdk:"helm_release_list""`
}
type HelmReleaseList struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewDnsDatasource is a helper function to simplify the provider implementation.
func NewDataplaneHelmReleaseDatasource() datasource.DataSource {
	return &dataplaneHelmReleaseDatasource{}
}

// rolesDatasource is the data source implementation.
type dataplaneHelmReleaseDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *dataplaneHelmReleaseDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataplane_helm_release"

}

// Schema defines the schema for the data source.
func (d *dataplaneHelmReleaseDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all Helm Releases for the Dataplane.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"helm_release_list": schema.ListNestedAttribute{
				Description: "List of Helm Releases / Dataplane Versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the Helm Release",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the Helm Release",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dataplaneHelmReleaseDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataplaneHelmReleaseDataSourceModel
	var helmReleaseList []HelmReleaseList
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
				helmRelease := HelmReleaseList{
					Name: types.StringValue(helmDto.Name),
					Id:   types.StringValue(helmDto.Id),
				}
				helmReleaseList = append(helmReleaseList, helmRelease)
			}
		}

		state.HelmReleaseList = append(state.HelmReleaseList, helmReleaseList...)
	} else {
		for _, helmDto := range *helmVersions.Get() {
			dns := HelmReleaseList{
				Name: types.StringValue(helmDto.Name),
				Id:   types.StringValue(helmDto.Id),
			}
			helmReleaseList = append(helmReleaseList, dns)
		}
		state.HelmReleaseList = append(state.HelmReleaseList, helmReleaseList...)
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
func (d *dataplaneHelmReleaseDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
