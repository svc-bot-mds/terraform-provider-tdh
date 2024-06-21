package tdh

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/auth"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &smtpResource{}
	_ resource.ResourceWithConfigure   = &smtpResource{}
	_ resource.ResourceWithImportState = &smtpResource{}
)

func NewSmtpResource() resource.Resource {
	return &smtpResource{}
}

type smtpResource struct {
	client *tdh.Client
}

type SmtpResourceModal struct {
	ID         types.String `tfsdk:"id"`
	Host       types.String `tfsdk:"host"`
	Port       types.String `tfsdk:"port"`
	From       types.String `tfsdk:"from"`
	UserName   types.String `tfsdk:"user_name"`
	Password   types.String `tfsdk:"password"`
	TlsEnabled types.String `tfsdk:"tls"`
	Auth       types.String `tfsdk:"auth"`
}

func (r *smtpResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smtp"
}

func (r *smtpResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tdh.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tdh.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *smtpResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "INIT__Schema")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Represents SMTP details. We can only edit the smtp details.<br>" +
			"**Note:** For SRE only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for this SMTP",
				Computed:    true,
			},
			"host": schema.StringAttribute{
				Description: "SMTP - Host Name",
				Required:    true,
			},
			"port": schema.StringAttribute{
				Description: "SMTP - Port. Can be passed to import an existing smtp details from TDH to terraform state during the update.",
				Required:    true,
			},
			"from": schema.StringAttribute{
				Description: "SMTP - Email Address",
				Required:    true,
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "SMTP - User name",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "SMTP - Password",
				Required:            true,
			},
			"tls": schema.StringAttribute{
				MarkdownDescription: "Whether TLS is enabled or not",
				Required:            true,
			},
			"auth": schema.StringAttribute{
				MarkdownDescription: "Whether authentication is enabled or not",
				Required:            true,
			},
		},
	}

	tflog.Info(ctx, "END__Schema")
}

// Create a new resource
func (r *smtpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "INIT__Create")
	// Retrieve values from plan
	var plan SmtpResourceModal
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	smtpRequest := &auth.SmtpRequest{
		Host:       plan.Host.ValueString(),
		Auth:       plan.Auth.ValueString(),
		UserName:   plan.UserName.ValueString(),
		From:       plan.From.ValueString(),
		Password:   plan.Password.ValueString(),
		TlsEnabled: plan.TlsEnabled.ValueString(),
		Port:       plan.Port.ValueString(),
	}

	tflog.Info(ctx, "req param", map[string]interface{}{"requestbody": smtpRequest})
	smtp, err := r.client.Auth.CreateSmtpDetails(*smtpRequest)
	if err != nil {
		apiErr := core.ApiError{}
		errors.As(err, &apiErr)
		resp.Diagnostics.AddError(
			"Submitting request to create Smtp",
			"There was some issue while creating the smtp."+
				" Unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	if saveFromSmtpResponse(&plan, &smtp) != 0 {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Create")
}

func (r *smtpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")

	// Retrieve values from plan
	var state SmtpResourceModal
	diags := req.Plan.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	smtpUpdateReq := auth.SmtpRequest{
		Host:       state.Host.ValueString(),
		Port:       state.Port.ValueString(),
		From:       state.From.ValueString(),
		UserName:   state.UserName.ValueString(),
		Password:   state.Password.ValueString(),
		TlsEnabled: state.TlsEnabled.ValueString(),
		Auth:       state.Auth.ValueString(),
	}

	// Update existing svc account
	if _, err := r.client.Auth.UpdateSmtpDetails(smtpUpdateReq); err != nil {
		resp.Diagnostics.AddError(
			"Updating the SMTP details",
			"Could not update smtp details, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Update")
}

func (r *smtpResource) Delete(ctx context.Context, request resource.DeleteRequest, resp *resource.DeleteResponse) {
	return
}

func (r *smtpResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *smtpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "INIT__Read")
	// Get current state
	var state SmtpResourceModal
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed certificate value from TDH
	smtp, err := r.client.Auth.GetSmtpDetails()
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading TDH SMTP Details",
			"Could not read TDH SMTP Details  "+err.Error(),
		)
		return
	}

	if saveFromSmtpResponse(&state, &smtp) != 0 {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "END__Read")
}

func saveFromSmtpResponse(state *SmtpResourceModal,
	smtpResponse *model.Smtp) int8 {
	state.From = types.StringValue(smtpResponse.FromEmail)
	state.Port = types.StringValue(smtpResponse.Port)
	state.Host = types.StringValue(smtpResponse.Host)
	state.UserName = types.StringValue(smtpResponse.UserName)
	state.TlsEnabled = types.StringValue(smtpResponse.Tls)
	state.Auth = types.StringValue(smtpResponse.Auth)

	return 0
}
