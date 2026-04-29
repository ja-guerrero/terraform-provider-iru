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

var _ resource.Resource = &customScriptResource{}
var _ resource.ResourceWithImportState = &customScriptResource{}
var _ resource.ResourceWithIdentity = &customScriptResource{}

func NewCustomScriptResource() resource.Resource {
	return &customScriptResource{}
}

type customScriptResource struct {
	client *client.Client
}

type customScriptResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Active             types.Bool   `tfsdk:"active"`
	ExecutionFrequency types.String `tfsdk:"execution_frequency"`
	Restart            types.Bool   `tfsdk:"restart"`
	Script             types.String `tfsdk:"script"`
	RemediationScript  types.String `tfsdk:"remediation_script"`
	ShowInSelfService  types.Bool   `tfsdk:"show_in_self_service"`
}

type customScriptResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *customScriptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_script"
}

func (r *customScriptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Iru Custom Script library item.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the Custom Script.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the Custom Script.",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether this library item is active.",
			},
			"execution_frequency": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The frequency at which the script is enforced. Options: `once` (runs once and never again), `every_15_min`, `every_day`, `no_enforcement` (manual run only).",
			},
			"restart": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to restart the computer if the script execution is successful. Use with caution as this may disrupt users.",
			},
			"script": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The shell or zsh script content. Must include a valid shebang (e.g., `#!/bin/zsh`).",
			},
			"remediation_script": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "An optional script that runs only if the primary script fails (exits non-zero).",
			},
			"show_in_self_service": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to make this script available for users to run manually in the Self Service app.",
			},
		},
	}
}

func (r *customScriptResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The unique identifier for the Custom Script.",
			},
		},
	}
}

func (r *customScriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data customScriptResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scriptRequest := client.CustomScript{
		Name:               data.Name.ValueString(),
		Active:             data.Active.ValueBool(),
		ExecutionFrequency: data.ExecutionFrequency.ValueString(),
		Restart:            data.Restart.ValueBool(),
		Script:             data.Script.ValueString(),
		RemediationScript:  data.RemediationScript.ValueString(),
		ShowInSelfService:  data.ShowInSelfService.ValueBool(),
	}

	var scriptResponse client.CustomScript
	err := r.client.DoRequest(ctx, "POST", "/api/v1/library/custom-scripts", scriptRequest, &scriptResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create custom script, got error: %s", err))
		return
	}

	data.ID = types.StringValue(scriptResponse.ID)
	data.Name = types.StringValue(scriptResponse.Name)
	data.Active = types.BoolValue(scriptResponse.Active)
	data.ExecutionFrequency = types.StringValue(scriptResponse.ExecutionFrequency)
	data.Restart = types.BoolValue(scriptResponse.Restart)
	data.Script = types.StringValue(scriptResponse.Script)
	data.RemediationScript = types.StringValue(scriptResponse.RemediationScript)
	data.ShowInSelfService = types.BoolValue(scriptResponse.ShowInSelfService)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := customScriptResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *customScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data customScriptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var identity *customScriptResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" && identity != nil {
		id = identity.ID.ValueString()
	}

	var scriptResponse client.CustomScript
	err := r.client.DoRequest(ctx, "GET", "/api/v1/library/custom-scripts/"+id, nil, &scriptResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom script, got error: %s", err))
		return
	}

	data.Name = types.StringValue(scriptResponse.Name)
	data.Active = types.BoolValue(scriptResponse.Active)
	data.ExecutionFrequency = types.StringValue(scriptResponse.ExecutionFrequency)
	data.Restart = types.BoolValue(scriptResponse.Restart)
	data.Script = types.StringValue(scriptResponse.Script)
	data.RemediationScript = types.StringValue(scriptResponse.RemediationScript)
	data.ShowInSelfService = types.BoolValue(scriptResponse.ShowInSelfService)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	newIdentity := customScriptResourceIdentityModel{ID: data.ID}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &newIdentity)...)
}

func (r *customScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data customScriptResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scriptRequest := client.CustomScript{
		Name:               data.Name.ValueString(),
		Active:             data.Active.ValueBool(),
		ExecutionFrequency: data.ExecutionFrequency.ValueString(),
		Restart:            data.Restart.ValueBool(),
		Script:             data.Script.ValueString(),
		RemediationScript:  data.RemediationScript.ValueString(),
		ShowInSelfService:  data.ShowInSelfService.ValueBool(),
	}

	var scriptResponse client.CustomScript
	err := r.client.DoRequest(ctx, "PATCH", "/api/v1/library/custom-scripts/"+data.ID.ValueString(), scriptRequest, &scriptResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update custom script, got error: %s", err))
		return
	}

	data.Name = types.StringValue(scriptResponse.Name)
	data.Active = types.BoolValue(scriptResponse.Active)
	data.ExecutionFrequency = types.StringValue(scriptResponse.ExecutionFrequency)
	data.Restart = types.BoolValue(scriptResponse.Restart)
	data.Script = types.StringValue(scriptResponse.Script)
	data.RemediationScript = types.StringValue(scriptResponse.RemediationScript)
	data.ShowInSelfService = types.BoolValue(scriptResponse.ShowInSelfService)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	identity := customScriptResourceIdentityModel{
		ID: data.ID,
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *customScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data customScriptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DoRequest(ctx, "DELETE", "/api/v1/library/custom-scripts/"+data.ID.ValueString(), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete custom script, got error: %s", err))
		return
	}
}

func (r *customScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
