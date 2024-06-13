package SRE

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/auth"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &smtpResource{}
	_ resource.ResourceWithConfigure   = &smtpResource{}
	_ resource.ResourceWithImportState = &smtpResource{}
)

func NewSmtpResource() resource.Resource {
	return &certificateResource{}
}

type smtpResource struct {
	client *tdh.Client
}

type smtpResourceModal struct {
	Host            types.String `tfsdk:"host"`
	Port            types.String `tfsdk:"port"`
	From            types.String `tfsdk:"from"`
	UserName        types.String `tfsdk:"user_name"`
	password        types.String `tfsdk:"password"`
	confirmPassword types.String `tfsdk:"confirmPassword"`
	tlsEnabled      types.String `tfsdk:"tls"`
	auth            types.String `tfsdk:"auth"`
}

func (r *smtpResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
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
		MarkdownDescription: "Represents SMTP details. We can only edit the smtp details",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Auto-generated ID after creating a certificate, and can be passed to import an existing user from TDH to terraform state.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host": schema.StringAttribute{
				Description: "SMTP - Host Name",
				Required:    true,
			},
			"port": schema.StringAttribute{
				Description: "SMTP - port details",
				Required:    true,
			},
			"From": schema.StringAttribute{
				Description: "SMTP - Email Address",
				Required:    true,
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "SMTP- User name",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "SMTP- passowrd",
				Computed:            true,
			},
			"confirm_password": schema.StringAttribute{
				MarkdownDescription: "SMTP- confirm passowrd",
				Required:            true,
			},
			"tls": schema.StringAttribute{
				MarkdownDescription: "TLS Enabled",
				Required:            true,
			},
			"auth": schema.StringAttribute{
				MarkdownDescription: "Authentication enabled flag",
				Required:            true,
			},
		},
	}

	tflog.Info(ctx, "END__Schema")
}

// Create a new resource
func (r *smtpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	return
}

func (r *smtpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "INIT__Update")

	// Retrieve values from plan
	var state smtpResourceModal
	diags := req.Plan.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	smtpUpdateReq := auth.SmtpRequest{
		Host:            state.Host.ValueString(),
		Port:            state.Port.ValueString(),
		From:            state.From.ValueString(),
		UserName:        state.UserName.ValueString(),
		Password:        state.password.ValueString(),
		ConfirmPassword: state.confirmPassword.ValueString(),
		TlsEnabled:      state.tlsEnabled.ValueString(),
		Auth:            state.auth.ValueString(),
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
	resource.ImportStatePassthroughID(ctx, path.Root("port"), req, resp)
}
func (r *smtpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	return
}
