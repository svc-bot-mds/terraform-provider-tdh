package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
)

var (
	_ datasource.DataSource              = &storageClassDataSource{}
	_ datasource.DataSourceWithConfigure = &storageClassDataSource{}
)

type StorageClassDataSourceModel struct {
	AccountId      types.String        `tfsdk:"cloud_account_id"`
	K8SClusterName types.String        `tfsdk:"k8s_cluster_name"`
	StorageClasses []StorageClassModel `tfsdk:"list"`
}

type StorageClassModel struct {
	Name        types.String `tfsdk:"name"`
	Provisioner types.String `tfsdk:"provisioner"`
}

func NewStorageClassDataSource() datasource.DataSource {
	return &storageClassDataSource{}
}

type storageClassDataSource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *storageClassDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_policies"
}

// Schema defines the schema for the data source.
func (d *storageClassDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch Storage Classes available based on the cloud provider account and K8s Cluster name.",
		Attributes: map[string]schema.Attribute{
			"cloud_account_id": schema.StringAttribute{
				Description: "ID of the Cloud Provider Account.",
				Required:    true,
			},
			"k8s_cluster_name": schema.StringAttribute{
				Description: "K8s Cluster Name.",
				Required:    true,
			},
			"list": schema.ListNestedAttribute{
				MarkdownDescription: "List of Storage Classes",
				Computed:            true,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the Storage class.",
							Computed:    true,
						},
						"provisioner": schema.StringAttribute{
							Description: "Provisioner name of the Storage Class.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *storageClassDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	var plan StorageClassDataSourceModel
	var state StorageClassDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	query := &infra_connector.StorageClassesQuery{
		AccountId:   plan.AccountId.ValueString(),
		ClusterName: plan.K8SClusterName.ValueString(),
	}
	storageClassDto, err := d.client.InfraConnector.GetK8sClusterStorageClasses(query)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Storage Classes",
			err.Error(),
		)
		return
	}
	tflog.Info(ctx, "fetched storage class", map[string]interface{}{"metadata": &storageClassDto})

	// Map cluster metadata body to model
	for _, list := range storageClassDto {
		storageClass := StorageClassModel{
			Name:        types.StringValue(list.StorageClassName),
			Provisioner: types.StringValue(list.Provisioner),
		}
		state.StorageClasses = append(state.StorageClasses, storageClass)
	}

	tflog.Debug(ctx, "before setting state", map[string]interface{}{
		"state": state,
	})
	// Set state
	diags := resp.State.Set(ctx, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *storageClassDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
