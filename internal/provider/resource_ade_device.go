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

var _ resource.Resource = &adeDeviceResource{}
var _ resource.ResourceWithImportState = &adeDeviceResource{}
var _ resource.ResourceWithIdentity = &adeDeviceResource{}

func NewADEDeviceResource() resource.Resource {
	return &adeDeviceResource{}
}

type adeDeviceResource struct {
	client *client.Client
}

type adeDeviceResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	SerialNumber        types.String `tfsdk:"serial_number"`
	Model               types.String `tfsdk:"model"`
	Description         types.String `tfsdk:"description"`
	AssetTag            types.String `tfsdk:"asset_tag"`
	Color               types.String `tfsdk:"color"`
	BlueprintID         types.String `tfsdk:"blueprint_id"`
	UserID              types.String `tfsdk:"user_id"`
	DEPAccount          types.String `tfsdk:"dep_account"`
	DeviceFamily        types.String `tfsdk:"device_family"`
	OS                  types.String `tfsdk:"os"`
	ProfileStatus       types.String `tfsdk:"profile_status"`
	IsEnrolled          types.Bool   `tfsdk:"is_enrolled"`
	UseBlueprintRouting types.Bool   `tfsdk:"use_blueprint_routing"`
}

type adeDeviceResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *adeDeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ade_device"
}

func (r *adeDeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Automated Device Enrollment (ADE) device record. These records represent devices synced from Apple Business Manager before they are fully enrolled in MDM. This resource allows managing assignments like blueprint and user. Note: ADE Devices cannot be created via Terraform; they must be imported after syncing from Apple.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the ADE Device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"serial_number": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The serial number of the Device.",
			},
			"model": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The model of the Device.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The description of the Device.",
			},
			"asset_tag": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The asset tag of the Device.",
			},
			"color": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The color of the Device.",
			},
			"blueprint_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The UUID of the blueprint assigned to the Device. Triggering a change will update the assignment in Iru.",
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The UUID of the user assigned to the Device. Triggering a change will update the assignment in Iru.",
			},
			"dep_account": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the ADE/DEP integration this device belongs to.",
			},
			"device_family": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The device family (e.g., `Mac`, `iPhone`, `iPad`).",
			},
			"os": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The operating system type (e.g., `OSX`, `iOS`).",
			},
			"profile_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the ADE enrollment profile (e.g., `assigned`, `pushed`).",
			},
			"is_enrolled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the device has completed MDM enrollment.",
			},
			"use_blueprint_routing": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to use Blueprint Routing for this device. If `true`, `blueprint_id` must be null.",
			},
		},
	}
}

func (r *adeDeviceResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The unique identifier for the ADE Device.",
			},
		},
	}
}

func (r *adeDeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *adeDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"ADE Device Creation Not Supported",
		"ADE Devices cannot be created via Terraform. Please use 'terraform import' to manage existing ADE devices.",
	)
}

func (r *adeDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data adeDeviceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var identity *adeDeviceResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" && identity != nil {
		id = identity.ID.ValueString()
	}

	var deviceResponse client.ADEDevice
	err := r.client.DoRequest(ctx, "GET", "/api/v1/integrations/apple/ade/devices/"+id, nil, &deviceResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ADE device, got error: %s", err))
		return
	}

	r.updateModelWithResponse(&data, &deviceResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	newIdentity := adeDeviceResourceIdentityModel{ID: data.ID}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &newIdentity)...)
}

func (r *adeDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state adeDeviceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := map[string]interface{}{}
	if !plan.BlueprintID.Equal(state.BlueprintID) {
		updateRequest["blueprint_id"] = plan.BlueprintID.ValueString()
	}
	if !plan.AssetTag.Equal(state.AssetTag) {
		updateRequest["asset_tag"] = plan.AssetTag.ValueString()
	}
	if !plan.UserID.Equal(state.UserID) {
		updateRequest["user_id"] = plan.UserID.ValueString()
	}
	if !plan.UseBlueprintRouting.Equal(state.UseBlueprintRouting) {
		updateRequest["use_blueprint_routing"] = plan.UseBlueprintRouting.ValueBool()
	}

	if len(updateRequest) > 0 {
		var deviceResponse client.ADEDevice
		err := r.client.DoRequest(ctx, "PATCH", "/api/v1/integrations/apple/ade/devices/"+plan.ID.ValueString(), updateRequest, &deviceResponse)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update ADE device, got error: %s", err))
			return
		}

		r.updateModelWithResponse(&plan, &deviceResponse)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	identity := adeDeviceResourceIdentityModel{
		ID: plan.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *adeDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// ADE Devices cannot be deleted via API in a way that Terraform should manage here.
	// We just remove it from Terraform state.
}

func (r *adeDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *adeDeviceResource) updateModelWithResponse(data *adeDeviceResourceModel, resp *client.ADEDevice) {
	data.ID = types.StringValue(resp.ID)
	data.SerialNumber = types.StringValue(resp.SerialNumber)
	data.Model = types.StringValue(resp.Model)
	data.Description = types.StringValue(resp.Description)
	data.AssetTag = types.StringValue(resp.AssetTag)
	data.Color = types.StringValue(resp.Color)
	data.BlueprintID = types.StringValue(resp.BlueprintID)
	data.UserID = types.StringValue(resp.UserID)
	data.DEPAccount = types.StringValue(resp.DEPAccount)
	data.DeviceFamily = types.StringValue(resp.DeviceFamily)
	data.OS = types.StringValue(resp.OS)
	data.ProfileStatus = types.StringValue(resp.ProfileStatus)
	data.IsEnrolled = types.BoolValue(resp.IsEnrolled)
	data.UseBlueprintRouting = types.BoolValue(resp.UseBlueprintRouting)
}
