package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ resource.Resource = &customAppResource{}
var _ resource.ResourceWithImportState = &customAppResource{}
var _ resource.ResourceWithIdentity = &customAppResource{}

func NewCustomAppResource() resource.Resource {
	return &customAppResource{}
}

type customAppResource struct {
	client *client.Client
}

type customAppResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	FileKey                types.String `tfsdk:"file_key"`
	InstallType            types.String `tfsdk:"install_type"`
	InstallEnforcement     types.String `tfsdk:"install_enforcement"`
	UnzipLocation          types.String `tfsdk:"unzip_location"`
	AuditScript            types.String `tfsdk:"audit_script"`
	PreinstallScript       types.String `tfsdk:"preinstall_script"`
	PostinstallScript      types.String `tfsdk:"postinstall_script"`
	ShowInSelfService      types.Bool   `tfsdk:"show_in_self_service"`
	SelfServiceCategoryID  types.String `tfsdk:"self_service_category_id"`
	SelfServiceRecommended types.Bool   `tfsdk:"self_service_recommended"`
	Active                 types.Bool   `tfsdk:"active"`
	Restart                types.Bool   `tfsdk:"restart"`
}

type customAppResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *customAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_app"
}

func (r *customAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Iru Custom App library item (PKG, ZIP, or IMG). This resource defines the installation and enforcement settings for custom software. Note: You must handle the file upload to S3 out-of-band and provide the `file_key` obtained from the Iru upload endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the Custom App.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name for this Custom App.",
			},
			"file_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The S3 key from the upload endpoint.",
			},
			"install_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Options: package, zip, image.",
			},
			"install_enforcement": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Options: install_once, continuously_enforce, no_enforcement.",
			},
			"unzip_location": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Required for install_type=zip.",
			},
			"audit_script": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Required for install_enforcement=continuously_enforce.",
			},
			"preinstall_script": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Script content to run before the application is installed.",
			},
			"postinstall_script": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Script content to run after the application is installed.",
			},
			"show_in_self_service": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to display this app in the Self Service catalog.",
			},
			"self_service_category_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The UUID of the Self Service category to display the app in. Required if `show_in_self_service` is `true`.",
			},
			"self_service_recommended": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to flag this app as recommended in Self Service.",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether this Custom App is active and available for installation.",
			},
			"restart": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to prompt for or force a restart after successful installation.",
			},
		},
	}
}

func (r *customAppResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The unique identifier for the Custom App.",
			},
		},
	}
}

func (r *customAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *customAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data customAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appRequest := r.mapToClient(&data)
	var appResponse client.CustomApp
	err := r.client.DoRequest(ctx, "POST", "/api/v1/library/custom-apps", appRequest, &appResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create custom app, got error: %s", err))
		return
	}

	r.mapFromClient(&data, &appResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := customAppResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *customAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data customAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var identity *customAppResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" && identity != nil {
		id = identity.ID.ValueString()
	}

	var appResponse client.CustomApp
	err := r.client.DoRequest(ctx, "GET", "/api/v1/library/custom-apps/"+id, nil, &appResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom app, got error: %s", err))
		return
	}

	r.mapFromClient(&data, &appResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	newIdentity := customAppResourceIdentityModel{ID: data.ID}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &newIdentity)...)
}

func (r *customAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data customAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appRequest := r.mapToClient(&data)
	var appResponse client.CustomApp
	err := r.client.DoRequest(ctx, "PATCH", "/api/v1/library/custom-apps/"+data.ID.ValueString(), appRequest, &appResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update custom app, got error: %s", err))
		return
	}

	r.mapFromClient(&data, &appResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := customAppResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *customAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data customAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DoRequest(ctx, "DELETE", "/api/v1/library/custom-apps/"+data.ID.ValueString(), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete custom app, got error: %s", err))
		return
	}
}

func (r *customAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *customAppResource) mapToClient(data *customAppResourceModel) client.CustomApp {
	return client.CustomApp{
		Name:                   data.Name.ValueString(),
		FileKey:                data.FileKey.ValueString(),
		InstallType:            data.InstallType.ValueString(),
		InstallEnforcement:     data.InstallEnforcement.ValueString(),
		UnzipLocation:          data.UnzipLocation.ValueString(),
		AuditScript:            data.AuditScript.ValueString(),
		PreinstallScript:       data.PreinstallScript.ValueString(),
		PostinstallScript:      data.PostinstallScript.ValueString(),
		ShowInSelfService:      data.ShowInSelfService.ValueBool(),
		SelfServiceCategoryID:  data.SelfServiceCategoryID.ValueString(),
		SelfServiceRecommended: data.SelfServiceRecommended.ValueBool(),
		Active:                 data.Active.ValueBool(),
		Restart:                data.Restart.ValueBool(),
	}
}

func (r *customAppResource) mapFromClient(data *customAppResourceModel, resp *client.CustomApp) {
	data.ID = types.StringValue(resp.ID)
	data.Name = types.StringValue(resp.Name)
	data.FileKey = types.StringValue(resp.FileKey)
	data.InstallType = types.StringValue(resp.InstallType)
	data.InstallEnforcement = types.StringValue(resp.InstallEnforcement)
	data.UnzipLocation = types.StringValue(resp.UnzipLocation)
	data.AuditScript = types.StringValue(resp.AuditScript)
	data.PreinstallScript = types.StringValue(resp.PreinstallScript)
	data.PostinstallScript = types.StringValue(resp.PostinstallScript)
	data.ShowInSelfService = types.BoolValue(resp.ShowInSelfService)
	data.SelfServiceCategoryID = types.StringValue(resp.SelfServiceCategoryID)
	data.SelfServiceRecommended = types.BoolValue(resp.SelfServiceRecommended)
	data.Active = types.BoolValue(resp.Active)
	data.Restart = types.BoolValue(resp.Restart)
}
