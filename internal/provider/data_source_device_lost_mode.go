package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceLostModeDataSource{}

func NewDeviceLostModeDataSource() datasource.DataSource {
	return &deviceLostModeDataSource{}
}

type deviceLostModeDataSource struct {
	client *client.Client
}

type deviceLostModeDataSourceModel struct {
	DeviceID    types.String `tfsdk:"device_id"`
	Status      types.String `tfsdk:"status"`
	Message     types.String `tfsdk:"message"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	Footnote    types.String `tfsdk:"footnote"`
}

func (d *deviceLostModeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_lost_mode"
}

func (d *deviceLostModeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Lost Mode details for a specific device. To manage Lost Mode, use `iru_device_action_enable_lost_mode`, `iru_device_action_disable_lost_mode` (standard unlock), or `iru_device_action_cancel_lost_mode` (error recovery).",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"message": schema.StringAttribute{
				Computed: true,
			},
			"phone_number": schema.StringAttribute{
				Computed: true,
			},
			"footnote": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *deviceLostModeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceLostModeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceLostModeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var lostMode struct {
		Status      string `json:"status"`
		Message     string `json:"message"`
		PhoneNumber string `json:"phone_number"`
		Footnote    string `json:"footnote"`
	}
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/details/lostmode", data.DeviceID.ValueString()), nil, &lostMode)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device lost mode details, got error: %s", err))
		return
	}

	data.Status = types.StringValue(lostMode.Status)
	data.Message = types.StringValue(lostMode.Message)
	data.PhoneNumber = types.StringValue(lostMode.PhoneNumber)
	data.Footnote = types.StringValue(lostMode.Footnote)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
