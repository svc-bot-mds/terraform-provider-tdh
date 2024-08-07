package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/role_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/service-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &rolesDatasource{}
	_ datasource.DataSourceWithConfigure = &rolesDatasource{}
)

// rolesDataSourceModel maps the data source schema data.
type rolesDataSourceModel struct {
	List []rolesModel `tfsdk:"list"`
	Id   types.String `tfsdk:"id"`
}

// rolesModel maps role schema data.
type rolesModel struct {
	RoleId      types.String `tfsdk:"role_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// NewRolesDataSource is a helper function to simplify the provider implementation.
func NewRolesDataSource() datasource.DataSource {
	return &rolesDatasource{}
}

// rolesDatasource is the data source implementation.
type rolesDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *rolesDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

// Schema defines the schema for the data source.
func (d *rolesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all roles applicable on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"list": schema.ListNestedAttribute{
				Description: "List of roles.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role_id": schema.StringAttribute{
							Description: "ID of the role.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the role.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the role.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *rolesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rolesDataSourceModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &service_metadata.RolesQuery{
		Type: role_type.TDH,
	}
	rolesResponse, err := d.client.ServiceMetadata.GetRoles(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH roles",
			err.Error(),
		)
		return
	}

	// Extract the roles from the unmarshalled struct
	for _, role := range rolesResponse.Embedded.ServiceRoleDTO[0].Roles {
		// list only customer role similar to ui
		if role.RoleMini.Type == "customer" {
			roleList := rolesModel{
				RoleId:      types.StringValue(role.RoleID),
				Name:        types.StringValue(role.Name),
				Description: types.StringValue(role.Description),
			}
			state.List = append(state.List, roleList)
		}
	}
	state.Id = types.StringValue(common.DataSource + common.RolesId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *rolesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
