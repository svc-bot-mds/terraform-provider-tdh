package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/policy_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	customer_metadata "github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &networkPoliciesDatasource{}
	_ datasource.DataSourceWithConfigure = &networkPoliciesDatasource{}
)

// instanceTypesDataSourceModel maps the data source schema data.
type networkPoliciesDataSourceModel struct {
	Policies []networkPoliciesModel `tfsdk:"policies"`
	Names    []string               `tfsdk:"names"`
	Id       types.String           `tfsdk:"id"`
}

// instanceTypesModel maps coffees schema data.
type networkPoliciesModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewNetworkPoliciesDataSource is a helper function to simplify the provider implementation.
func NewNetworkPoliciesDataSource() datasource.DataSource {
	return &networkPoliciesDatasource{}
}

// networkPoliciesDatasource is the data source implementation.
type networkPoliciesDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *networkPoliciesDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_policies"
}

// Schema defines the schema for the data source.
func (d *networkPoliciesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch all `NETWORK` type policies on TDH.",
		Attributes: map[string]schema.Attribute{
			"names": schema.SetAttribute{
				MarkdownDescription: "Names to search policies by. Ex: `[\"allow-all\"]` .",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource.",
			},
			"policies": schema.ListNestedAttribute{
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
func (d *networkPoliciesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networkPoliciesDataSourceModel
	var networkPolicyList []networkPoliciesModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &customer_metadata.PoliciesQuery{
		Type:  policy_type.NETWORK,
		Names: state.Names,
	}
	//state.Names.ElementsAs(ctx, query.Names, true)
	nwPolicies, err := d.client.CustomerMetadata.GetPolicies(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Network Policies",
			err.Error(),
		)
		return
	}

	if nwPolicies.Page.TotalPages > 1 {
		for i := 1; i <= nwPolicies.Page.TotalPages; i++ {
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
				policy := networkPoliciesModel{
					ID:   types.StringValue(mdsPolicyDTO.ID),
					Name: types.StringValue(mdsPolicyDTO.Name),
				}
				networkPolicyList = append(networkPolicyList, policy)
			}
		}

		tflog.Debug(ctx, "rabbitmq dto", map[string]interface{}{"dto": networkPolicyList})
		state.Policies = append(state.Policies, networkPolicyList...)
	} else {
		for _, mdsPolicyDTO := range *nwPolicies.Get() {
			networkPolicy := networkPoliciesModel{
				ID:   types.StringValue(mdsPolicyDTO.ID),
				Name: types.StringValue(mdsPolicyDTO.Name),
			}
			tflog.Debug(ctx, "nwPolicy dto", map[string]interface{}{"dto": networkPolicy})
			state.Policies = append(state.Policies, networkPolicy)
		}
	}

	state.Id = types.StringValue(common.DataSource + common.NetworkPoliciesId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *networkPoliciesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
