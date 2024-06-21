package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	service_metadata "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/service-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &networkPortsDataSource{}
	_ datasource.DataSourceWithConfigure = &networkPortsDataSource{}
)

// networkPortsDatasource maps the datasource schema
type networkPortsDataSourceModel struct {
	List        []networkPortsModel `tfsdk:"list"`
	Id          types.String        `tfsdk:"id"`
	ServiceType types.String        `tfsdk:"service_type"`
}

type networkPortsModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
	Port        types.Int64  `tfsdk:"port"`
}

func NewNetworkPortsDataSource() datasource.DataSource {
	return &networkPortsDataSource{}
}

type networkPortsDataSource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *networkPortsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_ports"
}

// Schema defines the schema for the data source.
func (d *networkPortsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all network ports supported for services on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource.",
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of the service to filter ports by. Supported values: %s .", supportedServiceTypesMarkdown()),
				Optional:            true,
			},
			"list": schema.ListNestedAttribute{
				Description: "List of network ports.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the port.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the port.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the port.",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The port number.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *networkPortsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networkPortsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	var query = &service_metadata.NetworkPortsQuery{}
	if !state.ServiceType.IsNull() {
		query.Type = state.ServiceType.ValueString()
	}

	networkPorts, err := d.client.ServiceMetadata.GetNetworkPorts(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Service Network ports",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "fetched ports", map[string]interface{}{"ports": networkPorts})
	// Map networkPorts body to model
	for _, networkPort := range networkPorts {
		networkPortsState := networkPortsModel{
			ID:          types.StringValue(networkPort.ID),
			Description: types.StringValue(networkPort.Description),
			Name:        types.StringValue(networkPort.Name),
			Port:        types.Int64Value(networkPort.Port),
		}

		tflog.Debug(ctx, "converted tfModel", map[string]interface{}{"model": networkPortsState})
		state.List = append(state.List, networkPortsState)
	}

	state.Id = types.StringValue(common.DataSource + common.NetworkPortsId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *networkPortsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
