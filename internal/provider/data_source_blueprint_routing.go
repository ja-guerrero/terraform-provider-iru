package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &blueprintRoutingDataSource{}

func NewBlueprintRoutingDataSource() datasource.DataSource {
	return &blueprintRoutingDataSource{}
}

type blueprintRoutingDataSource struct {
	client *client.Client
}

type blueprintRoutingDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	EnrollmentCode       types.String `tfsdk:"enrollment_code"`
	EnrollmentCodeActive types.Bool   `tfsdk:"enrollment_code_active"`
}

func (d *blueprintRoutingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_routing"
}

func (d *blueprintRoutingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides Blueprint Routing settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A fixed identifier for the singleton data source.",
			},
			"enrollment_code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The enrollment code for Blueprint Routing.",
			},
			"enrollment_code_active": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the enrollment code for Blueprint Routing is active.",
			},
		},
	}
}

func (d *blueprintRoutingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *blueprintRoutingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data blueprintRoutingDataSourceModel

	var response client.BlueprintRouting
	err := d.client.DoRequest(ctx, "GET", "/api/v1/blueprint-routing/", nil, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read blueprint routing, got error: %s", err))
		return
	}

	data.ID = types.StringValue("blueprint_routing_settings")
	data.EnrollmentCode = types.StringValue(response.EnrollmentCode.Code)
	data.EnrollmentCodeActive = types.BoolValue(response.EnrollmentCode.IsActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
