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
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &restoreDataSource{}
	_ datasource.DataSourceWithConfigure = &restoreDataSource{}
)

// NewRestoresDataSource is a helper function to simplify the provider implementation.
func NewRestoreDataSource() datasource.DataSource {
	return &restoreDataSource{}
}

// restoreDataSource is the data source implementation.
type restoreDataSource struct {
	client *tdh.Client
}

// localUsersDataSourceModel maps the datasource schema
type restoreDataSourceModel struct {
	Id   types.String           `tfsdk:"id"`
	List []restoreResponseModel `tfsdk:"list"`
}

type restoreResponseModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	DataPlaneId        types.String `tfsdk:"dataplane_id"`
	ServiceType        types.String `tfsdk:"service_type"`
	BackupId           types.String `tfsdk:"backup_id"`
	BackupName         types.String `tfsdk:"backup_name"`
	TargetInstance     types.String `tfsdk:"target_instance"`
	TargetInstanceName types.String `tfsdk:"target_instance_name"`
}

//// localUsersDataSourceModel maps the datasource schema
//type localUsersDataSourceModel struct {
//	Id       types.String     `tfsdk:"id"`
//	Username types.String     `tfsdk:"username"`
//	List     []localUserModel `tfsdk:"list"`
//}
//
//type localUserModel struct {
//	Id        types.String `tfsdk:"id"`
//	Username  types.String `tfsdk:"username"`
//	PolicyIds types.Set    `tfsdk:"policy_ids"`
//}
//type restoreResourceModel struct {
//	Id                 types.String `tfsdk:"id"`
//	Name               types.String `tfsdk:"name"`
//	DataPlaneId        types.String `tfsdk:"dataplane_id"`
//	ServiceType        types.String `tfsdk:"service_type"`
//	BackupId           types.String `tfsdk:"backup_id"`
//	BackupName         types.String `tfsdk:"backup_name"`
//	TargetInstance     types.String `tfsdk:"target_instance"`
//	TargetInstanceName types.String `tfsdk:"target_instance_name"`
//}
//
//type restoreResponseModel struct {
//	Id                 types.String           `tfsdk:"id"`
//	Name               types.String           `tfsdk:"name"`
//	DataPlaneId        types.String           `tfsdk:"dataplane_id"`
//	ServiceType        types.String           `tfsdk:"service_type"`
//	BackupId           types.String           `tfsdk:"backup_id"`
//	BackupName         types.String           `tfsdk:"backup_name"`
//	TargetInstance     types.String           `tfsdk:"target_instance"`
//	TargetInstanceName types.String           `tfsdk:"target_instance_name"`
//	RestoreList        []restoreResourceModel `tfsdk:"restore_list"`
//}

// restoreResourceMode maps the resource schema data.
func (d restoreDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_restores"
}

func (d restoreDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all the backups available for the service type on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the task.",
			},
			"list": schema.ListNestedAttribute{
				Description: "List of restores.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Backup ID",
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
						"dataplane_id": schema.StringAttribute{
							Description: "Dataplane ID",
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

func (d restoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	//TODO implement me
	var state restoreDataSourceModel
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

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Configure adds the provider configured client to the data source.
func (d *restoreDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
