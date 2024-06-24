package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"reflect"
)

var (
	_ datasource.DataSource              = &certificatesDatasource{}
	_ datasource.DataSourceWithConfigure = &certificatesDatasource{}
)

// certificatesDatasourceModel maps the data source schema data.
type certificatesDatasourceModel struct {
	Id   types.String        `tfsdk:"id"`
	List []certificatesModel `tfsdk:"list"`
}

type certificatesModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	DomainName     types.String `tfsdk:"domain_name"`
	ProviderType   types.String `tfsdk:"provider_type"`
	ExpiryTime     types.String `tfsdk:"expiry_time"`
	Status         types.String `tfsdk:"status"`
	Organization   types.String `tfsdk:"organization"`
	DataPlaneCount types.Int64  `tfsdk:"data_plane_count"`
	CreatedBy      types.String `tfsdk:"created_by"`
}

// NewCertificatesDatasource is a helper function to simplify the provider implementation.
func NewCertificatesDatasource() datasource.DataSource {
	return &certificatesDatasource{}
}

// certificatesDatasource is the data source implementation.
type certificatesDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *certificatesDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

// Schema defines the schema for the data source.
func (d *certificatesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch all certificates on TDH.\n" +
			"**Note:** For SRE only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"certificates": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the certificate.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the certificate.",
							Computed:    true,
						},
						"provider_type": schema.StringAttribute{
							Description: "Provider type of the certificate",
							Computed:    true,
						},
						"domain_name": schema.StringAttribute{
							Description: "Domain name of the certificate",
							Computed:    true,
						},
						"expiry_time": schema.StringAttribute{
							Description: "Expiry Time of the certificate",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the certificate",
							Computed:    true,
						},
						"data_plane_count": schema.Int64Attribute{
							Description: "Number of data plane using this certificate",
							Computed:    true,
						},
						"organization": schema.StringAttribute{
							Description: "Org name on which the certificate was created",
							Computed:    true,
						},
						"created_by": schema.StringAttribute{
							Description: "Name of the identity which this certificate was created by.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *certificatesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state certificatesDatasourceModel
	var certificateList []certificatesModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &infra_connector.CertificatesQuery{}

	certificates, err := d.client.InfraConnector.GetCertificates(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Certificates",
			err.Error(),
		)
		return
	}

	if certificates.Page.TotalPages > 1 {
		for i := 1; i <= certificates.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			_, err := d.client.InfraConnector.GetCertificates(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read certificates",
					err.Error(),
				)
				return
			}
			certificateList = d.convertToTfModels(&ctx, certificates.Get())
		}
		tflog.Debug(ctx, "all pages certificates dto", map[string]interface{}{"dto": certificateList})
	} else {
		tflog.Debug(ctx, "single page certificates dto", map[string]interface{}{"dto": certificateList})
		certificateList = d.convertToTfModels(&ctx, certificates.Get())
	}
	state.List = append(state.List, certificateList...)
	state.Id = types.StringValue(common.DataSource + common.CertificateId)

	tflog.Info(ctx, "final", map[string]interface{}{"dto": state})
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *certificatesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}

func (d *certificatesDatasource) convertToTfModels(ctx *context.Context, certificates *[]model.Certificate) []certificatesModel {
	tflog.Debug(*ctx, "converting to tfModels")
	var certificateList []certificatesModel
	for _, certificateDto := range *certificates {
		tflog.Info(*ctx, "Converting certificate Dto1", map[string]interface{}{"dto": certificateDto})
		certificate := certificatesModel{
			ID:           types.StringValue(certificateDto.Id),
			Name:         types.StringValue(certificateDto.Name),
			DomainName:   types.StringValue(certificateDto.DomainName),
			ProviderType: types.StringValue(certificateDto.Provider),
			ExpiryTime:   types.StringValue(certificateDto.ExpiryTime),
			Status:       types.StringValue(certificateDto.Status),
			Organization: types.StringValue(certificateDto.OrgId),
		}
		tflog.Info(*ctx, "converted certificate Dto", map[string]interface{}{"dto": certificate})

		if certificateDto.Deployemnts == nil {
			certificate.DataPlaneCount = types.Int64Value(0)
		} else {
			keys := reflect.ValueOf(certificateDto.Deployemnts).Len()
			certificate.DataPlaneCount = types.Int64Value(int64(keys))
		}
		certificateList = append(certificateList, certificate)
	}
	tflog.Debug(*ctx, "converted to tfModels")
	return certificateList
}
