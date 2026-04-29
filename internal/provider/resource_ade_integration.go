package provider

import (
	"bytes"
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

var _ resource.Resource = &adeIntegrationResource{}
var _ resource.ResourceWithImportState = &adeIntegrationResource{}
var _ resource.ResourceWithIdentity = &adeIntegrationResource{}

func NewADEIntegrationResource() resource.Resource {
	return &adeIntegrationResource{}
}

type adeIntegrationResource struct {
	client *client.Client
}

type adeIntegrationResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	BlueprintID         types.String `tfsdk:"blueprint_id"`
	Phone               types.String `tfsdk:"phone"`
	Email               types.String `tfsdk:"email"`
	MDMServerTokenFile  types.String `tfsdk:"mdm_server_token_file"`
	AccessTokenExpiry   types.String `tfsdk:"access_token_expiry"`
	ServerName          types.String `tfsdk:"server_name"`
	ServerUUID          types.String `tfsdk:"server_uuid"`
	AdminID             types.String `tfsdk:"admin_id"`
	OrgName             types.String `tfsdk:"org_name"`
	STokenFileName      types.String `tfsdk:"stoken_file_name"`
	DaysLeft            types.Int64  `tfsdk:"days_left"`
	Status              types.String `tfsdk:"status"`
	UseBlueprintRouting types.Bool   `tfsdk:"use_blueprint_routing"`
}

type adeIntegrationResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *adeIntegrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ade_integration"
}

func (r *adeIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Automated Device Enrollment (ADE) integration with Apple Business Manager. This resource handles the MDM server token (.p7m) and enrollment settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the ADE Integration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"blueprint_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The UUID of the default blueprint to associate with the integration. Required if `use_blueprint_routing` is `false`.",
			},
			"phone": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A support phone number for the integration (shown to users during enrollment).",
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A support email address for the integration (shown to users during enrollment).",
			},
			"mdm_server_token_file": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The base64-encoded content of the MDM server token file (.p7m) downloaded from Apple Business Manager.",
			},
			"use_blueprint_routing": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to use Blueprint Routing for this integration. If `true`, `blueprint_id` must be null.",
			},
			"access_token_expiry": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The access token expiry date.",
			},
			"server_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the ADE server.",
			},
			"server_uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the ADE server.",
			},
			"admin_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The admin ID of the ADE integration.",
			},
			"org_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization name.",
			},
			"stoken_file_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the server token file.",
			},
			"days_left": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of days left before expiry.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the ADE integration.",
			},
		},
	}
}

func (r *adeIntegrationResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The unique identifier for the ADE Integration.",
			},
		},
	}
}

func (r *adeIntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *adeIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data adeIntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fields := map[string]string{
		"phone": data.Phone.ValueString(),
		"email": data.Email.ValueString(),
	}

	if !data.BlueprintID.IsNull() {
		fields["blueprint_id"] = data.BlueprintID.ValueString()
	}

	if !data.UseBlueprintRouting.IsNull() {
		if data.UseBlueprintRouting.ValueBool() {
			fields["use_blueprint_routing"] = "true"
		} else {
			fields["use_blueprint_routing"] = "false"
		}
	}

	fileContent := []byte(data.MDMServerTokenFile.ValueString())

	var adeResponse client.ADEIntegration
	err := r.client.DoMultipartRequest(ctx, "POST", "/api/v1/integrations/apple/ade/", fields, "file", "token.p7m", bytes.NewReader(fileContent), &adeResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create ADE integration, got error: %s", err))
		return
	}

	r.updateModelWithADEIntegration(&data, &adeResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := adeIntegrationResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *adeIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data adeIntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var identity *adeIntegrationResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" && identity != nil {
		id = identity.ID.ValueString()
	}

	var adeResponse client.ADEIntegration
	err := r.client.DoRequest(ctx, "GET", "/api/v1/integrations/apple/ade/"+id, nil, &adeResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ADE integration, got error: %s", err))
		return
	}

	r.updateModelWithADEIntegration(&data, &adeResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	newIdentity := adeIntegrationResourceIdentityModel{ID: data.ID}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &newIdentity)...)
}

func (r *adeIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state adeIntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.MDMServerTokenFile.Equal(state.MDMServerTokenFile) {
		// Token changed, use Renew endpoint
		fields := map[string]string{
			"phone": plan.Phone.ValueString(),
			"email": plan.Email.ValueString(),
		}
		if !plan.BlueprintID.IsNull() {
			fields["blueprint_id"] = plan.BlueprintID.ValueString()
		}
		if !plan.UseBlueprintRouting.IsNull() {
			if plan.UseBlueprintRouting.ValueBool() {
				fields["use_blueprint_routing"] = "true"
			} else {
				fields["use_blueprint_routing"] = "false"
			}
		}

		fileContent := []byte(plan.MDMServerTokenFile.ValueString())

		var adeResponse client.ADEIntegration
		err := r.client.DoMultipartRequest(ctx, "POST", "/api/v1/integrations/apple/ade/"+plan.ID.ValueString()+"/renew", fields, "file", "token.p7m", bytes.NewReader(fileContent), &adeResponse)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to renew ADE integration, got error: %s", err))
			return
		}

		r.updateModelWithADEIntegration(&plan, &adeResponse)
	} else {
		// Normal update
		updateRequest := client.ADEIntegration{
			Phone: plan.Phone.ValueString(),
			Email: plan.Email.ValueString(),
		}
		if !plan.BlueprintID.IsNull() {
			updateRequest.BlueprintID = plan.BlueprintID.ValueString()
		}
		if !plan.UseBlueprintRouting.IsNull() {
			updateRequest.UseBlueprintRouting = plan.UseBlueprintRouting.ValueBool()
		}

		var adeResponse client.ADEIntegration
		err := r.client.DoRequest(ctx, "PATCH", "/api/v1/integrations/apple/ade/"+plan.ID.ValueString(), updateRequest, &adeResponse)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update ADE integration, got error: %s", err))
			return
		}

		r.updateModelWithADEIntegration(&plan, &adeResponse)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	identity := adeIntegrationResourceIdentityModel{
		ID: plan.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *adeIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data adeIntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DoRequest(ctx, "DELETE", "/api/v1/integrations/apple/ade/"+data.ID.ValueString(), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete ADE integration, got error: %s", err))
		return
	}
}

func (r *adeIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *adeIntegrationResource) updateModelWithADEIntegration(data *adeIntegrationResourceModel, adeResponse *client.ADEIntegration) {
	data.ID = types.StringValue(adeResponse.ID)
	if adeResponse.Blueprint != nil {
		data.BlueprintID = types.StringValue(adeResponse.Blueprint.ID)
	}

	phone := adeResponse.Phone
	if phone == "" {
		phone = adeResponse.Defaults.Phone
	}
	email := adeResponse.Email
	if email == "" {
		email = adeResponse.Defaults.Email
	}

	data.Phone = types.StringValue(phone)
	data.Email = types.StringValue(email)
	data.AccessTokenExpiry = types.StringValue(adeResponse.AccessTokenExpiry)
	data.ServerName = types.StringValue(adeResponse.ServerName)
	data.ServerUUID = types.StringValue(adeResponse.ServerUUID)
	data.AdminID = types.StringValue(adeResponse.AdminID)
	data.OrgName = types.StringValue(adeResponse.OrgName)
	data.STokenFileName = types.StringValue(adeResponse.STokenFileName)
	data.DaysLeft = types.Int64Value(int64(adeResponse.DaysLeft))
	data.Status = types.StringValue(adeResponse.Status)
	data.UseBlueprintRouting = types.BoolValue(adeResponse.UseBlueprintRouting)
}
