package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &organizationsDatasource{}
	_ datasource.DataSourceWithConfigure = &organizationsDatasource{}
)

// objectStoragesDatasourceModel maps the data source schema data.
type organizationsDatasourceModel struct {
	Id            types.String `tfsdk:"id"`
	Organizations []orgsModel  `tfsdk:"list"`
}

type orgsModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
	SreOrg types.Bool   `tfsdk:"sre_org"`
}

// NewOrganizationsDatasource is a helper function to simplify the provider implementation.
func NewOrganizationsDatasource() datasource.DataSource {
	return &organizationsDatasource{}
}

// objectStorageDatasource is the data source implementation.
type organizationsDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *organizationsDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

// Schema defines the schema for the data source.
func (d *organizationsDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch all organizations on TDH.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"list": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the organization.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the organization.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the organization.",
							Computed:    true,
						},
						"sre_org": schema.BoolAttribute{
							MarkdownDescription: "Denotes if the organization is `SRE` org.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *organizationsDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationsDatasourceModel
	var modelList []orgsModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &controller.FleetsQuery{}

	response, err := d.client.Controller.GetOrganizations(*query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Organizations",
			err.Error(),
		)
		return
	}

	if response.Page.TotalPages > 1 {
		for i := 1; i <= response.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			page, err := d.client.Controller.GetOrganizations(*query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read Organizations",
					err.Error(),
				)
				return
			}

			for _, dto := range *page.Get() {
				tfModel := d.convertToTfModel(dto)
				modelList = append(modelList, tfModel)
			}
		}

		tflog.Debug(ctx, "Organizations models: ", map[string]interface{}{"models": modelList})
		state.Organizations = append(state.Organizations, modelList...)
	} else {
		for _, dto := range *response.Get() {
			tflog.Info(ctx, "Converting dto: ", map[string]interface{}{"dto": dto})
			tfModel := d.convertToTfModel(dto)
			tflog.Info(ctx, "converted org model: ", map[string]interface{}{"model": tfModel})
			state.Organizations = append(state.Organizations, tfModel)
		}
	}
	state.Id = types.StringValue(common.DataSource + common.Organizations)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *organizationsDatasource) convertToTfModel(dto model.OrgModel) orgsModel {
	return orgsModel{
		ID:     types.StringValue(dto.OrgId),
		Name:   types.StringValue(dto.OrgName),
		Status: types.StringValue(dto.Status),
		SreOrg: types.BoolValue(dto.SreOrg),
	}
}

// Configure adds the provider configured client to the data source.
func (d *organizationsDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
