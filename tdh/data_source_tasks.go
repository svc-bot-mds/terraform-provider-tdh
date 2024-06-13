package tdh

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/task"
	"github.com/svc-bot-mds/terraform-provider-tdh/constants/common"
)

var (
	_ datasource.DataSource              = &tasksDataSource{}
	_ datasource.DataSourceWithConfigure = &tasksDataSource{}
)

// tasksDataSourceModel maps the datasource schema
type tasksDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	ResourceName types.String `tfsdk:"resource_name"`
	List         []taskModel  `tfsdk:"list"`
}

type taskModel struct {
	Id           types.String `tfsdk:"id"`
	Status       types.String `tfsdk:"status"`
	TaskName     types.String `tfsdk:"task_name"`
	ResourceName types.String `tfsdk:"resource_name"`
	ResourceId   types.String `tfsdk:"resource_id"`
}

func NewTasksDataSource() datasource.DataSource {
	return &tasksDataSource{}
}

type tasksDataSource struct {
	client *tdh.Client
}

// Metadata returns the data source type name.
func (d *tasksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tasks"
}

// Schema defines the schema for the data source.
func (d *tasksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Used to fetch running/completed task(s). At least one of `id` or `resource_name` is required; when both are present, preference will be given to `id`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the task.",
			},
			"resource_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the resource to filter tasks by.",
			},
			"list": schema.ListNestedAttribute{
				Description: "List of tasks.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the task.",
							Computed:    true,
						},
						"task_name": schema.StringAttribute{
							Description: "Name of the task.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the task.",
							Computed:    true,
						},
						"resource_name": schema.StringAttribute{
							Description: "Name of the resource related to this task.",
							Computed:    true,
							Optional:    true,
						},
						"resource_id": schema.StringAttribute{
							Description: "ID of the resource related to this task.",
							Computed:    true,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *tasksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*tdh.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *tasksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tasksDataSourceModel

	// Read Terraform configuration data into the model
	if resp.Diagnostics.Append(req.Config.Get(ctx, &state)...); resp.Diagnostics.HasError() {
		return
	}
	if d.validateInputs(state, &resp.Diagnostics).HasError() {
		return
	}

	if !state.Id.IsNull() {
		d.populateById(&ctx, &state, &resp.Diagnostics)
	} else {
		state.Id = types.StringValue(common.DataSource + common.TaskId)
		d.populateByFilters(&ctx, &state, &resp.Diagnostics)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
}

func (d *tasksDataSource) validateInputs(state tasksDataSourceModel, diag *diag.Diagnostics) *diag.Diagnostics {
	if state.Id.IsNull() && state.ResourceName.IsNull() {
		diag.AddError(
			"Invalid input",
			"At least one of `id` or `resource_name` is required",
		)
	}
	return diag
}

func (d *tasksDataSource) populateById(ctx *context.Context, state *tasksDataSourceModel, diag *diag.Diagnostics) {
	tflog.Info(*ctx, "populating by taskId: ", map[string]interface{}{"taskId": state.Id.ValueString()})
	var taskId = state.Id.ValueString()
	response, err := d.client.TaskService.GetTask(taskId)
	if err != nil {
		diag.AddError(
			"Unable to Read Task(s)",
			err.Error(),
		)
		return
	}
	tflog.Debug(*ctx, "READING task", map[string]interface{}{
		"task": response,
	})

	tfModel := d.convertToTfModel(*response)
	state.List = append(state.List, tfModel)
}

func (d *tasksDataSource) populateByFilters(ctx *context.Context, state *tasksDataSourceModel, diag *diag.Diagnostics) {
	tflog.Info(*ctx, "populating by filters: %s", map[string]interface{}{"state": state})
	query := &task.TasksQuery{
		ResourceName: state.ResourceName.ValueString(),
	}
	response, err := d.client.TaskService.GetTasks(query)
	if err != nil {
		diag.AddError(
			"Unable to Read Task(s)",
			err.Error(),
		)
		return
	}

	// Map DTO body to model
	for _, taskDto := range *response.Get() {
		tflog.Debug(*ctx, "READING task", map[string]interface{}{
			"task": taskDto,
		})
		tfModel := d.convertToTfModel(taskDto)
		tfModel.ResourceName = state.ResourceName
		state.List = append(state.List, tfModel)
	}

}

func (d *tasksDataSource) convertToTfModel(response model.Task) taskModel {
	var taskName = response.TaskType
	if response.DisplayName != "" {
		taskName = response.DisplayName
	}
	tfModel := taskModel{
		Id:       types.StringValue(response.Id),
		TaskName: types.StringValue(taskName),
		Status:   types.StringValue(response.Status),
	}
	if response.UiParams.ResourceName != "" {
		tfModel.ResourceName = types.StringValue(response.UiParams.ResourceName)
	}
	if response.UiParams.ResourceId != "" {
		tfModel.ResourceId = types.StringValue(response.UiParams.ResourceId)
	}
	return tfModel
}
