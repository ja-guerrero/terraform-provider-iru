package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ resource.Resource = &blueprintRoutingResource{}

func NewBlueprintRoutingResource() resource.Resource {
	return &blueprintRoutingResource{}
}

type blueprintRoutingResource struct {
	client *client.Client
}

type blueprintRoutingResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	EnrollmentCode       types.String `tfsdk:"enrollment_code"`
	EnrollmentCodeActive types.Bool   `tfsdk:"enrollment_code_active"`
}

func (r *blueprintRoutingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_routing"
}

func (r *blueprintRoutingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages global Blueprint Routing settings. Blueprint Routing allows devices to be automatically assigned to blueprints based on rules. This is a singleton resource that manages the enrollment code and its activation state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A fixed identifier for the singleton resource.",
			},
			"enrollment_code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The enrollment code for Blueprint Routing.",
			},
			"enrollment_code_active": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether the enrollment code for Blueprint Routing is active.",
			},
		},
	}
}

func (r *blueprintRoutingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *blueprintRoutingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data blueprintRoutingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := client.BlueprintRouting{}
	updateRequest.EnrollmentCode.IsActive = data.EnrollmentCodeActive.ValueBool()

	var response client.BlueprintRouting
	err := r.client.DoRequest(ctx, "PATCH", "/api/v1/blueprint-routing/", updateRequest, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update blueprint routing, got error: %s", err))
		return
	}

	data.ID = types.StringValue("blueprint_routing_settings")
	data.EnrollmentCode = types.StringValue(response.EnrollmentCode.Code)
	data.EnrollmentCodeActive = types.BoolValue(response.EnrollmentCode.IsActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *blueprintRoutingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data blueprintRoutingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var response client.BlueprintRouting
	err := r.client.DoRequest(ctx, "GET", "/api/v1/blueprint-routing/", nil, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read blueprint routing, got error: %s", err))
		return
	}

	data.ID = types.StringValue("blueprint_routing_settings")
	data.EnrollmentCode = types.StringValue(response.EnrollmentCode.Code)
	data.EnrollmentCodeActive = types.BoolValue(response.EnrollmentCode.IsActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *blueprintRoutingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data blueprintRoutingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := client.BlueprintRouting{}
	updateRequest.EnrollmentCode.IsActive = data.EnrollmentCodeActive.ValueBool()

	var response client.BlueprintRouting
	err := r.client.DoRequest(ctx, "PATCH", "/api/v1/blueprint-routing/", updateRequest, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update blueprint routing, got error: %s", err))
		return
	}

	data.ID = types.StringValue("blueprint_routing_settings")
	data.EnrollmentCode = types.StringValue(response.EnrollmentCode.Code)
	data.EnrollmentCodeActive = types.BoolValue(response.EnrollmentCode.IsActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *blueprintRoutingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Deleting the resource will just disable it
	updateRequest := client.BlueprintRouting{}
	updateRequest.EnrollmentCode.IsActive = false

	err := r.client.DoRequest(ctx, "PATCH", "/api/v1/blueprint-routing/", updateRequest, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable blueprint routing during delete, got error: %s", err))
		return
	}
}
