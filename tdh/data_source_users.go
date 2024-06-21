package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	customer_metadata "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &usersDatasource{}
	_ datasource.DataSourceWithConfigure = &usersDatasource{}
)

// instanceTypesDataSourceModel maps the data source schema data.
type usersDataSourceModel struct {
	Id    types.String `tfsdk:"id"`
	Users []userModel  `tfsdk:"users"`
}

type userModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

// NewUsersDataSource is a helper function to simplify the provider implementation.
func NewUsersDataSource() datasource.DataSource {
	return &usersDatasource{}
}

// networkPoliciesDatasource is the data source implementation.
type usersDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *usersDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

// Schema defines the schema for the data source.
func (d *usersDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch all users registered on TDH in an Org *(determined by the token used for provider)*.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"users": schema.ListNestedAttribute{
				Description: "List of users on TDH.",
				Computed:    true,
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the user.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the user.",
							Computed:    true,
							Optional:    true,
						},
						"email": schema.StringAttribute{
							Description: "Email of the user.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *usersDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel
	var userList []userModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &customer_metadata.UsersQuery{}

	users, err := d.client.CustomerMetadata.GetUsers(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH User Accounts",
			err.Error(),
		)
		return
	}

	if users.Page.TotalPages > 1 {
		for i := 1; i <= users.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalUsers, err := d.client.CustomerMetadata.GetUsers(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH User Accounts",
					err.Error(),
				)
				return
			}

			for _, userAccountDto := range *totalUsers.Get() {
				user := userModel{
					ID:    types.StringValue(userAccountDto.Id),
					Name:  types.StringValue(userAccountDto.Name),
					Email: types.StringValue(userAccountDto.Email),
				}
				userList = append(userList, user)
			}
		}

		tflog.Debug(ctx, "users dto", map[string]interface{}{"dto": userList})
		state.Users = append(state.Users, userList...)
	} else {
		for _, userAccountDto := range *users.Get() {
			user := userModel{
				ID:    types.StringValue(userAccountDto.Id),
				Name:  types.StringValue(userAccountDto.Name),
				Email: types.StringValue(userAccountDto.Email),
			}
			tflog.Debug(ctx, "converted userAccount dto", map[string]interface{}{"dto": user})
			state.Users = append(state.Users, user)
		}
	}

	state.Id = types.StringValue(common.DataSource + common.UsersId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *usersDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
