package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &adeDeviceDataSource{}

func NewADEDeviceDataSource() datasource.DataSource {
	return &adeDeviceDataSource{}
}

type adeDeviceDataSource struct {
	client *client.Client
}

type adeDeviceDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	SerialNumber        types.String `tfsdk:"serial_number"`
	Model               types.String `tfsdk:"model"`
	Description         types.String `tfsdk:"description"`
	AssetTag            types.String `tfsdk:"asset_tag"`
	Color               types.String `tfsdk:"color"`
	BlueprintID         types.String `tfsdk:"blueprint_id"`
	UserID              types.String `tfsdk:"user_id"`
	DEPAccount          types.String `tfsdk:"dep_account"`
	DeviceFamily        types.String `tfsdk:"device_family"`
	OS                  types.String `tfsdk:"os"`
	ProfileStatus       types.String `tfsdk:"profile_status"`
	IsEnrolled          types.Bool   `tfsdk:"is_enrolled"`
	UseBlueprintRouting types.Bool   `tfsdk:"use_blueprint_routing"`
}

func (d *adeDeviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ade_device"
}

func (d *adeDeviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get details for a specific ADE device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the ADE Device.",
			},
			"serial_number": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The serial number of the Device.",
			},
			"model": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The model of the Device.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The description of the Device.",
			},
			"asset_tag": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The asset tag of the Device.",
			},
			"color": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The color of the Device.",
			},
			"blueprint_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the blueprint assigned to the Device.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user assigned to the Device.",
			},
			"dep_account": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The DEP account of the Device.",
			},
			"device_family": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The device family of the Device.",
			},
			"os": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The OS of the Device.",
			},
			"profile_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The profile status of the Device.",
			},
			"is_enrolled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the device is enrolled.",
			},
			"use_blueprint_routing": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the device uses Blueprint Routing.",
			},
		},
	}
}

func (d *adeDeviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *adeDeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data adeDeviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var deviceResponse client.ADEDevice
	err := d.client.DoRequest(ctx, "GET", "/api/v1/integrations/apple/ade/devices/"+data.ID.ValueString(), nil, &deviceResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ADE device, got error: %s", err))
		return
	}

	data.SerialNumber = types.StringValue(deviceResponse.SerialNumber)
	data.Model = types.StringValue(deviceResponse.Model)
	data.Description = types.StringValue(deviceResponse.Description)
	data.AssetTag = types.StringValue(deviceResponse.AssetTag)
	data.Color = types.StringValue(deviceResponse.Color)
	data.BlueprintID = types.StringValue(deviceResponse.BlueprintID)
	data.UserID = types.StringValue(deviceResponse.UserID)
	data.DEPAccount = types.StringValue(deviceResponse.DEPAccount)
	data.DeviceFamily = types.StringValue(deviceResponse.DeviceFamily)
	data.OS = types.StringValue(deviceResponse.OS)
	data.ProfileStatus = types.StringValue(deviceResponse.ProfileStatus)
	data.IsEnrolled = types.BoolValue(deviceResponse.IsEnrolled)
	data.UseBlueprintRouting = types.BoolValue(deviceResponse.UseBlueprintRouting)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
