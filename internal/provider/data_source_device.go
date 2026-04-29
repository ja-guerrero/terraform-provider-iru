package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceDataSource{}

func NewDeviceDataSource() datasource.DataSource {
	return &deviceDataSource{}
}

type deviceDataSource struct {
	client *client.Client
}

type deviceDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	DeviceName   types.String `tfsdk:"device_name"`
	SerialNumber types.String `tfsdk:"serial_number"`
	Platform     types.String `tfsdk:"platform"`
	OSVersion    types.String `tfsdk:"os_version"`
	BlueprintID  types.String `tfsdk:"blueprint_id"`
	UserID       types.String `tfsdk:"user_id"`
	AssetTag     types.String `tfsdk:"asset_tag"`
}

func (d *deviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (d *deviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get detailed information for a specific device, including its name, serial number, platform, and assignment status.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"device_name": schema.StringAttribute{
				Computed: true,
			},
			"serial_number": schema.StringAttribute{
				Computed: true,
			},
			"platform": schema.StringAttribute{
				Computed: true,
			},
			"os_version": schema.StringAttribute{
				Computed: true,
			},
			"blueprint_id": schema.StringAttribute{
				Computed: true,
			},
			"user_id": schema.StringAttribute{
				Computed: true,
			},
			"asset_tag": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *deviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var device client.Device
	err := d.client.DoRequest(ctx, "GET", "/api/v1/devices/"+data.ID.ValueString(), nil, &device)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device, got error: %s", err))
		return
	}

	data.DeviceName = types.StringValue(device.DeviceName)
	data.SerialNumber = types.StringValue(device.SerialNumber)
	data.Platform = types.StringValue(device.Platform)
	data.OSVersion = types.StringValue(device.OSVersion)
	data.BlueprintID = types.StringValue(device.BlueprintID)
	data.UserID = types.StringValue(device.UserID)
	data.AssetTag = types.StringValue(device.AssetTag)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
