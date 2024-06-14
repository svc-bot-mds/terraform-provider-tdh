package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/oauth_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/service_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/SRE"
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

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &tdhProvider{}
}

// tdhProvider is the provider implementation.
type tdhProvider struct{}

// tdhProviderModel maps provider schema data to a Go type.
type tdhProviderModel struct {
	Host         types.String `tfsdk:"host"`
	Type         types.String `tfsdk:"type"`
	ApiToken     types.String `tfsdk:"api_token"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	OrgId        types.String `tfsdk:"org_id"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

// Metadata returns the provider type name.
func (p *tdhProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tdh"
}

// Schema defines the provider-level schema for configuration data.
func (p *tdhProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with VMware Managed Data Services",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "URI for TDH API. May also be provided via *TDH_HOST* environment variable.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "OAuthType for the TDH API. It can be `api_token` or `client_credentials` or `user_creds`",
				Required:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "(Required for `api_token`) API Token for TDH API. May also be provided via *TDH_API_TOKEN* environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "(Required for `client_credentials`) Client Id for TDH API. May also be provided via *TDH_CLIENT_ID* environment variable.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "(Required for `client_credentials`) Client Secret for TDH API. May also be provided via *TDH_CLIENT_SECRET* environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "(Required for `client_credentials`) Organization Id for TDH API. May also be provided via *TDH_ORG_ID* environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "(Required for `user_creds`) Username for TDH API.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "(Required for `user_creds`) Password for TDH API.",
				Optional:            true,
				Sensitive:           true,
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

	host := os.Getenv("TDH_HOST")
	apiToken := os.Getenv("TDH_API_TOKEN")
	clientSecret := os.Getenv("TDH_CLIENT_SECRET")
	clientId := os.Getenv("TDH_CLIENT_ID")
	orgId := os.Getenv("TDH_ORG_ID")
	username := os.Getenv("TDH_USERNAME")
	password := os.Getenv("TDH_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	if config.Type.ValueString() == oauth_type.ApiToken {
		if !config.ApiToken.IsNull() {
			apiToken = config.ApiToken.ValueString()
		}
	}
	if config.Type.ValueString() == oauth_type.ClientCredentials {
		if !config.ClientId.IsNull() {
			clientId = config.ClientId.ValueString()
		}

		if !config.ClientSecret.IsNull() {
			clientSecret = config.ClientSecret.ValueString()
		}

		if !config.OrgId.IsNull() {
			orgId = config.OrgId.ValueString()
		}
	}
	if config.Type.ValueString() == oauth_type.UserCredentials {
		if !config.Username.IsNull() {
			username = config.Username.ValueString()
		}
		if !config.Password.IsNull() {
			password = config.Password.ValueString()
		}
		if !config.OrgId.IsNull() {
			orgId = config.OrgId.ValueString()
		}
	}
	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing TDH API Host",
			"The provider cannot create the TDH API client as there is a missing or empty value for the TDH API host. "+
				"Set the host value in the configuration or use the TDH_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiToken == "" && config.Type.ValueString() == oauth_type.ApiToken {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing TDH API Token",
			"The provider cannot create the TDH API client as there is a missing or empty value for the TDH API Token. "+
				"Set the password value in the configuration or use the TDH_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if config.Type.ValueString() == oauth_type.ClientCredentials {
		if clientId == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("client_id"),
				"Missing TDH API Client Id",
				"The provider cannot create the TDH API client as there is a missing or empty value for the TDH API Client Id. "+
					"Set the password value in the configuration or use the TDH_CLIENT_ID environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}

		if clientSecret == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("client_secret"),
				"Missing TDH API Client Secret",
				"The provider cannot create the TDH API client as there is a missing or empty value for the TDH API Client Secret. "+
					"Set the password value in the configuration or use the TDH_CLIENT_SECRET environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}

		if orgId == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("org_id"),
				"Missing TDH API Org Id",
				"The provider cannot create the TDH API client as there is a missing or empty value for the TDH API Org Id. "+
					"Set the password value in the configuration or use the TDH_ORG_ID environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}
	}

	if config.Type.ValueString() == oauth_type.UserCredentials {
		if username == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("Username"),
				"Unknown TDH API Username",
				"The provider cannot create the TDH API client as there is an unknown configuration value for the TDH API Username. "+
					"Either target apply the source of the value first, set the value statically in the configuration, or use the TDH_USERNAME environment variable. ",
			)
		}

		if password == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("Password"),
				"Unknown TDH API Password",
				"The provider cannot create the TDH API client as there is an unknown configuration value for the TDH API Password. "+
					"Either target apply the source of the value first, set the value statically in the configuration, or use the TDH_PASSWORD environment variable. ",
			)
		}

		if orgId == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("org id"),
				"Unknown TDH API Org Id",
				"The provider cannot create the TDH API client as there is an unknown configuration value for the TDH API Org Id. "+
					"Either target apply the source of the value first, set the value statically in the configuration, or use the TDH_ORG_ID environment variable.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "tdh_host", host)
	if config.Type.ValueString() == oauth_type.ClientCredentials {
		ctx = tflog.SetField(ctx, "tdh_client_id", clientId)
		ctx = tflog.SetField(ctx, "tdh_client_secret", clientSecret)
		ctx = tflog.SetField(ctx, "tdh_org_id", orgId)
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "tdh_client_secret")
	} else if config.Type.ValueString() == oauth_type.UserCredentials {
		ctx = tflog.SetField(ctx, "tdh_username", username)
		ctx = tflog.SetField(ctx, "tdh_password", password)
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "tdh_password")
	} else {
		ctx = tflog.SetField(ctx, "tdh_api_token", apiToken)
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "tdh_api_token")
	}

	tflog.Debug(ctx, "Creating TDH client")

	// Create a new TDH client using the configuration values
	client, err := tdh.NewClient(&host, &model.ClientAuth{
		ApiToken:     apiToken,
		ClientSecret: clientSecret,
		ClientId:     clientId,
		OrgId:        orgId,
		OAuthAppType: config.Type.ValueString(),
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
		SRE.NewCertificatesDatasource,
		SRE.NewDnsDatasource,
		SRE.NewSmtpDatasource,
		SRE.NewFleetHealthDatasource,
		SRE.NewDataplaneDatasource,
		SRE.NewDataplaneHelmReleaseDatasource,
		SRE.NewDataplaneKubernetesClusterDatasource,
		NewObjectStorageDatasource,
		NewTasksDataSource,
		NewLocalUsersDataSource,
		NewBackupDataSource,
		NewRestoreDataSource,
		NewEligibleSharedDataplanesDatasource,
		NewEligibleDedicatedDataplanesDatasource,
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
		SRE.NewCertificateResource,
		SRE.NewSmtpResource,
		NewObjectStorageResource,
		NewLocalUserResource,
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
