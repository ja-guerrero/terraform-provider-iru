package provider

import (
	"context"
	"fmt"
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

var _ resource.Resource = &deviceNoteResource{}
var _ resource.ResourceWithImportState = &deviceNoteResource{}
var _ resource.ResourceWithIdentity = &deviceNoteResource{}

func NewDeviceNoteResource() resource.Resource {
	return &deviceNoteResource{}
}

type deviceNoteResource struct {
	client *client.Client
}

type deviceNoteResourceModel struct {
	ID        types.String `tfsdk:"id"`
	DeviceID  types.String `tfsdk:"device_id"`
	Content   types.String `tfsdk:"content"`
	Author    types.String `tfsdk:"author"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type deviceNoteResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *deviceNoteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_note"
}

func (r *deviceNoteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Note attached to a specific Device. Notes are useful for tracking internal administrative metadata about a device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the Note.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The UUID of the device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The content of the note.",
			},
			"author": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The author of the note.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When the note was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When the note was last updated.",
			},
		},
	}
}

func (r *deviceNoteResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The unique identifier for the Note (format: device_id:note_id).",
			},
		},
	}
}

func (r *deviceNoteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *deviceNoteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data deviceNoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	payload := map[string]string{
		"content": data.Content.ValueString(),
	}

	var noteResp client.DeviceNote
	err := r.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/notes", deviceID), payload, &noteResp)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create device note, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", deviceID, noteResp.ID))
	data.Author = types.StringValue(noteResp.Author)
	data.CreatedAt = types.StringValue(noteResp.CreatedAt)
	data.UpdatedAt = types.StringValue(noteResp.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := deviceNoteResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *deviceNoteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data deviceNoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "The ID must be in the format device_id:note_id")
		return
	}
	deviceID := idParts[0]
	noteID := idParts[1]

	var noteResp client.DeviceNote
	err := r.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/notes/%s", deviceID, noteID), nil, &noteResp)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device note, got error: %s", err))
		return
	}

	data.DeviceID = types.StringValue(deviceID)
	data.Content = types.StringValue(noteResp.Content)
	data.Author = types.StringValue(noteResp.Author)
	data.CreatedAt = types.StringValue(noteResp.CreatedAt)
	data.UpdatedAt = types.StringValue(noteResp.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *deviceNoteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data deviceNoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	deviceID := idParts[0]
	noteID := idParts[1]

	payload := map[string]string{
		"content": data.Content.ValueString(),
	}

	var noteResp client.DeviceNote
	err := r.client.DoRequest(ctx, "PATCH", fmt.Sprintf("/api/v1/devices/%s/notes/%s", deviceID, noteID), payload, &noteResp)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update device note, got error: %s", err))
		return
	}

	data.UpdatedAt = types.StringValue(noteResp.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *deviceNoteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data deviceNoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	deviceID := idParts[0]
	noteID := idParts[1]

	err := r.client.DoRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/devices/%s/notes/%s", deviceID, noteID), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete device note, got error: %s", err))
		return
	}
}

func (r *deviceNoteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
