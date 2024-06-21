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
	_ datasource.DataSource              = &serviceExtensionsDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceExtensionsDataSource{}
)

type serviceExtensionsDataSourceModel struct {
	Id          types.String            `tfsdk:"id"`
	ServiceType types.String            `tfsdk:"service_type"`
	List        []serviceExtensionModel `tfsdk:"list"`
}

type serviceExtensionModel struct {
	Id          types.String       `tfsdk:"id"`
	Name        types.String       `tfsdk:"name"`
	Description types.String       `tfsdk:"description"`
	ServiceType types.String       `tfsdk:"service_type"`
	Version     types.String       `tfsdk:"version"`
	Metadata    *extensionMetadata `tfsdk:"metadata"`
}

type extensionMetadata struct {
	Port types.String `tfsdk:"port"`
}

func NewServiceExtensionsDataSource() datasource.DataSource {
	return &serviceExtensionsDataSource{}
}

type serviceExtensionsDataSource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *serviceExtensionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_extensions"
}

func (d *serviceExtensionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to get extensions for a service supported by TDH. Can be used while creating resource `tdh_cluster` of that.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource.",
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of the service. Supported values: %s .", supportedServiceTypesMarkdown()),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("POSTGRES", "MYSQL", "RABBITMQ", "REDIS"),
				},
			},
			"list": schema.ListNestedAttribute{
				Description: "List of available service versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the extension.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the extension.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the extension.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Version of the extension.",
							Computed:    true,
						},
						"service_type": schema.StringAttribute{
							Description: "Service type supporting this extension.",
							Computed:    true,
						},
						"metadata": schema.SingleNestedAttribute{
							Description: "The metadata of the extension.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"port": schema.StringAttribute{
									Description: "Port of the extension.",
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

// Configure adds the provider configured client to the data source.
func (d *serviceExtensionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *serviceExtensionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "INIT__READ")
	var state serviceExtensionsDataSourceModel

	//Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	tflog.Info(ctx, "READ tfState")

	query := controller.ServiceExtensionsQuery{}
	if !state.ServiceType.IsNull() {
		query.ServiceType = state.ServiceType.ValueString()
	}
	response, err := d.client.Controller.GetServiceExtensions(&query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Service extensions:",
			err.Error(),
		)
		return
	}
	for _, ext := range *response.Get() {
		model := serviceExtensionModel{
			Id:          types.StringValue(ext.Id),
			Name:        types.StringValue(ext.Name),
			Description: types.StringValue(ext.Description),
			ServiceType: types.StringValue(ext.ServiceType),
			Version:     types.StringValue(ext.Version),
			Metadata: &extensionMetadata{
				Port: types.StringValue(ext.Metadata.Port),
			},
		}
		state.List = append(state.List, model)
	}
	state.Id = types.StringValue(common.DataSource + common.ServiceVersionsId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
