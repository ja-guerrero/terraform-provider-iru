package provider

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ resource.Resource = &customProfileResource{}
var _ resource.ResourceWithImportState = &customProfileResource{}
var _ resource.ResourceWithIdentity = &customProfileResource{}

func NewCustomProfileResource() resource.Resource {
	return &customProfileResource{}
}

type customProfileResource struct {
	client *client.Client
}

type customProfileResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Active        types.Bool   `tfsdk:"active"`
	ProfileFile   types.String `tfsdk:"profile_file"`
	MDMIdentifier types.String `tfsdk:"mdm_identifier"`
	RunsOnMac     types.Bool   `tfsdk:"runs_on_mac"`
	RunsOnIPhone  types.Bool   `tfsdk:"runs_on_iphone"`
	RunsOnIPad    types.Bool   `tfsdk:"runs_on_ipad"`
	RunsOnTV      types.Bool   `tfsdk:"runs_on_tv"`
	RunsOnVision  types.Bool   `tfsdk:"runs_on_vision"`
}

type customProfileResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *customProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_profile"
}

func (r *customProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Iru Custom Profile library item.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the Custom Profile.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the Custom Profile.",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether this library item is active.",
			},
			"profile_file": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The content of the `.mobileconfig` file. Must be a valid Apple Configuration Profile XML.",
				PlanModifiers: []planmodifier.String{
					profileFileEquivalentPlanModifier{},
				},
			},
			"mdm_identifier": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique MDM identifier (PayloadIdentifier) extracted from the profile.",
			},
			"runs_on_mac": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the profile runs on macOS.",
			},
			"runs_on_iphone": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the profile runs on iOS.",
			},
			"runs_on_ipad": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the profile runs on iPadOS.",
			},
			"runs_on_tv": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the profile runs on tvOS.",
			},
			"runs_on_vision": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the profile runs on visionOS.",
			},
		},
	}
}

func (r *customProfileResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The unique identifier for the Custom Profile.",
			},
		},
	}
}

func (r *customProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data customProfileResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fields := map[string]string{
		"name":           data.Name.ValueString(),
		"active":         fmt.Sprintf("%t", data.Active.ValueBool()),
		"runs_on_mac":    fmt.Sprintf("%t", data.RunsOnMac.ValueBool()),
		"runs_on_iphone": fmt.Sprintf("%t", data.RunsOnIPhone.ValueBool()),
		"runs_on_ipad":   fmt.Sprintf("%t", data.RunsOnIPad.ValueBool()),
		"runs_on_tv":     fmt.Sprintf("%t", data.RunsOnTV.ValueBool()),
		"runs_on_vision": fmt.Sprintf("%t", data.RunsOnVision.ValueBool()),
	}

	fileContent := []byte(data.ProfileFile.ValueString())

	var profileResponse client.CustomProfile
	err := r.client.DoMultipartRequest(ctx, "POST", "/api/v1/library/custom-profiles", fields, "file", "profile.mobileconfig", bytes.NewReader(fileContent), &profileResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create custom profile, got error: %s", err))
		return
	}

	r.updateModelWithResponse(&data, &profileResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := customProfileResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *customProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data customProfileResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var identity *customProfileResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" && identity != nil {
		id = identity.ID.ValueString()
	}

	var profileResponse client.CustomProfile
	err := r.client.DoRequest(ctx, "GET", "/api/v1/library/custom-profiles/"+id, nil, &profileResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom profile, got error: %s", err))
		return
	}

	r.updateModelWithResponse(&data, &profileResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	newIdentity := customProfileResourceIdentityModel{ID: data.ID}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &newIdentity)...)
}

func (r *customProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state customProfileResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fields := map[string]string{
		"name":           plan.Name.ValueString(),
		"active":         fmt.Sprintf("%t", plan.Active.ValueBool()),
		"runs_on_mac":    fmt.Sprintf("%t", plan.RunsOnMac.ValueBool()),
		"runs_on_iphone": fmt.Sprintf("%t", plan.RunsOnIPhone.ValueBool()),
		"runs_on_ipad":   fmt.Sprintf("%t", plan.RunsOnIPad.ValueBool()),
		"runs_on_tv":     fmt.Sprintf("%t", plan.RunsOnTV.ValueBool()),
		"runs_on_vision": fmt.Sprintf("%t", plan.RunsOnVision.ValueBool()),
	}

	var fileContent []byte
	var fileReader *bytes.Reader
	if !plan.ProfileFile.Equal(state.ProfileFile) {
		fileContent = []byte(plan.ProfileFile.ValueString())
		fileReader = bytes.NewReader(fileContent)
	}

	var profileResponse client.CustomProfile
	err := r.client.DoMultipartRequest(ctx, "PATCH", "/api/v1/library/custom-profiles/"+plan.ID.ValueString(), fields, "file", "profile.mobileconfig", fileReader, &profileResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update custom profile, got error: %s", err))
		return
	}

	r.updateModelWithResponse(&plan, &profileResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	identity := customProfileResourceIdentityModel{
		ID: plan.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *customProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data customProfileResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DoRequest(ctx, "DELETE", "/api/v1/library/custom-profiles/"+data.ID.ValueString(), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete custom profile, got error: %s", err))
		return
	}
}

func (r *customProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *customProfileResource) updateModelWithResponse(data *customProfileResourceModel, resp *client.CustomProfile) {
	data.ID = types.StringValue(resp.ID)
	data.Name = types.StringValue(resp.Name)
	data.Active = types.BoolValue(resp.Active)
	data.MDMIdentifier = types.StringValue(resp.MDMIdentifier)
	data.RunsOnMac = types.BoolValue(resp.RunsOnMac)
	data.RunsOnIPhone = types.BoolValue(resp.RunsOnIPhone)
	data.RunsOnIPad = types.BoolValue(resp.RunsOnIPad)
	data.RunsOnTV = types.BoolValue(resp.RunsOnTV)
	data.RunsOnVision = types.BoolValue(resp.RunsOnVision)

	if resp.Profile != "" {
		data.ProfileFile = types.StringValue(resp.Profile)
	}
}

// profileFileEquivalentPlanModifier suppresses a planned change to profile_file
// when the prior state and the configured value are semantically equivalent
// .mobileconfig XML — i.e. they differ only in whitespace, line endings, or
// inter-tag indentation. Real content changes still produce a diff.
type profileFileEquivalentPlanModifier struct{}

func (m profileFileEquivalentPlanModifier) Description(ctx context.Context) string {
	return "Suppresses diffs when state and config differ only in whitespace/formatting."
}

func (m profileFileEquivalentPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m profileFileEquivalentPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if normalizeProfileXML(req.StateValue.ValueString()) == normalizeProfileXML(req.ConfigValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

var interTagWhitespace = regexp.MustCompile(`>\s+<`)

func normalizeProfileXML(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = interTagWhitespace.ReplaceAllString(s, "><")
	return strings.TrimSpace(s)
}
