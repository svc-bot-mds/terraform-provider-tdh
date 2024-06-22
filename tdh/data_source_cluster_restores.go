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
	_ datasource.DataSource              = &restoresDataSource{}
	_ datasource.DataSourceWithConfigure = &restoresDataSource{}
)

// NewRestoresDataSource is a helper function to simplify the provider implementation.
func NewRestoresDataSource() datasource.DataSource {
	return &restoresDataSource{}
}

// restoresDataSource is the data source implementation.
type restoresDataSource struct {
	client *tdh.Client
}

// restoresDataSourceModel maps the datasource schema
type restoresDataSourceModel struct {
	Id          types.String           `tfsdk:"id"`
	ServiceType types.String           `tfsdk:"service_type"`
	List        []restoreResponseModel `tfsdk:"list"`
}

type restoreResponseModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	DataPlaneId        types.String `tfsdk:"data_plane_id"`
	ServiceType        types.String `tfsdk:"service_type"`
	BackupId           types.String `tfsdk:"backup_id"`
	BackupName         types.String `tfsdk:"backup_name"`
	TargetInstance     types.String `tfsdk:"target_instance"`
	TargetInstanceName types.String `tfsdk:"target_instance_name"`
}

// Metadata restoreResourceMode maps the resource schema data.
func (d *restoresDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_restores"
}

func (d *restoresDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all the backups available for the service type on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of the service. Supported values: %s .", supportedDataServiceTypesMarkdown()),
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("POSTGRES", "MYSQL", "REDIS"),
				},
			},
			"list": schema.ListNestedAttribute{
				Description: "List of restores.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Restore ID",
							Optional:    true,
						},
						"service_type": schema.StringAttribute{
							Description: "Service Type of the cluster.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("POSTGRES", "MYSQL", "REDIS"),
							},
						},
						"name": schema.StringAttribute{
							Description: "Name of the Backup.",
							Computed:    true,
						},
						"data_plane_id": schema.StringAttribute{
							Description: "Data plane ID",
							Computed:    true,
						},
						"backup_name": schema.StringAttribute{
							Description: "Backup name",
							Computed:    true,
						},
						"backup_id": schema.StringAttribute{
							Description: "Backup ID",
							Computed:    true,
						},
						"target_instance": schema.StringAttribute{
							Description: "Target instance",
							Computed:    true,
						},
						"target_instance_name": schema.StringAttribute{
							Description: "Target Instance Name",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *restoresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state restoresDataSourceModel
	var restoreList []restoreResponseModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := controller.RestoreQuery{}

	restore, err := d.client.Controller.GetClusterRestores(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Service Restore",
			err.Error(),
		)
		return
	}

	if restore.Page.TotalPages > 1 {
		for i := 1; i <= restore.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalClusters, err := d.client.Controller.GetClusterRestores(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read Service restore",
					err.Error(),
				)
				return
			}

			for _, mdsClusterDTO := range *totalClusters.Get() {
				cluster := restoreResponseModel{
					Id:                 types.StringValue(mdsClusterDTO.Id),
					Name:               types.StringValue(mdsClusterDTO.Name),
					ServiceType:        types.StringValue(mdsClusterDTO.ServiceType),
					DataPlaneId:        types.StringValue(mdsClusterDTO.DataPlaneId),
					BackupId:           types.StringValue(mdsClusterDTO.BackupId),
					BackupName:         types.StringValue(mdsClusterDTO.BackupName),
					TargetInstance:     types.StringValue(mdsClusterDTO.TargetInstance),
					TargetInstanceName: types.StringValue(mdsClusterDTO.TargetInstanceName),
				}

				restoreList = append(restoreList, cluster)
			}
		}

		tflog.Debug(ctx, "READING dto", map[string]interface{}{"dto": restoreList})
		state.List = append(state.List, restoreList...)
	} else {
		for _, mdsClusterDTO := range *restore.Get() {
			cluster := restoreResponseModel{
				Id:                 types.StringValue(mdsClusterDTO.Id),
				Name:               types.StringValue(mdsClusterDTO.Name),
				ServiceType:        types.StringValue(mdsClusterDTO.ServiceType),
				DataPlaneId:        types.StringValue(mdsClusterDTO.DataPlaneId),
				BackupId:           types.StringValue(mdsClusterDTO.BackupId),
				BackupName:         types.StringValue(mdsClusterDTO.BackupName),
				TargetInstance:     types.StringValue(mdsClusterDTO.TargetInstance),
				TargetInstanceName: types.StringValue(mdsClusterDTO.TargetInstanceName),
			}

			tflog.Debug(ctx, "mdsClusterDto dto", map[string]interface{}{"dto": cluster})

			state.List = append(state.List, cluster)

		}
	}

	state.Id = types.StringValue(common.DataSource + common.ClusterRestoresId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Configure adds the provider configured client to the data source.
func (d *restoresDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
