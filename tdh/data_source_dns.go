package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	infra_connector "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &dnsDatasource{}
	_ datasource.DataSourceWithConfigure = &dnsDatasource{}
)

// DnsDataSourceModel maps the data source schema data.
type DnsDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	List []DNSModel   `tfsdk:"list"`
}
type DNSModel struct {
	Id       types.String  `tfsdk:"id"`
	Servers  []ServerModel `tfsdk:"servers"`
	Name     types.String  `tfsdk:"name"`
	Provider types.String  `tfsdk:"provider"`
	Domain   types.String  `tfsdk:"domain"`
}

// ServerModel maps role schema data.
type ServerModel struct {
	Host       types.String `tfsdk:"host"`
	Port       types.Int64  `tfsdk:"port"`
	Protocol   types.String `tfsdk:"protocol"`
	ServerType types.String `tfsdk:"server_type"`
}

// NewDnsDatasource is a helper function to simplify the provider implementation.
func NewDnsDatasource() datasource.DataSource {
	return &dnsDatasource{}
}

// rolesDatasource is the data source implementation.
type dnsDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *dnsDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns"

}

// Schema defines the schema for the data source.
func (d *dnsDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch DNS configurations available on TDH.\n ## Note: For SRE only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"list": schema.ListNestedAttribute{
				Description: "List of DNS Config.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the DNS config",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the DNS Config",
							Computed:    true,
						},
						"provider": schema.StringAttribute{
							Description: "Provider of the DNS Config",
							Computed:    true,
						},
						"domain": schema.StringAttribute{
							Description: "Domain of the DNS Config",
							Computed:    true,
						},
						"servers": schema.ListNestedAttribute{
							Description: "List of servers.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"host": schema.StringAttribute{
										Description: "Host of the server.",
										Computed:    true,
									},
									"protocol": schema.StringAttribute{
										Description: "Protocol of the server.",
										Computed:    true,
									},
									"port": schema.Int64Attribute{
										Description: "Port of the server",
										Computed:    true,
									},
									"server_type": schema.StringAttribute{
										Description: "Type of the server.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dnsDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DnsDataSourceModel
	var dnsList []DNSModel
	var serverlist []ServerModel
	tflog.Info(ctx, "INIT__READ DNS config")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &infra_connector.DNSQuery{}
	dnsResponse, err := d.client.InfraConnector.GetDnsconfig(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH DNS Config List",
			err.Error(),
		)
		return
	}
	if dnsResponse.Page.TotalPages > 1 {
		for i := 1; i <= dnsResponse.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalPolicies, err := d.client.InfraConnector.GetDnsconfig(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH DNS Config List",
					err.Error(),
				)
				return
			}

			for _, dnsDto := range *totalPolicies.Get() {
				dns := DNSModel{
					Name:     types.StringValue(dnsDto.Name),
					Domain:   types.StringValue(dnsDto.Domain),
					Provider: types.StringValue(dnsDto.Provider),
				}
				for _, serverList := range dnsDto.ServerList {
					dnsServer := ServerModel{
						Host:       types.StringValue(serverList.Host),
						Port:       types.Int64Value(serverList.Port),
						Protocol:   types.StringValue(serverList.Protocol),
						ServerType: types.StringValue(serverList.ServerType),
					}

					serverlist = append(serverlist, dnsServer)
				}

				dns.Servers = append(dns.Servers, serverlist...)
				dnsList = append(dnsList, dns)
			}
		}

		state.List = append(state.List, dnsList...)
	} else {
		for _, dnsDto := range *dnsResponse.Get() {
			dns := DNSModel{
				Name:     types.StringValue(dnsDto.Name),
				Domain:   types.StringValue(dnsDto.Domain),
				Provider: types.StringValue(dnsDto.Provider),
			}
			for _, serverList := range dnsDto.ServerList {
				dnsServer := ServerModel{
					Host:       types.StringValue(serverList.Host),
					Port:       types.Int64Value(serverList.Port),
					Protocol:   types.StringValue(serverList.Protocol),
					ServerType: types.StringValue(serverList.ServerType),
				}

				serverlist = append(serverlist, dnsServer)
			}

			dns.Servers = append(dns.Servers, serverlist...)
			dnsList = append(dnsList, dns)
		}
		state.List = append(state.List, dnsList...)
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
func (d *dnsDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
