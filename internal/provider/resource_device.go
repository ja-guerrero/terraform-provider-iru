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

var _ resource.Resource = &deviceResource{}
var _ resource.ResourceWithImportState = &deviceResource{}
var _ resource.ResourceWithIdentity = &deviceResource{}

func NewDeviceResource() resource.Resource {
	return &deviceResource{}
}

type deviceResource struct {
	client *client.Client
}

type deviceResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DeviceName  types.String `tfsdk:"device_name"`
	AssetTag    types.String `tfsdk:"asset_tag"`
	BlueprintID types.String `tfsdk:"blueprint_id"`
	UserID      types.String `tfsdk:"user_id"`
	// Read-only fields
	SerialNumber types.String `tfsdk:"serial_number"`
	Model        types.String `tfsdk:"model"`
	OSVersion    types.String `tfsdk:"os_version"`
	Platform     types.String `tfsdk:"platform"`
}

type deviceResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *deviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *deviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Iru Device. Note: Devices cannot be created via Terraform, only imported and managed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the Device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the Device. Note: To rename a device, use the `iru_device_action_set_name` action.",
			},
			"asset_tag": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The custom asset tag assigned to the Device.",
			},
			"blueprint_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The UUID of the blueprint assigned to the Device. Changing this will trigger a blueprint move.",
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The UUID of the user assigned to the Device.",
			},
			"serial_number": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The serial number of the Device.",
			},
			"model": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The model of the Device.",
			},
			"os_version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The OS version of the Device.",
			},
			"platform": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The platform of the Device.",
			},
		},
	}
}

func (r *deviceResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The unique identifier for the Device.",
			},
		},
	}
}

func (r *deviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Device Creation Not Supported",
		"Devices cannot be created via Terraform. Please use 'terraform import' to manage existing devices.",
	)
}

func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data deviceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var identity *deviceResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" && identity != nil {
		id = identity.ID.ValueString()
	}

	var deviceResponse client.Device
	err := r.client.DoRequest(ctx, "GET", "/api/v1/devices/"+id, nil, &deviceResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device, got error: %s", err))
		return
	}

	data.DeviceName = types.StringValue(deviceResponse.DeviceName)
	data.AssetTag = types.StringValue(deviceResponse.AssetTag)
	data.BlueprintID = types.StringValue(deviceResponse.BlueprintID)
	data.UserID = types.StringValue(deviceResponse.UserID)
	data.SerialNumber = types.StringValue(deviceResponse.SerialNumber)
	data.Model = types.StringValue(deviceResponse.Model)
	data.OSVersion = types.StringValue(deviceResponse.OSVersion)
	data.Platform = types.StringValue(deviceResponse.Platform)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	newIdentity := deviceResourceIdentityModel{ID: data.ID}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &newIdentity)...)
}

func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state deviceResourceModel

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

	if len(updateRequest) > 0 {
		var deviceResponse client.Device
		err := r.client.DoRequest(ctx, "PATCH", "/api/v1/devices/"+plan.ID.ValueString(), updateRequest, &deviceResponse)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update device, got error: %s", err))
			return
		}

		// Update state with response values
		plan.DeviceName = types.StringValue(deviceResponse.DeviceName)
		plan.AssetTag = types.StringValue(deviceResponse.AssetTag)
		plan.BlueprintID = types.StringValue(deviceResponse.BlueprintID)
		plan.UserID = types.StringValue(deviceResponse.UserID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	identity := deviceResourceIdentityModel{
		ID: plan.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data deviceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DoRequest(ctx, "DELETE", "/api/v1/devices/"+data.ID.ValueString(), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete device, got error: %s", err))
		return
	}
}

func (r *deviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
