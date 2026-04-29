package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &prismCountDataSource{}

func NewPrismCountDataSource() datasource.DataSource {
	return &prismCountDataSource{}
}

type prismCountDataSource struct {
	client *client.Client
}

type prismCountDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Category   types.String `tfsdk:"category"`
	TotalCount types.Int64  `tfsdk:"total_count"`
}

func (d *prismCountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prism_count"
}

func (d *prismCountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve the total number of items within a specific Prism reporting category (e.g., total number of installed apps or devices).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"category": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Prism category to count (e.g., `apps`, `device_information`, `certificates`).",
			},
			"total_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The total count of items in the specified category.",
			},
		},
	}
}

func (d *prismCountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *prismCountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data prismCountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var response client.PrismCount
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/prism/count?category=%s", data.Category.ValueString()), nil, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read prism count, got error: %s", err))
		return
	}

	data.ID = types.StringValue("prism_count_" + data.Category.ValueString())
	data.TotalCount = types.Int64Value(int64(response.Count))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
