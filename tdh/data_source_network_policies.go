package tdh

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/constants/policy_type"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
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
	List        []networkPolicyModel `tfsdk:"list"`
	Names       []string             `tfsdk:"names"`
	ServiceType types.String         `tfsdk:"service_type"`
	Id          types.String         `tfsdk:"id"`
}

// instanceTypesModel maps coffees schema data.
type networkPolicyModel struct {
	ID          types.String      `tfsdk:"id"`
	Name        types.String      `tfsdk:"name"`
	Description types.String      `tfsdk:"description"`
	NetworkSpec *NetworkSpecModel `tfsdk:"network_spec"`
	ResourceIds types.Set         `tfsdk:"resource_ids"`
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
		MarkdownDescription: "Used to fetch all network policies on TDH.",
		Attributes: map[string]schema.Attribute{
			"names": schema.SetAttribute{
				MarkdownDescription: "Names to search policies by. Ex: `[\"allow-all\"]`.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource.",
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Filter network policies based on type of service ports they are allowing. Supported values: %s .", supportedServiceTypesMarkdown()),
				Optional:            true,
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
						"description": schema.StringAttribute{
							Description: "Description of the policy",
							Optional:    true,
							Computed:    true,
						},
						"resource_ids": schema.SetAttribute{
							Description: "IDs of service resources/instances being managed by the policy.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"network_spec": schema.SingleNestedAttribute{
							MarkdownDescription: "Network config allowing access to service resource.",
							Required:            true,
							CustomType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"cidr": types.StringType,
									"network_port_ids": types.SetType{
										ElemType: types.StringType,
									},
								},
							},
							Attributes: map[string]schema.Attribute{
								"cidr": schema.StringAttribute{
									MarkdownDescription: "CIDR value to allow access from.",
									Required:            true,
								},
								"network_port_ids": schema.SetAttribute{
									MarkdownDescription: "IDs of network ports open up for access.",
									Required:            true,
									ElementType:         types.StringType,
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
func (d *networkPoliciesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networkPoliciesDataSourceModel
	var networkPolicyList []networkPolicyModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	query := &customer_metadata.PoliciesQuery{
		Type: policy_type.NETWORK,
	}
	if len(state.Names) > 0 {
		query.Names = state.Names
	}
	if !state.ServiceType.IsNull() {
		query.ServiceType = state.ServiceType.ValueString()
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
			pageResponse, err := d.client.CustomerMetadata.GetPolicies(query)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read TDH Policies",
					err.Error(),
				)
				return
			}
			if networkPolicyList = append(networkPolicyList, d.convertToTfModels(&ctx, &resp.Diagnostics, pageResponse.Get())...); resp.Diagnostics.HasError() {
				return
			}
		}
	} else {
		if networkPolicyList = d.convertToTfModels(&ctx, &resp.Diagnostics, nwPolicies.Get()); resp.Diagnostics.HasError() {
			return
		}
	}
	tflog.Debug(ctx, "network policy list", map[string]interface{}{"dto": networkPolicyList})
	state.List = append(state.List, networkPolicyList...)

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

func (d *networkPoliciesDatasource) convertToTfModels(ctx *context.Context, diag *diag.Diagnostics, nwPolicies *[]model.Policy) []networkPolicyModel {
	tflog.Debug(*ctx, "converting to tfModels")
	var networkPolicyList []networkPolicyModel
	for _, dto := range *nwPolicies {
		tfModel := networkPolicyModel{
			ID:          types.StringValue(dto.ID),
			Name:        types.StringValue(dto.Name),
			Description: types.StringValue(dto.Description),
		}
		tfNetworkSpecModels := make([]*NetworkSpecModel, len(dto.NetworkSpec))
		for i, networkSpec := range dto.NetworkSpec {
			tfNetworkSpecModels[i] = &NetworkSpecModel{
				Cidr: types.StringValue(networkSpec.CIDR),
			}
			networkPortIds, _ := types.SetValueFrom(*ctx, types.StringType, networkSpec.NetworkPortIds)
			tfNetworkSpecModels[i].NetworkPortIds = networkPortIds
		}
		tfModel.NetworkSpec = tfNetworkSpecModels[0]

		resourceIds, diags := types.SetValueFrom(*ctx, types.StringType, dto.ResourceIds)
		if diag.Append(diags...); diag.HasError() {
			return networkPolicyList
		}
		tfModel.ResourceIds = resourceIds

		tflog.Debug(*ctx, "network policy", map[string]interface{}{"tfModel": tfModel})
		networkPolicyList = append(networkPolicyList, tfModel)
	}
	tflog.Debug(*ctx, "converted to tfModels")
	return networkPolicyList
}
