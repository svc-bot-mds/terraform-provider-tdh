package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &clusterTargetVersionsDataSource{}
	_ datasource.DataSourceWithConfigure = &clusterTargetVersionsDataSource{}
)

type clusterTargetVersionsDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	CurrentVersion types.String `tfsdk:"current_version"`
	List           []string     `tfsdk:"list"`
}

func NewClusterTargetVersionsDataSource() datasource.DataSource {
	return &clusterTargetVersionsDataSource{}
}

type clusterTargetVersionsDataSource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *clusterTargetVersionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_target_versions"
}

func (d *clusterTargetVersionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch target versions for a given service cluster. *(Can be used while upgrading resource `tdh_cluster`)*",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the cluster.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"current_version": schema.StringAttribute{
				Description: "Current version of the cluster.",
				Computed:    true,
			},
			"list": schema.SetAttribute{
				Description: "List of supported versions eligible for upgrade.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *clusterTargetVersionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *clusterTargetVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "INIT__READ")
	var state clusterTargetVersionsDataSourceModel

	//Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	tflog.Info(ctx, "READ tfState")

	response, err := d.client.UpgradeService.GetClusterTargetVersions(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Cluster Target Versions:",
			err.Error(),
		)
		return
	}
	state.List = append(state.List, response.TargetVersions...)
	state.CurrentVersion = types.StringValue(response.Version)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
