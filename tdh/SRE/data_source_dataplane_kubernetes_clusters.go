package SRE

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &dataplaneKubernetesClusterDatasource{}
	_ datasource.DataSourceWithConfigure = &dataplaneKubernetesClusterDatasource{}
)

// DnsDataSourceModel maps the data source schema data.
type DataplaneKubernetesClusterDatasourceModel struct {
	Id             types.String `tfsdk:"id"`
	CloudAccountId types.String `tfsdk:"account_id"`
	TkcList        []TkcList    `tfsdk:"tkc""`
}
type TkcList struct {
	Name        types.String `tfsdk:"name"`
	IsAvailable types.Bool   `tfsdk:"is_available"`
}

// NewDnsDatasource is a helper function to simplify the provider implementation.
func NewDataplaneKubernetesClusterDatasource() datasource.DataSource {
	return &dataplaneKubernetesClusterDatasource{}
}

// rolesDatasource is the data source implementation.
type dataplaneKubernetesClusterDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *dataplaneKubernetesClusterDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_clusters"

}

// Schema defines the schema for the data source.
func (d *dataplaneKubernetesClusterDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all Helm Releases for the Dataplane.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "Id of the cloud provider account",
			},
			"tkc": schema.ListNestedAttribute{
				Description: "List of Kubernetes Clusters. Id used while onboarding the dataplane",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the Kubernetes Cluster",
							Computed:    true,
						},
						"is_available": schema.BoolAttribute{
							Computed:    true,
							Description: "Denotes if the Kubernetes Cluster is enabled/not . Use the Kubernetes Cluster with the flag set to true while onboarding the dataplane",
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dataplaneKubernetesClusterDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataplaneKubernetesClusterDatasourceModel
	var tkcLists []TkcList
	tflog.Info(ctx, "INIT -- READ service roles")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	tkcList, err := d.client.InfraConnector.GetTkcList(state.CloudAccountId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Kubernetes Clusters",
			err.Error(),
		)
		return
	}
	for _, tkcs := range tkcList {
		tkc := TkcList{
			Name:        types.StringValue(tkcs.Name),
			IsAvailable: types.BoolValue(tkcs.IsAvailable),
		}
		tkcLists = append(tkcLists, tkc)
	}
	state.TkcList = append(state.TkcList, tkcLists...)
	state.Id = types.StringValue(common.DataSource + common.TKC)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *dataplaneKubernetesClusterDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
