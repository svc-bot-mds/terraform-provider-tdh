package SRE

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &fleetHealthDatasource{}
	_ datasource.DataSourceWithConfigure = &fleetHealthDatasource{}
)

// FleetHealthDataSourceModel maps the data source schema data.
type FleetHealthDataSourceModel struct {
	Id          types.String       `tfsdk:"id"`
	FleetHealth []FleetHealthModel `tfsdk:"fleethealth""`
}
type FleetHealthModel struct {
	TotalOrgCount       types.Int64              `tfsdk:"total_org_count"`
	TotalHealthyCount   types.Int64              `tfsdk:"total_healthy_org_count"`
	TotalUnHealthyCount types.Int64              `tfsdk:"total_unhealthy_org_count"`
	DedicatedDataplanes types.Int64              `tfsdk:"dedicated_dataplanes"`
	HealthyDataplanes   types.Int64              `tfsdk:"healthy_dataplanes"`
	SharedDataplanes    types.Int64              `tfsdk:"shared_dataplanes"`
	TotalDataplanes     types.Int64              `tfsdk:"total_dataplanes"`
	UnhealthyDataplanes types.Int64              `tfsdk:"unhealthy_dataplanes"`
	ClusterCount        []ClusterCountModel      `tfsdk:"cluster_counts"`
	ResourceByService   []ResourceByServiceModel `tfsdk:"resource_by_service"`
	FleetDetails        []FleetsModel            `tfsdk:"fleets"`
}

type ResourceByServiceModel struct {
	DataPlaneName types.String `tfsdk:"dataplane_name"`
	ServiceType   types.String `tfsdk:"resource_service_type"`
	Cpu           types.String `tfsdk:"cpu"`
	Memory        types.String `tfsdk:"memory"`
	Storage       types.String `tfsdk:"storage"`
}

type ClusterCountModel struct {
	Count       types.Int64  `tfsdk:"count"`
	ServiceType types.String `tfsdk:"service_type"`
}

type FleetsModel struct {
	OrgName                  types.String           `tfsdk:"org_name"`
	ClusterStatus            ClusterStatus          `tfsdk:"cluster_status"`
	ClusterCounts            types.Int64            `tfsdk:"cluster_counts"`
	CustomerClusterInfo      []CustomerClusterModel `tfsdk:"customer_cluster_info"`
	CustomerCumulativeStatus types.String           `tfsdk:"org_status"`
	SreOrg                   types.Bool             `tfsdk:"sre_org"`
}

type ClusterStatus struct {
	Critical types.Int64 `tfsdk:"critical"`
	Warning  types.Int64 `tfsdk:"warning"`
	Healthy  types.Int64 `tfsdk:"healthy"`
}

type CustomerClusterModel struct {
	ClusterId    types.String `tfsdk:"cluster_id"`
	ClusterName  types.String `tfsdk:"cluster_name"`
	InstanceSize types.String `tfsdk:"instance_size"`
	Status       types.String `tfsdk:"status"`
	ServiceType  types.String `tfsdk:"service_type"`
}

// NewFleetHealthDataSource is a helper function to simplify the provider implementation.
func NewFleetHealthDatasource() datasource.DataSource {
	return &fleetHealthDatasource{}
}

// fleetHealthDatasource is the data source implementation.
type fleetHealthDatasource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *fleetHealthDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fleet_health"

}

// Schema defines the schema for the data source.
func (d *fleetHealthDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Used to fetch Fleet Health Details for SRE",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The testing framework requires an id attribute to be present in every data source and resource",
			},
			"fleethealth": schema.ListNestedAttribute{
				Description: "Fleet Health Details.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{

						"total_org_count": schema.Int64Attribute{
							Description: "Total Count of Organizations",
							Computed:    true,
						},
						"total_healthy_org_count": schema.Int64Attribute{
							Description: "Total Count of Healthy Organizations",
							Computed:    true,
						},
						"total_unhealthy_org_count": schema.Int64Attribute{
							Description: "Total Count of UnHealthy Organizations",
							Computed:    true,
						},
						"dedicated_dataplanes": schema.Int64Attribute{
							Description: "Total Count of Dedicated Dataplanes",
							Computed:    true,
						},
						"healthy_dataplanes": schema.Int64Attribute{
							Description: "Total Count of Healthy Dataplanes",
							Computed:    true,
						},
						"shared_dataplanes": schema.Int64Attribute{
							Description: "Total Count of Shared Dataplanes",
							Computed:    true,
						},
						"total_dataplanes": schema.Int64Attribute{
							Description: "Total Count of Dataplanes",
							Computed:    true,
						},
						"unhealthy_dataplanes": schema.Int64Attribute{
							Description: "Total Count of UnHealthy Dataplnes",
							Computed:    true,
						},
						"cluster_counts": schema.ListNestedAttribute{
							Description: "Cluster Count by Service.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"count": schema.Int64Attribute{
										Description: "Count of the Cluster",
										Computed:    true,
									},
									"service_type": schema.StringAttribute{
										Description: "Service Type",
										Computed:    true,
									},
								},
							},
						},
						"resource_by_service": schema.ListNestedAttribute{
							Description: "Resource by Service",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"dataplane_name": schema.StringAttribute{
										Description: "Dataplane Name",
										Computed:    true,
									},
									"resource_service_type": schema.StringAttribute{
										Description: "Service Type",
										Computed:    true,
									},
									"storage": schema.StringAttribute{
										Description: "Storage",
										Computed:    true,
									},
									"memory": schema.StringAttribute{
										Description: "memory",
										Computed:    true,
									},
									"cpu": schema.StringAttribute{
										Description: "CPU",
										Computed:    true,
									},
								},
							},
						},
						"fleets": schema.ListNestedAttribute{
							Description: "Organization Details",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"org_name": schema.StringAttribute{
										Description: "Organization Name",
										Computed:    true,
									},
									"org_status": schema.StringAttribute{
										Description: "Organization Status",
										Computed:    true,
									},
									"cluster_status": schema.SingleNestedAttribute{
										Description: "Additional info for the cluster.",
										Required:    true,
										Attributes: map[string]schema.Attribute{
											"critical": schema.Int64Attribute{
												Description: "Critical -Cluster Count.",
												Required:    true,
											},
											"warning": schema.Int64Attribute{
												Description: "Warning - Cluster Count.",
												Required:    true,
											},
											"healthy": schema.Int64Attribute{
												Description: "Healthy - Cluster Count.",
												Required:    false,
												Optional:    true,
											},
										},
									},
									"cluster_counts": schema.Int64Attribute{
										Description: "Cluster Count",
										Computed:    true,
									},
									"customer_cluster_info": schema.ListNestedAttribute{
										Description: "Organization - Cluster Details",
										Computed:    true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"cluster_id": schema.StringAttribute{
													Description: "Cluster Id",
													Computed:    true,
												},
												"cluster_name": schema.StringAttribute{
													Description: "Cluster Name",
													Computed:    true,
												},
												"instance_size": schema.StringAttribute{
													Description: "Instance Size",
													Computed:    true,
												},
												"status": schema.StringAttribute{
													Description: "Cluster Status",
													Computed:    true,
												},
												"service_type": schema.StringAttribute{
													Description: "Service Type of the Cluster",
													Computed:    true,
												},
											},
										},
									},
									"sre_org": schema.BoolAttribute{
										Description: "Flag which denotes if the organization is SRE organization/Default Org of SRE",
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
func (d *fleetHealthDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state FleetHealthDataSourceModel
	tflog.Info(ctx, "INIT -- READ cluster counts")
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	orgHealth, err := d.client.InfraConnector.GetOrgHealthDetails()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Org Health Details",
			err.Error(),
		)
		return
	}

	dataplaneCount, err := d.client.InfraConnector.GetDataplaneCounts()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Org Health Details",
			err.Error(),
		)
		return
	}

	clusterCount, err := d.client.Controller.GetClusterCountByService()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Cluster Count Details",
			err.Error(),
		)
		return
	}

	resourceByService, err := d.client.Controller.GetResourceByService()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH Cluster Resource Details grouped by Service",
			err.Error(),
		)
		return
	}

	fleetsQuery := &controller.FleetsQuery{}
	srecustomerInfo, err := d.client.Controller.GetFleetDetails(fleetsQuery)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read TDH customer details for SRE",
			err.Error())
	}

	fleetHealthDto := FleetHealthModel{
		TotalHealthyCount:   types.Int64Value(orgHealth.TotalHealthyOrgsCount),
		TotalOrgCount:       types.Int64Value(orgHealth.TotalOrgCount),
		TotalUnHealthyCount: types.Int64Value(orgHealth.TotalUnhealthyOrgsCount),
		TotalDataplanes:     types.Int64Value(dataplaneCount.TotalDataplanes),
		HealthyDataplanes:   types.Int64Value(dataplaneCount.HealthyDataplanes),
		UnhealthyDataplanes: types.Int64Value(dataplaneCount.UnhealthyDataplanes),
		SharedDataplanes:    types.Int64Value(dataplaneCount.SharedDataplanes),
		DedicatedDataplanes: types.Int64Value(dataplaneCount.DedicatedDataplanes),
	}

	for _, cc := range clusterCount {
		ccList := ClusterCountModel{
			ServiceType: types.StringValue(cc.ServiceType),
			Count:       types.Int64Value(cc.Count),
		}

		fleetHealthDto.ClusterCount = append(fleetHealthDto.ClusterCount, ccList)
	}

	for _, cc := range resourceByService {
		ccList := ResourceByServiceModel{
			ServiceType:   types.StringValue(cc.DataPlaneName),
			DataPlaneName: types.StringValue(cc.DataPlaneName),
			Cpu:           types.StringValue(cc.Cpu),
			Memory:        types.StringValue(cc.Memory),
			Storage:       types.StringValue(cc.Storage),
		}

		fleetHealthDto.ResourceByService = append(fleetHealthDto.ResourceByService, ccList)
	}

	if srecustomerInfo.Page.TotalPages > 1 {
		for i := 1; i <= srecustomerInfo.Page.TotalPages; i++ {
			fleetsQuery.PageQuery.Index = i - 1
			totalFleets, err := d.client.Controller.GetFleetDetails(fleetsQuery)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read Customer Details",
					err.Error(),
				)
				return
			}

			for _, fleetsDto := range *totalFleets.Get() {
				fleets := FleetsModel{
					OrgName:                  types.StringValue(fleetsDto.Name),
					SreOrg:                   types.BoolValue(fleetsDto.SreOrg),
					CustomerCumulativeStatus: types.StringValue(fleetsDto.CustomerCumulativeStatus),
					ClusterCounts:            types.Int64Value(fleetsDto.ClusterCounts),
				}
				for _, cc := range fleetsDto.CustomerClusterInfo {
					ccList := CustomerClusterModel{
						ClusterId:    types.StringValue(cc.ClusterId),
						ClusterName:  types.StringValue(cc.ClusterName),
						InstanceSize: types.StringValue(cc.InstanceSize),
						Status:       types.StringValue(cc.Status),
						ServiceType:  types.StringValue(cc.ServiceType),
					}

					fleets.CustomerClusterInfo = append(fleets.CustomerClusterInfo, ccList)
				}
				fleetHealthDto.FleetDetails = append(fleetHealthDto.FleetDetails, fleets)
			}
		}

	} else {
		for _, fleetsDto := range *srecustomerInfo.Get() {
			tflog.Info(ctx, "Converting fleet Dto", map[string]interface{}{"dto": fleetsDto})
			fleets := FleetsModel{
				OrgName:                  types.StringValue(fleetsDto.Name),
				SreOrg:                   types.BoolValue(fleetsDto.SreOrg),
				CustomerCumulativeStatus: types.StringValue(fleetsDto.CustomerCumulativeStatus),
				ClusterCounts:            types.Int64Value(fleetsDto.ClusterCounts),
			}

			customerClusterModel := ClusterStatus{
				Critical: types.Int64Value(fleetsDto.ClusterStatus.Critical),
				Warning:  types.Int64Value(fleetsDto.ClusterStatus.Warning),
				Healthy:  types.Int64Value(fleetsDto.ClusterStatus.Healthy),
			}
			fleets.ClusterStatus = customerClusterModel
			for _, cc := range fleetsDto.CustomerClusterInfo {
				ccList := CustomerClusterModel{
					ClusterId:    types.StringValue(cc.ClusterId),
					ClusterName:  types.StringValue(cc.ClusterName),
					InstanceSize: types.StringValue(cc.InstanceSize),
					Status:       types.StringValue(cc.Status),
					ServiceType:  types.StringValue(cc.ServiceType),
				}

				fleets.CustomerClusterInfo = append(fleets.CustomerClusterInfo, ccList)
			}
			fleetHealthDto.FleetDetails = append(fleetHealthDto.FleetDetails, fleets)
		}
	}
	state.FleetHealth = append(state.FleetHealth, fleetHealthDto)
	state.Id = types.StringValue(common.DataSource + common.ServiceRolesId)
	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *fleetHealthDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}
