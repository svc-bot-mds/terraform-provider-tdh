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
	_ datasource.DataSource              = &k8sClustersDatasource{}
	_ datasource.DataSourceWithConfigure = &k8sClustersDatasource{}
)

// K8sClustersDatasourceModel maps the data source schema data.
type K8sClustersDatasourceModel struct {
	Id             types.String `tfsdk:"id"`
	CloudAccountId types.String `tfsdk:"account_id"`
	List           []K8sCluster `tfsdk:"list"`
}

type K8sCluster struct {
	Name      types.String `tfsdk:"name"`
	Available types.Bool   `tfsdk:"available"`
	CpPresent types.Bool   `tfsdk:"cp_present"`
}

// NewK8sClustersDatasource is a helper function to simplify the provider implementation.
func NewK8sClustersDatasource() datasource.DataSource {
	return &k8sClustersDatasource{}
}

// rolesDatasource is the data source implementation.
type k8sClustersDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *k8sClustersDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_k8s_clusters"

}

// Schema defines the schema for the data source.
func (d *k8sClustersDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all Kubernetes Clusters within an provider account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource.",
			},
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the provider account",
			},
			"list": schema.ListNestedAttribute{
				Description: "List of Kubernetes Clusters. Can be used while creating resource `tdh_data_plane`.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the Kubernetes Cluster.",
							Computed:    true,
						},
						"available": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Denotes if the Kubernetes Cluster is available for onboarding or not.",
						},
						"cp_present": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Denotes if Control Plane is available. Use the Kubernetes Cluster with this set to `true` while onboarding Data Plane on TDH Control Plane.",
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *k8sClustersDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state K8sClustersDatasourceModel
	var clusterList []K8sCluster
	tflog.Info(ctx, "INIT__READ")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	clusters, err := d.client.InfraConnector.GetAccountClusters(state.CloudAccountId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Kubernetes Clusters",
			err.Error(),
		)
		return
	}
	for _, dto := range clusters {
		tfModel := K8sCluster{
			Name:      types.StringValue(dto.Name),
			Available: types.BoolValue(dto.IsAvailable),
			CpPresent: types.BoolValue(dto.IsCpPresent),
		}
		clusterList = append(clusterList, tfModel)
	}
	state.List = append(state.List, clusterList...)
	state.Id = types.StringValue(common.DataSource + common.K8sCluster)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *k8sClustersDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
