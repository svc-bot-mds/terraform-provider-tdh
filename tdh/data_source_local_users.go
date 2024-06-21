package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &localUsersDataSource{}
	_ datasource.DataSourceWithConfigure = &localUsersDataSource{}
)

// localUsersDataSourceModel maps the datasource schema
type localUsersDataSourceModel struct {
	Id       types.String     `tfsdk:"id"`
	Username types.String     `tfsdk:"username"`
	List     []localUserModel `tfsdk:"list"`
}

type localUserModel struct {
	Id        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	PolicyIds types.Set    `tfsdk:"policy_ids"`
}

func NewLocalUsersDataSource() datasource.DataSource {
	return &localUsersDataSource{}
}

type localUsersDataSource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *localUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_users"
}

// Schema defines the schema for the data source.
func (d *localUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch local users present on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the task.",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the local user to filter results by.",
			},
			"list": schema.ListNestedAttribute{
				Description: "List of local users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the task.",
							Computed:    true,
						},
						"username": schema.StringAttribute{
							Description: "Name of the local user.",
							Computed:    true,
						},
						"policy_ids": schema.SetAttribute{
							Description: "IDs of the policies attached to this local user.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *localUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *localUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state localUsersDataSourceModel

	// Read Terraform configuration data into the model
	if resp.Diagnostics.Append(req.Config.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "populating by filters: %s", map[string]interface{}{"state": state})
	query := &customer_metadata.LocalUsersQuery{}
	if !state.Username.IsNull() {
		query.Username = state.Username.ValueString()
	}
	response, err := d.client.CustomerMetadata.GetLocalUsers(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Local Users(s)",
			err.Error(),
		)
		return
	}

	// Map DTO body to model
	state.List = make([]localUserModel, 0)
	for _, dto := range *response.Get() {
		tflog.Debug(ctx, "READING local-user", map[string]interface{}{
			"local-user": dto,
		})
		tfModel, hasError := d.convertToTfModel(&ctx, &resp.Diagnostics, dto)
		if hasError {
			return
		}
		state.List = append(state.List, *tfModel)
	}

	state.Id = types.StringValue(common.DataSource + common.LocalUsersId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
}

func (d *localUsersDataSource) convertToTfModel(ctx *context.Context, diag *diag.Diagnostics, response model.LocalUser) (*localUserModel, bool) {
	tfModel := localUserModel{
		Id:       types.StringValue(response.Id),
		Username: types.StringValue(response.Username),
	}
	policyIds, diags := types.SetValueFrom(*ctx, types.StringType, response.PolicyIds)
	if diag.Append(diags...); diag.HasError() {
		return nil, false
	}
	tfModel.PolicyIds = policyIds

	return &tfModel, false
}
