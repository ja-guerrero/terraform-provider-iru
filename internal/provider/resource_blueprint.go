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

var _ resource.Resource = &blueprintResource{}
var _ resource.ResourceWithImportState = &blueprintResource{}
var _ resource.ResourceWithIdentity = &blueprintResource{}

func NewBlueprintResource() resource.Resource {
	return &blueprintResource{}
}

type blueprintResource struct {
	client *client.Client
}

type blueprintResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Icon                 types.String `tfsdk:"icon"`
	Color                types.String `tfsdk:"color"`
	Type                 types.String `tfsdk:"type"`
	EnrollmentCode       types.String `tfsdk:"enrollment_code"`
	EnrollmentCodeActive types.Bool   `tfsdk:"enrollment_code_active"`
	SourceID             types.String `tfsdk:"source_id"`
	SourceType           types.String `tfsdk:"source_type"`
}

type blueprintResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *blueprintResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (r *blueprintResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Iru Blueprint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the Blueprint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the Blueprint.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description of the Blueprint.",
			},
			"icon": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The icon of the Blueprint (e.g., 'ss-files').",
			},
			"color": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The color of the Blueprint (e.g., 'aqua-800').",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The type of the Blueprint. Options: `classic`, `map`. Classic blueprints are standard lists of library items, while maps allow for conditional assignment logic.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enrollment_code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The enrollment code for the Blueprint.",
			},
			"enrollment_code_active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the enrollment code is active.",
			},
			"source_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the source blueprint to clone from. Only used during creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The type of the source blueprint to clone from. Only used during creation. Options: `classic`, `map`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *blueprintResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The unique identifier for the Blueprint.",
			},
		},
	}
}

func (r *blueprintResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *blueprintResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data blueprintResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	blueprintRequest := client.Blueprint{
		Name: data.Name.ValueString(),
	}
	blueprintRequest.EnrollmentCode.IsActive = data.EnrollmentCodeActive.ValueBool()
	if !data.Type.IsNull() {
		blueprintRequest.Type = data.Type.ValueString()
	}
	if !data.SourceID.IsNull() {
		blueprintRequest.Source.ID = data.SourceID.ValueString()
	}
	if !data.SourceType.IsNull() {
		blueprintRequest.Source.Type = data.SourceType.ValueString()
	}

	var blueprintResponse client.Blueprint
	err := r.client.DoRequest(ctx, "POST", "/api/v1/blueprints", blueprintRequest, &blueprintResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create blueprint, got error: %s", err))
		return
	}

	r.updateModelWithBlueprint(&data, &blueprintResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := blueprintResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *blueprintResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data blueprintResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var identity *blueprintResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" && identity != nil {
		id = identity.ID.ValueString()
	}

	var blueprintResponse client.Blueprint
	err := r.client.DoRequest(ctx, "GET", "/api/v1/blueprints/"+id, nil, &blueprintResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read blueprint, got error: %s", err))
		return
	}

	r.updateModelWithBlueprint(&data, &blueprintResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	newIdentity := blueprintResourceIdentityModel{ID: data.ID}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &newIdentity)...)
}

func (r *blueprintResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data blueprintResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	blueprintRequest := client.Blueprint{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Icon:        data.Icon.ValueString(),
		Color:       data.Color.ValueString(),
	}
	blueprintRequest.EnrollmentCode.IsActive = data.EnrollmentCodeActive.ValueBool()

	var blueprintResponse client.Blueprint
	err := r.client.DoRequest(ctx, "PATCH", "/api/v1/blueprints/"+data.ID.ValueString(), blueprintRequest, &blueprintResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update blueprint, got error: %s", err))
		return
	}

	r.updateModelWithBlueprint(&data, &blueprintResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := blueprintResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *blueprintResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data blueprintResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DoRequest(ctx, "DELETE", "/api/v1/blueprints/"+data.ID.ValueString(), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete blueprint, got error: %s", err))
		return
	}
}

func (r *blueprintResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *blueprintResource) updateModelWithBlueprint(data *blueprintResourceModel, blueprintResponse *client.Blueprint) {
	data.ID = types.StringValue(blueprintResponse.ID)
	data.Name = types.StringValue(blueprintResponse.Name)
	data.Description = types.StringValue(blueprintResponse.Description)
	data.Icon = types.StringValue(blueprintResponse.Icon)
	data.Color = types.StringValue(blueprintResponse.Color)
	data.Type = types.StringValue(blueprintResponse.Type)
	data.EnrollmentCode = types.StringValue(blueprintResponse.EnrollmentCode.Code)
	data.EnrollmentCodeActive = types.BoolValue(blueprintResponse.EnrollmentCode.IsActive)
}
