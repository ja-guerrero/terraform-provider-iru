package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ resource.Resource = &blueprintLibraryItemResource{}

func NewBlueprintLibraryItemResource() resource.Resource {
	return &blueprintLibraryItemResource{}
}

type blueprintLibraryItemResource struct {
	client *client.Client
}

type blueprintLibraryItemResourceModel struct {
	ID               types.String `tfsdk:"id"`
	BlueprintID      types.String `tfsdk:"blueprint_id"`
	LibraryItemID    types.String `tfsdk:"library_item_id"`
	AssignmentNodeID types.String `tfsdk:"assignment_node_id"`
}

func (r *blueprintLibraryItemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_library_item"
}

func (r *blueprintLibraryItemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns a specific Library Item (App, Profile, Script, etc.) to a Blueprint. For Assignment Maps, an optional `assignment_node_id` can be specified to place the item in a specific node.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the assignment (composite key).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"blueprint_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The UUID of the blueprint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"library_item_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The UUID of the library item.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assignment_node_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The UUID of the assignment node (for Assignment Maps).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *blueprintLibraryItemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *blueprintLibraryItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data blueprintLibraryItemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bpID := data.BlueprintID.ValueString()
	itemID := data.LibraryItemID.ValueString()

	payload := map[string]string{
		"library_item_id": itemID,
	}
	if !data.AssignmentNodeID.IsNull() {
		payload["assignment_node_id"] = data.AssignmentNodeID.ValueString()
	}

	// Response is a list of IDs? Postman says: ["id1", "id2"...]
	var response []string
	err := r.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/blueprints/%s/assign-library-item", bpID), payload, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to assign library item, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", bpID, itemID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *blueprintLibraryItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data blueprintLibraryItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify assignment via List
	// GET /blueprints/{id}/library-items
	var items []client.BlueprintLibraryItem
	err := r.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/blueprints/%s/library-items", data.BlueprintID.ValueString()), nil, &items)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list blueprint items, got error: %s", err))
		return
	}

	found := false
	for _, item := range items {
		if item.ID == data.LibraryItemID.ValueString() {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *blueprintLibraryItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Not supported, requires replace (handled by schema)
}

func (r *blueprintLibraryItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data blueprintLibraryItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Attempting DELETE on library-items endpoint as best guess fix for Postman error
	err := r.client.DoRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/blueprints/%s/library-items/%s", data.BlueprintID.ValueString(), data.LibraryItemID.ValueString()), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove library item, got error: %s", err))
		return
	}
}
