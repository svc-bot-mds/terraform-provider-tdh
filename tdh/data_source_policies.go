package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	customer_metadata "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
	"github.com/svc-bot-mds/terraform-provider-tdh/tdh/utils"
)

var (
	_ datasource.DataSource              = &mdsPoliciesDatasource{}
	_ datasource.DataSourceWithConfigure = &mdsPoliciesDatasource{}
)

// instanceTypesDataSourceModel maps the data source schema data.
type mdsPoliciesDatasourceModel struct {
	List         []mdsPoliciesModel `tfsdk:"list"`
	Id           types.String       `tfsdk:"id"`
	Names        types.List         `tfsdk:"names"`
	Type         types.String       `tfsdk:"type"`
	IdentityType types.String       `tfsdk:"identity_type"`
}

// instanceTypesModel maps coffees schema data.
type mdsPoliciesModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewMdsPoliciesDatasource is a helper function to simplify the provider implementation.
func NewMdsPoliciesDatasource() datasource.DataSource {
	return &mdsPoliciesDatasource{}
}

// networkPoliciesDatasource is the data source implementation.
type mdsPoliciesDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *mdsPoliciesDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policies"
}

// Schema defines the schema for the data source.
func (d *mdsPoliciesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch all user access control policies for services.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource.",
			},
			"names": schema.ListAttribute{
				MarkdownDescription: "Names to search policies by. Ex: `[\"read-only-postgres\"]` .",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of policies to list. Supported values: %s.", supportedServiceTypesMarkdown()),
				Optional:            true,
				Validators: []validator.String{
					utils.ServiceTypeValidator,
				},
			},
			"identity_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Type of identity, to list policies dedicated to that type. Supported values: %s.", supportedIdentityTypesMarkdown()),
				Optional:            true,
				Validators: []validator.String{
					utils.IdentityTypeValidator,
				},
			},
			"list": schema.ListNestedAttribute{
				Description: "List of fetched policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the policy.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the policy.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *mdsPoliciesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state mdsPoliciesDatasourceModel
	var policies []mdsPoliciesModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &customer_metadata.PoliciesQuery{}
	if !state.Type.IsNull() {
		query.Type = state.Type.ValueString()
	}
	if !state.IdentityType.IsNull() {
		query.IdentityType = state.IdentityType.ValueString()
	}
	tflog.Info(ctx, "policies-query", map[string]interface{}{
		"query": query,
	})

	response, err := d.client.CustomerMetadata.GetPolicies(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Policies",
			err.Error(),
		)
		return
	}

	if response.Page.TotalPages > 1 {
		for i := 1; i <= response.Page.TotalPages; i++ {
			query.PageQuery.Index = i - 1
			totalPolicies, err := d.client.CustomerMetadata.GetPolicies(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH Policies",
					err.Error(),
				)
				return
			}

			for _, mdsPolicyDTO := range *totalPolicies.Get() {
				policy := mdsPoliciesModel{
					ID:   types.StringValue(mdsPolicyDTO.ID),
					Name: types.StringValue(mdsPolicyDTO.Name),
				}
				policies = append(policies, policy)
			}
		}

		tflog.Debug(ctx, "READING dto", map[string]interface{}{"dto": policies})
		state.List = append(state.List, policies...)
	} else {
		for _, mdsPolicyDTO := range *response.Get() {
			policy := mdsPoliciesModel{
				ID:   types.StringValue(mdsPolicyDTO.ID),
				Name: types.StringValue(mdsPolicyDTO.Name),
			}
			tflog.Debug(ctx, "READING dto", map[string]interface{}{"dto": policy})
			state.List = append(state.List, policy)
		}
	}

	state.Id = types.StringValue(common.DataSource + common.PoliciesId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *mdsPoliciesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
