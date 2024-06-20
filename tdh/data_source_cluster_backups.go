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
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &backupDataSource{}
	_ datasource.DataSourceWithConfigure = &backupDataSource{}
)

// NewBackupDataSource is a helper function to simplify the provider implementation.
func NewBackupDataSource() datasource.DataSource {
	return &backupDataSource{}
}

// instanceTypesDataSource is the data source implementation.
type backupDataSource struct {
	client *tdh.Client
}

type backupResponseModel struct {
	Id          types.String          `tfsdk:"id"`
	ServiceType types.String          `tfsdk:"service_type"`
	List        []backupResourceModel `tfsdk:"list"`
}

// backupResourceModel maps the resource schema data.
type backupResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	ClusterId         types.String   `tfsdk:"cluster_id"`
	ClusterName       types.String   `tfsdk:"cluster_name"`
	ClusterVersion    types.String   `tfsdk:"cluster_version"`
	ServiceType       types.String   `tfsdk:"service_type"`
	Size              types.String   `tfsdk:"size"`
	Status            types.String   `tfsdk:"status"`
	BackupTriggerType types.String   `tfsdk:"backup_trigger_type"`
	Provider          types.String   `tfsdk:"provider_name"`
	Region            types.String   `tfsdk:"region"`
	DataPlaneId       types.String   `tfsdk:"data_plane_id"`
	Metadata          BackupMetadata `tfsdk:"metadata"`
}

// BackupMetadata maps order item data.
type BackupMetadata struct {
	ClusterName    types.String              `tfsdk:"cluster_name"`
	ClusterSize    types.String              `tfsdk:"cluster_size"`
	BackupLocation types.String              `tfsdk:"backup_location"`
	Databases      []string                  `tfsdk:"databases"`
	Extensions     []BackupMetadataExtension `tfsdk:"extensions"`
}

type BackupMetadataExtension struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

func (d *backupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_backups"
}

func (d *backupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all the backups available for the service type on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Backup ID",
				Optional:    true,
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of the service. Supported values: %s .", supportedDataServiceTypesMarkdown()),
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("POSTGRES", "MYSQL", "REDIS"),
				},
			},
			"list": schema.ListNestedAttribute{
				Description: "List of the backups.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the Backup.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the Backup.",
							Computed:    true,
						},
						"cluster_id": schema.StringAttribute{
							Description: "ID of the cluster.",
							Computed:    true,
						},
						"cluster_name": schema.StringAttribute{
							Description: "Name of the cluster.",
							Computed:    true,
						},
						"cluster_version": schema.StringAttribute{
							Description: "Version of the cluster.",
							Computed:    true,
						},
						"service_type": schema.StringAttribute{
							Description: "Service Type of the cluster.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the backup.",
							Computed:    true,
						},
						"size": schema.StringAttribute{
							Description: "Size of the cluster.",
							Computed:    true,
						},
						"backup_trigger_type": schema.StringAttribute{
							Description: "The type of trigger for the backup.",
							Computed:    true,
						},
						"provider_name": schema.StringAttribute{
							Description: "The provider of the cluster.",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Description: "The region of the cluster.",
							Computed:    true,
						},
						"data_plane_id": schema.StringAttribute{
							Description: "The ID of the data plane.",
							Computed:    true,
						},
						"metadata": schema.SingleNestedAttribute{
							Description: "The metadata of the backup.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"cluster_name": schema.StringAttribute{
									Description: "Name of the cluster.",
									Computed:    true,
								},
								"cluster_size": schema.StringAttribute{
									Description: "Size of the Instance Type.",
									Computed:    true,
								},
								"backup_location": schema.StringAttribute{
									Description: "Backup Location",
									Computed:    true,
								},
								"databases": schema.SetAttribute{
									Description: "List of databases part of backup.",
									Computed:    true,
									ElementType: types.StringType,
								},
								"extensions": schema.SetNestedAttribute{
									MarkdownDescription: "List of extensions part of backup. *(Specific to service `POSTGRES`)*",
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
						},
					},
				},
			},
		},
	}

}

func (d *backupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state backupResponseModel
	var tfModels []backupResourceModel

	tflog.Debug(ctx, "Service type:", map[string]interface{}{"dto": state.ServiceType})
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	query := controller.BackupsQuery{
		ServiceType: state.ServiceType.ValueString(),
	}

	backups, err := d.client.Controller.GetClusterBackups(&query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Service Backups",
			err.Error(),
		)
		return
	}

	if backups.Page.TotalPages > 1 {
		for i := 1; i <= backups.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			backups, err := d.client.Controller.GetClusterBackups(&query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read Service Backups",
					err.Error(),
				)
				return
			}
			tfModels = append(tfModels, d.convertToTfModels(&ctx, backups.Get())...)
		}
		tflog.Debug(ctx, "READING tfModels", map[string]interface{}{"models": tfModels})
	} else {
		tfModels = d.convertToTfModels(&ctx, backups.Get())
	}
	state.List = append(state.List, tfModels...)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *backupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}

func (d *backupDataSource) convertToTfModels(ctx *context.Context, backupDtoList *[]model.ClusterBackup) []backupResourceModel {
	tflog.Info(*ctx, "converting to tfModels")
	var tfModels []backupResourceModel
	for _, mdsClusterDTO := range *backupDtoList {
		tfModel := backupResourceModel{
			ID:                types.StringValue(mdsClusterDTO.Id),
			Name:              types.StringValue(mdsClusterDTO.Name),
			ClusterId:         types.StringValue(mdsClusterDTO.ClusterId),
			ClusterName:       types.StringValue(mdsClusterDTO.ClusterName),
			ClusterVersion:    types.StringValue(mdsClusterDTO.ClusterVersion),
			ServiceType:       types.StringValue(mdsClusterDTO.ServiceType),
			Size:              types.StringValue(mdsClusterDTO.Size),
			Status:            types.StringValue(mdsClusterDTO.Status),
			BackupTriggerType: types.StringValue(mdsClusterDTO.BackupTriggerType),
			Provider:          types.StringValue(mdsClusterDTO.Provider),
			Region:            types.StringValue(mdsClusterDTO.Region),
			DataPlaneId:       types.StringValue(mdsClusterDTO.DataPlaneId),
		}
		metadata := BackupMetadata{
			ClusterName:    types.StringValue(mdsClusterDTO.Metadata.ClusterName),
			ClusterSize:    types.StringValue(mdsClusterDTO.Metadata.ClusterSize),
			BackupLocation: types.StringValue(mdsClusterDTO.Metadata.BackupLocation),
			Databases:      mdsClusterDTO.Metadata.Databases,
		}
		for _, extension := range mdsClusterDTO.Metadata.PostgresExtensions {
			metadataExtension := BackupMetadataExtension{
				Name:    types.StringValue(extension.Name),
				Version: types.StringValue(extension.Version),
			}
			metadata.Extensions = append(metadata.Extensions, metadataExtension)
		}

		tfModel.Metadata = metadata

		tflog.Debug(*ctx, "dto", map[string]interface{}{"dto": tfModel})
		tfModels = append(tfModels, tfModel)
	}
	tflog.Info(*ctx, "converted to tfModels")
	return tfModels
}
