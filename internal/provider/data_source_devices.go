package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &devicesDataSource{}

func NewDevicesDataSource() datasource.DataSource {
	return &devicesDataSource{}
}

type devicesDataSource struct {
	client *client.Client
}

type devicesDataSourceModel struct {
	ID           types.String  `tfsdk:"id"`
	Limit        types.Int64   `tfsdk:"limit"`
	Offset       types.Int64   `tfsdk:"offset"`
	SerialNumber types.String  `tfsdk:"serial_number"`
	AssetTag     types.String  `tfsdk:"asset_tag"`
	DeviceName   types.String  `tfsdk:"device_name"`
	Platform     types.String  `tfsdk:"platform"`
	UserID       types.String  `tfsdk:"user_id"`
	BlueprintID  types.String  `tfsdk:"blueprint_id"`
	Devices      []deviceModel `tfsdk:"devices"`
}

type deviceModel struct {
	ID           types.String `tfsdk:"id"`
	DeviceName   types.String `tfsdk:"device_name"`
	SerialNumber types.String `tfsdk:"serial_number"`
	Model        types.String `tfsdk:"model"`
	OSVersion    types.String `tfsdk:"os_version"`
	Platform     types.String `tfsdk:"platform"`
	LastCheckIn  types.String `tfsdk:"last_check_in"`
}

func (d *devicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devices"
}

func (d *devicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all devices in the Iru instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of results to return.",
			},
			"offset": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of results to skip.",
			},
			"serial_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by serial number. Supports partial matches.",
			},
			"asset_tag": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by asset tag. Supports partial matches.",
			},
			"device_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by device name. Supports partial matches.",
			},
			"platform": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by platform. Options: `Mac`, `iPad`, `iPhone`, `AppleTV`, `Vision`.",
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by the UUID of the assigned user.",
			},
			"blueprint_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by the UUID of the assigned blueprint.",
			},
			"devices": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for the Device.",
						},
						"device_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the Device.",
						},
						"serial_number": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The serial number of the Device.",
						},
						"model": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The model of the Device.",
						},
						"os_version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The OS version of the Device.",
						},
						"platform": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The platform of the Device.",
						},
						"last_check_in": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The last check-in time of the Device.",
						},
					},
				},
			},
		},
	}
}

func (d *devicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *devicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data devicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allDevices []client.Device
	offset := 0
	if !data.Offset.IsNull() {
		offset = int(data.Offset.ValueInt64())
	}
	limit := 300
	if !data.Limit.IsNull() {
		limit = int(data.Limit.ValueInt64())
	}

	for {
		params := url.Values{}
		params.Add("limit", fmt.Sprintf("%d", limit))
		params.Add("offset", fmt.Sprintf("%d", offset))

		if !data.SerialNumber.IsNull() {
			params.Add("serial_number", data.SerialNumber.ValueString())
		}
		if !data.AssetTag.IsNull() {
			params.Add("asset_tag", data.AssetTag.ValueString())
		}
		if !data.DeviceName.IsNull() {
			params.Add("device_name", data.DeviceName.ValueString())
		}
		if !data.Platform.IsNull() {
			params.Add("platform", data.Platform.ValueString())
		}
		if !data.UserID.IsNull() {
			params.Add("user_id", data.UserID.ValueString())
		}
		if !data.BlueprintID.IsNull() {
			params.Add("blueprint_id", data.BlueprintID.ValueString())
		}

		path := "/api/v1/devices?" + params.Encode()
		var pageResponse []client.Device
		err := d.client.DoRequest(ctx, "GET", path, nil, &pageResponse)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read devices, got error: %s", err))
			return
		}

		allDevices = append(allDevices, pageResponse...)

		if !data.Limit.IsNull() && len(allDevices) >= limit {
			allDevices = allDevices[:limit]
			break
		}

		if len(pageResponse) < limit {
			break
		}
		offset += len(pageResponse)
	}

	data.ID = types.StringValue("devices")
	data.Devices = make([]deviceModel, 0, len(allDevices))
	for _, device := range allDevices {
		data.Devices = append(data.Devices, deviceModel{
			ID:           types.StringValue(device.ID),
			DeviceName:   types.StringValue(device.DeviceName),
			SerialNumber: types.StringValue(device.SerialNumber),
			Model:        types.StringValue(device.Model),
			OSVersion:    types.StringValue(device.OSVersion),
			Platform:     types.StringValue(device.Platform),
			LastCheckIn:  types.StringValue(device.LastCheckIn),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
