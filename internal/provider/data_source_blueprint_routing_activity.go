package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &blueprintRoutingActivityDataSource{}

func NewBlueprintRoutingActivityDataSource() datasource.DataSource {
	return &blueprintRoutingActivityDataSource{}
}

type blueprintRoutingActivityDataSource struct {
	client *client.Client
}

type blueprintRoutingActivityDataSourceModel struct {
	Limit      types.Int64                                   `tfsdk:"limit"`
	Activities []blueprintRoutingActivityDataSourceItemModel `tfsdk:"activities"`
}

type blueprintRoutingActivityDataSourceItemModel struct {
	ID           types.Int64  `tfsdk:"id"`
	ActivityTime types.String `tfsdk:"activity_time"`
	ActivityType types.String `tfsdk:"activity_type"`
	DeviceID     types.String `tfsdk:"device_id"`
}

func (d *blueprintRoutingActivityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_routing_activity"
}

func (d *blueprintRoutingActivityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides Blueprint Routing activity information.",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of results to return. Default is 300.",
			},
			"activities": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"activity_time": schema.StringAttribute{
							Computed: true,
						},
						"activity_type": schema.StringAttribute{
							Computed: true,
						},
						"device_id": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *blueprintRoutingActivityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *blueprintRoutingActivityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data blueprintRoutingActivityDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limit := 300
	if !data.Limit.IsNull() {
		limit = int(data.Limit.ValueInt64())
	}

	var response client.BlueprintRoutingActivityList
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/blueprint-routing/activity?limit=%d", limit), nil, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read blueprint routing activity, got error: %s", err))
		return
	}

	for _, item := range response.Results {
		activity := blueprintRoutingActivityDataSourceItemModel{
			ID:           types.Int64Value(int64(item.ID)),
			ActivityTime: types.StringValue(item.ActivityTime),
			ActivityType: types.StringValue(item.ActivityType),
			DeviceID:     types.StringValue(item.DeviceID),
		}
		data.Activities = append(data.Activities, activity)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
