package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/identity_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/service_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &tdhProvider{}
)

const (
	EnvHost     = "TDH_HOST"
	EnvUsername = "TDH_USERNAME"
	EnvPassword = "TDH_PASSWORD"
	EnvOrgId    = "TDH_ORG_ID"
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &tdhProvider{}
}

// tdhProvider is the provider implementation.
type tdhProvider struct{}

// tdhProviderModel maps provider schema data to a Go type.
type tdhProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Type     types.String `tfsdk:"type"`
	OrgId    types.String `tfsdk:"org_id"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// Metadata returns the provider type name.
func (p *tdhProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tdh"
}

// Schema defines the provider-level schema for configuration data.
func (p *tdhProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with VMware Tanzu Data Hub",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("URI for TDH API. *(may also be provided via `%s` environment variable)*", EnvHost),
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("OAuth Type for the TDH API. *(must be `%s` if used, attribute kept for backward compatibility)*", oauth_type.UserCredentials),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(oauth_type.UserCredentials),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Username for TDH API. *(may also be provided via `%s` environment variable)*", EnvUsername),
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Password for TDH API. *(may also be provided via `%s` environment variable)*", EnvPassword),
				Optional:            true,
				Sensitive:           true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Organization Id for TDH API. *(may also be provided via `%s` environment variable)*", EnvOrgId),
				Optional:            true,
			},
		},
	}
}

func (p *tdhProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring TDH client")

	// Retrieve provider data from configuration
	var config tdhProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv(EnvHost)
	orgId := os.Getenv(EnvOrgId)
	username := os.Getenv(EnvUsername)
	password := os.Getenv(EnvPassword)

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	//if config.Type.ValueString() == oauth_type.ApiToken {
	//	if !config.ApiToken.IsNull() {
	//		apiToken = config.ApiToken.ValueString()
	//	}
	//}
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}
	if !config.OrgId.IsNull() {
		orgId = config.OrgId.ValueString()
	}
	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing TDH API Host",
			"The provider cannot create the TDH API client as there is a missing or empty value for the TDH API host. "+
				fmt.Sprintf("Set the host value in the configuration or use the '%s' environment variable. ", EnvHost)+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("Username"),
			"Unknown TDH API Username",
			"The provider cannot create the TDH API client as there is an unknown configuration value for the TDH API Username. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the '%s' environment variable. ", EnvUsername),
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("Password"),
			"Unknown TDH API Password",
			"The provider cannot create the TDH API client as there is an unknown configuration value for the TDH API Password. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the '%s' environment variable. ", EnvPassword),
		)
	}

	if orgId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("org id"),
			"Unknown TDH API Org Id",
			"The provider cannot create the TDH API client as there is an unknown configuration value for the TDH API Org Id. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the '%s' environment variable.", EnvOrgId),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tdh_host", host)
	ctx = tflog.SetField(ctx, "tdh_org_id", orgId)
	ctx = tflog.SetField(ctx, "tdh_username", username)
	ctx = tflog.SetField(ctx, "tdh_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "tdh_password")

	tflog.Debug(ctx, "Creating TDH client")

	// Create a new TDH client using the configuration values
	client, err := tdh.NewClient(&host, &model.ClientAuth{
		OAuthAppType: oauth_type.UserCredentials,
		OrgId:        orgId,
		Username:     username,
		Password:     password,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create TDH API Client",
			"An unexpected error occurred when creating the TDH API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"TDH Client Error: "+err.Error(),
		)
		return
	}

	// Make the TDH client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured TDH client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *tdhProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewInstanceTypesDataSource,
		NewRegionsDataSource,
		NewNetworkPoliciesDataSource,
		NewNetworkPortsDataSource,
		NewUsersDataSource,
		NewRolesDataSource,
		NewMdsPoliciesDatasource,
		NewServiceAccountsDataSource,
		NewPolicyTypesDataSource,
		NewClusterMetadataDataSource,
		NewClustersDatasource,
		NewServiceRolesDatasource,
		NewCloudAccountsDatasource,
		NewProviderTypesDataSource,
		NewCertificatesDatasource,
		NewDnsDatasource,
		NewSmtpDatasource,
		NewFleetHealthDatasource,
		NewDataPlaneDatasource,
		NewDataPlaneHelmReleaseDatasource,
		NewK8sClustersDatasource,
		NewObjectStorageDatasource,
		NewTasksDataSource,
		NewLocalUsersDataSource,
		NewBackupDataSource,
		NewRestoresDataSource,
		NewStorageClassDataSource,
		NewEligibleDataPlanesDatasource,
		NewClusterVersionsDataSource,
		NewServiceExtensionsDataSource,
		NewClusterTargetVersionsDataSource,
		NewOrganizationsDatasource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *tdhProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
		NewClusterNetworkPoliciesAssociationResource,
		NewUserResource,
		NewServiceAccountResource,
		NewPolicyResource,
		NewNetworkPolicyResource,
		NewDataPlaneResource,
		NewCloudAccountResource,
		NewCertificateResource,
		NewSmtpResource,
		NewObjectStorageResource,
		NewLocalUserResource,
		NewClusterBackupResource,
	}
}

func supportedServiceTypesMarkdown() string {
	var sb strings.Builder
	serviceTypes := service_type.GetAll()
	sb.WriteString(fmt.Sprintf("`%s`", serviceTypes[0]))
	for _, serviceType := range serviceTypes[1:] {
		sb.WriteString(fmt.Sprintf(", `%s`", serviceType))
	}
	return sb.String()
}

func supportedDataServiceTypesMarkdown() string {
	var sb strings.Builder
	serviceTypes := service_type.GetAll()
	sb.WriteString(fmt.Sprintf("`%s`", serviceTypes[0]))
	for _, serviceType := range serviceTypes[1:] {
		if serviceType == service_type.RABBITMQ {
			continue
		}
		sb.WriteString(fmt.Sprintf(", `%s`", serviceType))
	}
	return sb.String()
}

func supportedIdentityTypesMarkdown() string {
	var sb strings.Builder
	serviceTypes := identity_type.GetAll()
	sb.WriteString(fmt.Sprintf("`%s`", serviceTypes[0]))
	for _, serviceType := range serviceTypes[1:] {
		sb.WriteString(fmt.Sprintf(", `%s`", serviceType))
	}
	return sb.String()
}
