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

var _ datasource.DataSource = &adeDevicesDataSource{}

func NewADEDevicesDataSource() datasource.DataSource {
	return &adeDevicesDataSource{}
}

type adeDevicesDataSource struct {
	client *client.Client
}

type adeDevicesDataSourceModel struct {
	ID            types.String     `tfsdk:"id"`
	BlueprintID   types.String     `tfsdk:"blueprint_id"`
	UserID        types.String     `tfsdk:"user_id"`
	DEPAccount    types.String     `tfsdk:"dep_account"`
	DeviceFamily  types.String     `tfsdk:"device_family"`
	Model         types.String     `tfsdk:"model"`
	OS            types.String     `tfsdk:"os"`
	ProfileStatus types.String     `tfsdk:"profile_status"`
	SerialNumber  types.String     `tfsdk:"serial_number"`
	Devices       []adeDeviceModel `tfsdk:"devices"`
}

type adeDeviceModel struct {
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

func (d *adeDevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ade_devices"
}

func (d *adeDevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all ADE devices in the Iru instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"blueprint_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by blueprint ID.",
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by user ID.",
			},
			"dep_account": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by DEP account.",
			},
			"device_family": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by device family.",
			},
			"model": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by model.",
			},
			"os": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by OS.",
			},
			"profile_status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by profile status.",
			},
			"serial_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by serial number.",
			},
			"devices": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *adeDevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *adeDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data adeDevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allDevices []client.ADEDevice
	page := 1

	for {
		params := url.Values{}
		params.Add("page", fmt.Sprintf("%d", page))

		if !data.BlueprintID.IsNull() {
			params.Add("blueprint_id", data.BlueprintID.ValueString())
		}
		if !data.UserID.IsNull() {
			params.Add("user_id", data.UserID.ValueString())
		}
		if !data.DEPAccount.IsNull() {
			params.Add("dep_account", data.DEPAccount.ValueString())
		}
		if !data.DeviceFamily.IsNull() {
			params.Add("device_family", data.DeviceFamily.ValueString())
		}
		if !data.Model.IsNull() {
			params.Add("model", data.Model.ValueString())
		}
		if !data.OS.IsNull() {
			params.Add("os", data.OS.ValueString())
		}
		if !data.ProfileStatus.IsNull() {
			params.Add("profile_status", data.ProfileStatus.ValueString())
		}
		if !data.SerialNumber.IsNull() {
			params.Add("serial_number", data.SerialNumber.ValueString())
		}

		path := "/api/v1/integrations/apple/ade/devices?" + params.Encode()

		type adeDevicesResponse struct {
			Results []client.ADEDevice `json:"results"`
			Next    string             `json:"next"`
		}
		var listResp adeDevicesResponse
		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ADE devices, got error: %s", err))
			return
		}

		allDevices = append(allDevices, listResp.Results...)
		if listResp.Next == "" || len(listResp.Results) == 0 {
			break
		}
		page++
	}

	data.ID = types.StringValue("ade_devices")
	data.Devices = make([]adeDeviceModel, 0, len(allDevices))
	for _, device := range allDevices {
		data.Devices = append(data.Devices, adeDeviceModel{
			ID:                  types.StringValue(device.ID),
			SerialNumber:        types.StringValue(device.SerialNumber),
			Model:               types.StringValue(device.Model),
			Description:         types.StringValue(device.Description),
			AssetTag:            types.StringValue(device.AssetTag),
			Color:               types.StringValue(device.Color),
			BlueprintID:         types.StringValue(device.BlueprintID),
			UserID:              types.StringValue(device.UserID),
			DEPAccount:          types.StringValue(device.DEPAccount),
			DeviceFamily:        types.StringValue(device.DeviceFamily),
			OS:                  types.StringValue(device.OS),
			ProfileStatus:       types.StringValue(device.ProfileStatus),
			IsEnrolled:          types.BoolValue(device.IsEnrolled),
			UseBlueprintRouting: types.BoolValue(device.UseBlueprintRouting),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
