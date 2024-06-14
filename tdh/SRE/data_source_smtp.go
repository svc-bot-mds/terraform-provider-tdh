package SRE

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &smtpDatasource{}
	_ datasource.DataSourceWithConfigure = &smtpDatasource{}
)

// SmtpDatasourceModel maps the data source schema data.
type SmtpDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	SmptpDetails []SmtpModel  `tfsdk:"smtp""`
}
type SmtpModel struct {
	AuthenticationEnabled types.String `tfsdk:"auth"`
	TlsEnabled            types.String `tfsdk:"tls"`
	FromEmail             types.String `tfsdk:"from_email"`
	UserName              types.String `tfsdk:"user_name"`
	Host                  types.String `tfsdk:"host"`
	Port                  types.String `tfsdk:"port"`
}

// NewSmtpDatasource is a helper function to simplify the provider implementation.
func NewSmtpDatasource() datasource.DataSource {
	return &smtpDatasource{}
}

// smtpDatasource is the data source implementation.
type smtpDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *smtpDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp"

}

// Schema defines the schema for the data source.
func (d *smtpDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch SMTP details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"smtp": schema.ListNestedAttribute{
				Description: "SMTP Details",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{

						"auth": schema.StringAttribute{
							Description: " Authentication Enabled ?",
							Computed:    true,
						},
						"tls": schema.StringAttribute{
							Description: " TLS Enabled",
							Computed:    true,
						},
						"from_email": schema.StringAttribute{
							Description: "SMTP email address",
							Computed:    true,
						},
						"user_name": schema.StringAttribute{
							Description: "SMTP user name",
							Computed:    true,
						},
						"host": schema.StringAttribute{
							Description: "SMTP Host",
							Computed:    true,
						},
						"port": schema.StringAttribute{
							Description: "SMTP Port",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *smtpDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SmtpDataSourceModel
	tflog.Info(ctx, "INIT -- READ service roles")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	smtp, err := d.client.Auth.GetSmtpDetails()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH SMTP Details",
			err.Error(),
		)
		return
	}

	smtpDetails := SmtpModel{
		AuthenticationEnabled: types.StringValue(smtp.Auth),
		TlsEnabled:            types.StringValue(smtp.Tls),
		FromEmail:             types.StringValue(smtp.FromEmail),
		UserName:              types.StringValue(smtp.UserName),
		Host:                  types.StringValue(smtp.Host),
		Port:                  types.StringValue(smtp.Port),
	}

	state.SmptpDetails = append(state.SmptpDetails, smtpDetails)
	state.Id = types.StringValue(common.DataSource + common.ServiceRolesId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *smtpDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
