package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &blueprintDataSource{}

func NewBlueprintDataSource() datasource.DataSource {
	return &blueprintDataSource{}
}

type blueprintDataSource struct {
	client *client.Client
}

type blueprintSingleDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Type                 types.String `tfsdk:"type"`
	EnrollmentCode       types.String `tfsdk:"enrollment_code"`
	EnrollmentCodeActive types.Bool   `tfsdk:"enrollment_code_active"`
}

func (d *blueprintDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (d *blueprintDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get details for a specific blueprint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Blueprint.",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"enrollment_code": schema.StringAttribute{
				Computed: true,
			},
			"enrollment_code_active": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *blueprintDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *blueprintDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data blueprintSingleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var blueprint client.Blueprint
	err := d.client.DoRequest(ctx, "GET", "/api/v1/blueprints/"+data.ID.ValueString(), nil, &blueprint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read blueprint, got error: %s", err))
		return
	}

	data.Name = types.StringValue(blueprint.Name)
	data.Description = types.StringValue(blueprint.Description)
	data.Type = types.StringValue(blueprint.Type)
	data.EnrollmentCode = types.StringValue(blueprint.EnrollmentCode.Code)
	data.EnrollmentCodeActive = types.BoolValue(blueprint.EnrollmentCode.IsActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
