package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &clusterVersionsDataSource{}
	_ datasource.DataSourceWithConfigure = &clusterVersionsDataSource{}
)

type clusterVersionsDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	ServiceType types.String `tfsdk:"service_type"`
	Provider    types.String `tfsdk:"provider_type"`
	List        []string     `tfsdk:"list"`
}

func NewClusterVersionsDataSource() datasource.DataSource {
	return &clusterVersionsDataSource{}
}

type clusterVersionsDataSource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *clusterVersionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	// TODO: rename it to "_service_versions" after changing it in MDS UI example
	resp.TypeName = req.ProviderTypeName + "_cluster_versions"
}

func (d *clusterVersionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch versions for service cluster supported by TDH. Can be used while creating resource `tdh_cluster`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource.",
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of the service. Supported values: %s .", supportedServiceTypesMarkdown()),
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("POSTGRES", "MYSQL", "RABBITMQ", "REDIS"),
				},
			},
			"provider_type": schema.StringAttribute{
				MarkdownDescription: "Short code of provider. Ex: `tkgs`, `tkgm`, `openshift` etc.",
				Required:            true,
			},
			"list": schema.SetAttribute{
				Description: "List of available cluster versions.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *clusterVersionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *clusterVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "INIT__READ cluster_versions")
	var state clusterVersionsDataSourceModel

	//Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	tflog.Info(ctx, "READ tfState")

	query := controller.ServiceVersionsQuery{
		ServiceType:  state.ServiceType.ValueString(),
		Provider:     state.Provider.ValueString(),
		TemplateType: "CLUSTER",
		Action:       "CREATE",
	}
	list, err := d.client.Controller.GetServiceVersions(&query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Cluster versions:",
			err.Error(),
		)
		return
	}
	state.Id = types.StringValue(common.DataSource + common.ClusterVersions)
	state.List = append(state.List, list...)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
