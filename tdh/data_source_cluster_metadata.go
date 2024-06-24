package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/service_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/validators"
)

var (
	_ datasource.DataSource              = &clusterMetadataDataSource{}
	_ datasource.DataSourceWithConfigure = &clusterMetadataDataSource{}
)

type clusterMetadataDataSourceModel struct {
	Id           types.String     `tfsdk:"id"`
	ProviderName types.String     `tfsdk:"provider_name"`
	Name         types.String     `tfsdk:"name"`
	ServiceType  types.String     `tfsdk:"service_type"`
	Status       types.String     `tfsdk:"status"`
	Vhosts       []vHostModel     `tfsdk:"vhosts"`
	Queues       []queueModel     `tfsdk:"queues"`
	Exchanges    []exchangeModel  `tfsdk:"exchanges"`
	Bindings     []bindingModel   `tfsdk:"bindings"`
	Databases    []databaseModel  `tfsdk:"databases"`
	Extensions   []extensionModel `tfsdk:"extensions"`
}

type vHostModel struct {
	Name types.String `tfsdk:"name"`
}

type queueModel struct {
	Name  types.String `tfsdk:"name"`
	VHost types.String `tfsdk:"vhost"`
}

type exchangeModel struct {
	Name  types.String `tfsdk:"name"`
	VHost types.String `tfsdk:"vhost"`
}

type bindingModel struct {
	Source          types.String `tfsdk:"source"`
	VHost           types.String `tfsdk:"vhost"`
	RoutingKey      types.String `tfsdk:"routing_key"`
	Destination     types.String `tfsdk:"destination"`
	DestinationType types.String `tfsdk:"destination_type"`
}

type databaseModel struct {
	Name     types.String  `tfsdk:"name"`
	Owner    types.String  `tfsdk:"owner"`
	Schemas  []schemaModel `tfsdk:"schemas"`
	Tables   []tableModel  `tfsdk:"tables"`
	Routines []string      `tfsdk:"routines"`
}

type schemaModel struct {
	Name   types.String `tfsdk:"name"`
	Owner  types.String `tfsdk:"owner"`
	Tables []tableModel `tfsdk:"tables"`
}

type tableModel struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Columns []string     `tfsdk:"columns"`
}

type extensionModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

func NewClusterMetadataDataSource() datasource.DataSource {
	return &clusterMetadataDataSource{}
}

type clusterMetadataDataSource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *clusterMetadataDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_metadata"
}

// Schema defines the schema for the data source.
func (d *clusterMetadataDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch metadata of a cluster by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the cluster.",
				Required:    true,
				Validators: []validator.String{
					validators.UUIDValidator{},
				},
			},
			"provider_name": schema.StringAttribute{
				Description: "Name of the data-plane's cloud provider where cluster is deployed.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Computed:    true,
			},
			"service_type": schema.StringAttribute{
				Description: "Type of the service of the cluster.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the cluster.",
				Computed:    true,
			},
			"vhosts": schema.ListNestedAttribute{
				MarkdownDescription: "List of the vHosts. *(Specific to `RABBITMQ` service)*",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the vHost.",
							Computed:    true,
						},
					},
				},
			},
			"queues": schema.ListNestedAttribute{
				MarkdownDescription: "List of the Queues. *(Specific to `RABBITMQ` service)*",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the queue.",
							Computed:    true,
						},
						"vhost": schema.StringAttribute{
							Description: "vHost of the queue.",
							Computed:    true,
						},
					},
				},
			},
			"exchanges": schema.ListNestedAttribute{
				MarkdownDescription: "List of the Exchanges. *(Specific to `RABBITMQ` service)*",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the exchange.",
							Computed:    true,
						},
						"vhost": schema.StringAttribute{
							Description: "vHost of the exchange.",
							Computed:    true,
						},
					},
				},
			},
			"bindings": schema.ListNestedAttribute{
				MarkdownDescription: "List of the Bindings. *(Specific to `RABBITMQ` service)*",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "Source exchange.",
							Computed:    true,
						},
						"vhost": schema.StringAttribute{
							Description: "vHost name.",
							Computed:    true,
						},
						"routing_key": schema.StringAttribute{
							Description: "Routing key.",
							Computed:    true,
						},
						"destination": schema.StringAttribute{
							Description: "Destination exchange.",
							Computed:    true,
						},
						"destination_type": schema.StringAttribute{
							Description: "Type of the destination.",
							Computed:    true,
						},
					},
				},
			},
			"databases": schema.ListNestedAttribute{
				MarkdownDescription: "List of the Databases. *(Specific to `POSTGRES` & `MYSQL` service)*",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the database.",
							Computed:    true,
						},
						"owner": schema.StringAttribute{
							Description: "Name of the database owner.",
							Computed:    true,
						},
						"schemas": schema.ListNestedAttribute{
							MarkdownDescription: "List of the schemas in the database. *(Specific to `POSTGRES` service)*",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the schema.",
										Computed:    true,
									},
									"owner": schema.StringAttribute{
										Description: "Name of the schema owner.",
										Computed:    true,
									},
									"tables": schema.ListNestedAttribute{
										Description: "List of the tables in the schema.",
										Computed:    true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Description: "Name of the table.",
													Computed:    true,
												},
												"type": schema.StringAttribute{
													Description: "Type of the table.",
													Computed:    true,
												},
												"columns": schema.SetAttribute{
													Description: "List of the columns in the table.",
													Computed:    true,
													ElementType: types.StringType,
												},
											},
										},
									},
								},
							},
						},
						"tables": schema.ListNestedAttribute{
							MarkdownDescription: "List of the tables in the database. *(Specific to `MYSQL` service)*",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the table.",
										Computed:    true,
									},
									"type": schema.StringAttribute{
										Description: "Type of the table.",
										Computed:    true,
									},
									"columns": schema.SetAttribute{
										Description: "List of the columns in the table.",
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"routines": schema.SetAttribute{
							MarkdownDescription: "List of the routines in the database. *(Specific to `MYSQL` service)*",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"extensions": schema.ListNestedAttribute{
				MarkdownDescription: "List of the extensions in cluster. *(Specific to `POSTGRES` service)*",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the extension.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Version of the extension.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *clusterMetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	var stateModel clusterMetadataDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &stateModel)...)
	clusterMetadata, err := d.client.Controller.GetClusterMetaData(stateModel.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Cluster Metadata",
			err.Error(),
		)
		return
	}
	tflog.Info(ctx, "fetched metadata", map[string]interface{}{"metadata": &clusterMetadata})

	// Map cluster metadata body to model
	stateModel = clusterMetadataDataSourceModel{
		Id:           types.StringValue(clusterMetadata.Id),
		Name:         types.StringValue(clusterMetadata.Name),
		ProviderName: types.StringValue(clusterMetadata.Provider),
		ServiceType:  types.StringValue(clusterMetadata.ServiceType),
		Status:       types.StringValue(clusterMetadata.Status),
	}

	switch stateModel.ServiceType.ValueString() {
	case service_type.RABBITMQ:
		d.populateForRabbitmq(&ctx, &stateModel, clusterMetadata)
		break
	case service_type.POSTGRES:
		d.populateForPostgres(&ctx, &stateModel, clusterMetadata)
		break
	case service_type.MYSQL:
		d.populateForMysql(&ctx, &stateModel, clusterMetadata)
		break
	}
	tflog.Debug(ctx, "before setting state", map[string]interface{}{
		"state": stateModel,
	})
	// Set state
	diags := resp.State.Set(ctx, &stateModel)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *clusterMetadataDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}

func (d *clusterMetadataDataSource) populateForRabbitmq(ctx *context.Context, state *clusterMetadataDataSourceModel, response *model.ClusterMetaData) {
	tflog.Debug(*ctx, "populating vHosts")
	if len(response.VHosts) > 0 {
		vhosts := make([]vHostModel, 0)
		// Map response body to model
		for _, vhost := range response.VHosts {
			vhost := vHostModel{
				Name: types.StringValue(vhost.Name),
			}
			vhosts = append(vhosts, vhost)
		}
		state.Vhosts = vhosts
	}
	tflog.Debug(*ctx, "populating bindings")
	if len(response.Bindings) > 0 {
		bindings := make([]bindingModel, 0)
		for _, binding := range response.Bindings {
			binding := bindingModel{
				Source:          types.StringValue(binding.Source),
				DestinationType: types.StringValue(binding.DestinationType),
				Destination:     types.StringValue(binding.Destination),
				VHost:           types.StringValue(binding.VHost),
				RoutingKey:      types.StringValue(binding.RoutingKey),
			}
			bindings = append(bindings, binding)
		}
		state.Bindings = bindings
	}
	tflog.Debug(*ctx, "populating queues")
	if len(response.Queues) > 0 {
		queues := make([]queueModel, 0)
		for _, queue := range response.Queues {
			queue := queueModel{
				Name:  types.StringValue(queue.Name),
				VHost: types.StringValue(queue.VHost),
			}
			queues = append(queues, queue)
		}
		state.Queues = queues
	}
	tflog.Debug(*ctx, "populating exchanges")
	if len(response.Exchanges) > 0 {
		exchanges := make([]exchangeModel, 0)
		for _, exchange := range response.Exchanges {
			exchange := exchangeModel{
				Name:  types.StringValue(exchange.Name),
				VHost: types.StringValue(exchange.VHost),
			}
			exchanges = append(exchanges, exchange)
		}
		state.Exchanges = exchanges
	}
	tflog.Debug(*ctx, "populated rabbitmq metadata")
}

func (d *clusterMetadataDataSource) populateForPostgres(ctx *context.Context, state *clusterMetadataDataSourceModel, response *model.ClusterMetaData) {
	tflog.Debug(*ctx, "populating POSTGRES databases")
	if len(response.Databases) > 0 {
		dbModels := make([]databaseModel, 0)
		for _, databaseDto := range response.Databases {
			databaseTfModel := databaseModel{
				Name:  types.StringValue(databaseDto.Name),
				Owner: types.StringValue(databaseDto.Owner),
			}
			schemaModels := make([]schemaModel, 0)
			for _, schemaDto := range databaseDto.Schemas {
				schemaTfModel := schemaModel{
					Name:  types.StringValue(schemaDto.Name),
					Owner: types.StringValue(schemaDto.Owner),
				}
				tableModels := make([]tableModel, 0)
				for _, tableDto := range schemaDto.Tables {
					tableTfModel := tableModel{
						Name: types.StringValue(tableDto.Name),
						Type: types.StringValue(tableDto.Type),
					}
					tableModels = append(tableModels, tableTfModel)
				}
				tflog.Debug(*ctx, "CHECK TABLES: ", map[string]interface{}{
					"tables": schemaTfModel.Tables,
				})
				tflog.Debug(*ctx, "CHECK TABLES: ", map[string]interface{}{
					"table-models": tableModels,
				})
				schemaTfModel.Tables = tableModels
				schemaModels = append(schemaModels, schemaTfModel)
			}
			tflog.Debug(*ctx, "CHECK SCHEMAS: ", map[string]interface{}{
				"schemas": databaseTfModel.Schemas,
			})
			tflog.Debug(*ctx, "CHECK SCHEMAS: ", map[string]interface{}{
				"schema-models": schemaModels,
			})
			databaseTfModel.Schemas = schemaModels
			dbModels = append(dbModels, databaseTfModel)
		}
		state.Databases = dbModels
	}
	tflog.Debug(*ctx, "populating extensionsData")
	if len(response.PostgresExtensionData) > 0 {
		list := make([]extensionModel, 0)
		for _, extension := range response.PostgresExtensionData {
			exchange := extensionModel{
				Name:    types.StringValue(extension.Name),
				Version: types.StringValue(extension.Version),
			}
			list = append(list, exchange)
		}
		state.Extensions = list
	}
	tflog.Debug(*ctx, "populated POSTGRES metadata")
}

func (d *clusterMetadataDataSource) populateForMysql(ctx *context.Context, state *clusterMetadataDataSourceModel, response *model.ClusterMetaData) {
	tflog.Debug(*ctx, "populating MYSQL databases")
	if len(response.Databases) > 0 {
		dbModels := make([]databaseModel, 0)
		for _, databaseDto := range response.Databases {
			databaseTfModel := databaseModel{
				Name:     types.StringValue(databaseDto.Name),
				Owner:    types.StringValue(databaseDto.Owner),
				Routines: databaseDto.Routines,
			}
			tableModels := make([]tableModel, 0)
			for _, tableDto := range databaseDto.Tables {
				tableTfModel := tableModel{
					Name:    types.StringValue(tableDto.Name),
					Type:    types.StringValue(tableDto.Type),
					Columns: tableDto.Columns,
				}
				tableModels = append(tableModels, tableTfModel)
			}
			dbModels = append(dbModels, databaseTfModel)
		}
		state.Databases = dbModels
	}
	tflog.Debug(*ctx, "populated MYSQL metadata")
}
