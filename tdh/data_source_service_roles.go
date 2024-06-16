package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	service_metadata "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/service-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
)

var (
	_ datasource.DataSource              = &serviceRolesDatasource{}
	_ datasource.DataSourceWithConfigure = &serviceRolesDatasource{}
)

// ServiceRolesDataSourceModel maps the data source schema data.
type ServiceRolesDataSourceModel struct {
	Id   types.String        `tfsdk:"id"`
	List []ServiceRolesModel `tfsdk:"list"`
	Type types.String        `tfsdk:"type"`
}

// ServiceRolesModel maps role schema data.
type ServiceRolesModel struct {
	RoleId       types.String `tfsdk:"role_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	PermissionId types.String `tfsdk:"permission_id"`
}

// NewServiceRolesDatasource is a helper function to simplify the provider implementation.
func NewServiceRolesDatasource() datasource.DataSource {
	return &serviceRolesDatasource{}
}

// rolesDatasource is the data source implementation.
type serviceRolesDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *serviceRolesDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_roles"

}

// Schema defines the schema for the data source.
func (d *serviceRolesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all roles applicable for services on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"type": schema.StringAttribute{
				Description: "Type of the service on TDH.",
				Required:    true,
				Validators: []validator.String{
					utils.ServiceTypeValidator,
				},
			},
			"list": schema.ListNestedAttribute{
				Description: "List of service roles.",
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
						"type": schema.StringAttribute{
							Description: "Type of the role.",
							Computed:    true,
						},
						"permission_id": schema.StringAttribute{
							Description: "ID of the permission for the role.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *serviceRolesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ServiceRolesDataSourceModel
	tflog.Info(ctx, "INIT -- READ service roles")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &service_metadata.RolesQuery{
		Type: state.Type.ValueString(),
	}
	rolesResponse, err := d.client.ServiceMetadata.GetServiceRoles(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Service roles",
			err.Error(),
		)
		return
	}

	for _, role := range rolesResponse.Embedded.ServiceRoleDTO[0].Roles {
		roleList := ServiceRolesModel{
			RoleId:       types.StringValue(role.RoleID),
			Name:         types.StringValue(role.Name),
			Description:  types.StringValue(role.Description),
			Type:         types.StringValue(role.Type),
			PermissionId: types.StringValue(role.Permissions[0].PermissionId),
		}
		state.List = append(state.List, roleList)
	}
	state.Id = types.StringValue(common.DataSource + common.ServiceRolesId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *serviceRolesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
