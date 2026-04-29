package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceDetailsDataSource{}

func NewDeviceDetailsDataSource() datasource.DataSource {
	return &deviceDetailsDataSource{}
}

type deviceDetailsDataSource struct {
	client *client.Client
}

type deviceDetailsDataSourceModel struct {
	DeviceID     types.String `tfsdk:"device_id"`
	DeviceName   types.String `tfsdk:"device_name"`
	Model        types.String `tfsdk:"model"`
	Platform     types.String `tfsdk:"platform"`
	OSVersion    types.String `tfsdk:"os_version"`
	SerialNumber types.String `tfsdk:"serial_number"`
	AssetTag     types.String `tfsdk:"asset_tag"`
	BlueprintID  types.String `tfsdk:"blueprint_id"`
	MDMEnabled   types.String `tfsdk:"mdm_enabled"`
	Supervised   types.String `tfsdk:"supervised"`
	LastCheckIn  types.String `tfsdk:"last_check_in"`
}

func (d *deviceDetailsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_details"
}

func (d *deviceDetailsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get full details for a specific device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"device_name": schema.StringAttribute{
				Computed: true,
			},
			"model": schema.StringAttribute{
				Computed: true,
			},
			"platform": schema.StringAttribute{
				Computed: true,
			},
			"os_version": schema.StringAttribute{
				Computed: true,
			},
			"serial_number": schema.StringAttribute{
				Computed: true,
			},
			"asset_tag": schema.StringAttribute{
				Computed: true,
			},
			"blueprint_id": schema.StringAttribute{
				Computed: true,
			},
			"mdm_enabled": schema.StringAttribute{
				Computed: true,
			},
			"supervised": schema.StringAttribute{
				Computed: true,
			},
			"last_check_in": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *deviceDetailsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceDetailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceDetailsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var details client.DeviceDetails
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/details", data.DeviceID.ValueString()), nil, &details)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device details, got error: %s", err))
		return
	}

	data.DeviceName = types.StringValue(details.General.DeviceName)
	data.Model = types.StringValue(details.General.Model)
	data.Platform = types.StringValue(details.General.Platform)
	data.OSVersion = types.StringValue(details.General.OSVersion)
	data.SerialNumber = types.StringValue(details.General.SerialNumber)
	data.AssetTag = types.StringValue(details.General.AssetTag)
	data.BlueprintID = types.StringValue(details.General.BlueprintID)
	data.MDMEnabled = types.StringValue(details.MDM.Enabled)
	data.Supervised = types.StringValue(details.MDM.Supervised)
	data.LastCheckIn = types.StringValue(details.MDM.LastCheckIn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
