package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/service_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &backupDataSource{}
	_ datasource.DataSourceWithConfigure = &backupDataSource{}
)

// NewInstanceTypesDataSource is a helper function to simplify the provider implementation.
func NewBackupDataSource() datasource.DataSource {
	return &backupDataSource{}
}

// instanceTypesDataSource is the data source implementation.
type backupDataSource struct {
	client *tdh.Client
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
	BackupTriggerType types.String   `tfsdk:"backup_trigger_type"`
	Provider          types.String   `tfsdk:"provider_name"`
	Region            types.String   `tfsdk:"region"`
	DataPlaneId       types.String   `tfsdk:"data_plane_id"`
	Metadata          BackupMetadata `tfsdk:"metadata"`
	// TODO add upgrade related fields
}

// clusterMetadataModel maps order item data.
type BackupMetadata struct {
	ClusterName    types.String `tfsdk:"cluster_name"`
	ClusterSize    types.String `tfsdk:"cluster_size"`
	BackupLocation types.String `tfsdk:"backup_location"`
}

type backupResponseModel struct {
	Id          types.String          `tfsdk:"id"`
	BackupList  []backupResourceModel `tfsdk:"backup_list"`
	ServiceType types.String          `tfsdk:"service_type"`
}

type mdsBackupResponse struct {
	Embedded struct {
		MDSBackupDTOs []backupResourceModel `json:"mdsBackupDTOes"`
	} `json:"_embedded"`
}

func (d backupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (d backupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all the backups available for the service type on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Backup ID",
				Optional:    true,
			},
			"service_type": schema.StringAttribute{
				Description: "Service Type of the cluster.",
				Required:    true,
			},
			"backup_list": schema.ListNestedAttribute{
				Description: "Backup List",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the Backup.",
							Optional:    true,
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
							Required:    true,
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
							},
						},
					},
				},
			},
		},
	}

}

func (d backupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	//TODO implement me
	var state backupResponseModel
	var backupList []backupResourceModel
	//tflog.Debug(ctx, "rabbitmq dto", map[string]interface{}{"dto": clusterList})
	tflog.Debug(ctx, "Service type list :", map[string]interface{}{"dto": state.ServiceType})
	query := controller.BackupQuery{
		ServiceType: state.ServiceType.ValueString(),
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if err := service_type.ValidateRoleType(state.ServiceType.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"invalid type",
			err.Error())
		return
	}

	backups, err := d.client.Controller.GetBackups(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Service Types",
			err.Error(),
		)
		return
	}

	if backups.Page.TotalPages > 1 {
		for i := 1; i <= backups.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalClusters, err := d.client.Controller.GetBackups(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH Clusters",
					err.Error(),
				)
				return
			}

			for _, mdsClusterDto := range *totalClusters.Get() {
				cluster := backupResourceModel{
					ID:   types.StringValue(mdsClusterDto.Id),
					Name: types.StringValue(mdsClusterDto.Name),

					ClusterId:         types.StringValue(mdsClusterDto.ClusterId),
					ClusterName:       types.StringValue(mdsClusterDto.ClusterName),
					ClusterVersion:    types.StringValue(mdsClusterDto.ClusterVersion),
					ServiceType:       types.StringValue(mdsClusterDto.ServiceType),
					Size:              types.StringValue(mdsClusterDto.Size),
					BackupTriggerType: types.StringValue(mdsClusterDto.BackupTriggerType),
					Provider:          types.StringValue(mdsClusterDto.Provider),
					Region:            types.StringValue(mdsClusterDto.Region),
					DataPlaneId:       types.StringValue(mdsClusterDto.DataPlaneId),
				}

				metadata := BackupMetadata{
					ClusterName:    types.StringValue(mdsClusterDto.Metadata.ClusterName),
					ClusterSize:    types.StringValue(mdsClusterDto.Metadata.ClusterSize),
					BackupLocation: types.StringValue(mdsClusterDto.Metadata.BackupLocation),
				}
				cluster.Metadata = metadata
				backupList = append(backupList, cluster)
			}
		}

		tflog.Debug(ctx, "rabbitmq dto", map[string]interface{}{"dto": backupList})
		state.BackupList = append(state.BackupList, backupList...)
	} else {
		for _, mdsClusterDto := range *backups.Get() {
			cluster := backupResourceModel{
				ID:   types.StringValue(mdsClusterDto.Id),
				Name: types.StringValue(mdsClusterDto.Name),

				ClusterId:         types.StringValue(mdsClusterDto.ClusterId),
				ClusterName:       types.StringValue(mdsClusterDto.ClusterName),
				ClusterVersion:    types.StringValue(mdsClusterDto.ClusterVersion),
				ServiceType:       types.StringValue(mdsClusterDto.ServiceType),
				Size:              types.StringValue(mdsClusterDto.Size),
				BackupTriggerType: types.StringValue(mdsClusterDto.BackupTriggerType),
				Provider:          types.StringValue(mdsClusterDto.Provider),
				Region:            types.StringValue(mdsClusterDto.Region),
				DataPlaneId:       types.StringValue(mdsClusterDto.DataPlaneId),
			}
			//backupList = append(backupList, cluster)
			metadata := BackupMetadata{
				ClusterName:    types.StringValue(mdsClusterDto.Metadata.ClusterName),
				ClusterSize:    types.StringValue(mdsClusterDto.Metadata.ClusterSize),
				BackupLocation: types.StringValue(mdsClusterDto.Metadata.BackupLocation),
			}
			cluster.Metadata = metadata

			tflog.Debug(ctx, "mdsClusterDto dto", map[string]interface{}{"dto": cluster})

			tflog.Debug(ctx, "mdsClusterDto dto", map[string]interface{}{"dto": mdsClusterDto})
			state.BackupList = append(state.BackupList, cluster)

		}
	}

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
