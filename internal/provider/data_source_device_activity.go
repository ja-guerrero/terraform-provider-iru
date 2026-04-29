package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceActivityDataSource{}

func NewDeviceActivityDataSource() datasource.DataSource {
	return &deviceActivityDataSource{}
}

type deviceActivityDataSource struct {
	client *client.Client
}

type deviceActivityDataSourceModel struct {
	DeviceID types.String          `tfsdk:"device_id"`
	Limit    types.Int64           `tfsdk:"limit"`
	Activity []deviceActivityModel `tfsdk:"activity"`
}

type deviceActivityModel struct {
	ID               types.Int64  `tfsdk:"id"`
	CreatedAt        types.String `tfsdk:"created_at"`
	ActionType       types.String `tfsdk:"action_type"`
	BlueprintRouting types.Bool   `tfsdk:"blueprint_routing"`
}

func (d *deviceActivityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_activity"
}

func (d *deviceActivityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List activity for a device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Limit results (max 300).",
			},
			"activity": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"created_at": schema.StringAttribute{
							Computed: true,
						},
						"action_type": schema.StringAttribute{
							Computed: true,
						},
						"blueprint_routing": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *deviceActivityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceActivityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceActivityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limit := int64(300)
	if !data.Limit.IsNull() {
		limit = data.Limit.ValueInt64()
	}

	var response client.DeviceActivityList
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/activity?limit=%d", data.DeviceID.ValueString(), limit), nil, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device activity, got error: %s", err))
		return
	}

	for _, act := range response.Activity.Results {
		data.Activity = append(data.Activity, deviceActivityModel{
			ID:               types.Int64Value(int64(act.ID)),
			CreatedAt:        types.StringValue(act.CreatedAt),
			ActionType:       types.StringValue(act.ActionType),
			BlueprintRouting: types.BoolValue(act.BlueprintRouting),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
