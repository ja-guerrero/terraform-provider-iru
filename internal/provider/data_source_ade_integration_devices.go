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

var _ datasource.DataSource = &adeIntegrationDevicesDataSource{}

func NewADEIntegrationDevicesDataSource() datasource.DataSource {
	return &adeIntegrationDevicesDataSource{}
}

type adeIntegrationDevicesDataSource struct {
	client *client.Client
}

type adeIntegrationDevicesDataSourceModel struct {
	ID         types.String     `tfsdk:"id"`
	ADETokenID types.String     `tfsdk:"ade_token_id"`
	Devices    []adeDeviceModel `tfsdk:"devices"`
}

func (d *adeIntegrationDevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ade_integration_devices"
}

func (d *adeIntegrationDevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List devices associated with a specific ADE token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"ade_token_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the ADE token.",
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
					},
				},
			},
		},
	}
}

func (d *adeIntegrationDevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *adeIntegrationDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data adeIntegrationDevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allDevices []client.ADEDevice
	page := 1

	for {
		params := url.Values{}
		params.Add("page", fmt.Sprintf("%d", page))

		path := fmt.Sprintf("/api/v1/integrations/apple/ade/%s/devices?%s", data.ADETokenID.ValueString(), params.Encode())

		type adeDevicesResponse struct {
			Results []client.ADEDevice `json:"results"`
			Next    string             `json:"next"`
		}
		var listResp adeDevicesResponse
		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ADE integration devices, got error: %s", err))
			return
		}

		allDevices = append(allDevices, listResp.Results...)
		if listResp.Next == "" || len(listResp.Results) == 0 {
			break
		}
		page++
	}

	data.ID = types.StringValue("ade_integration_devices_" + data.ADETokenID.ValueString())
	data.Devices = make([]adeDeviceModel, 0, len(allDevices))
	for _, device := range allDevices {
		data.Devices = append(data.Devices, adeDeviceModel{
			ID:            types.StringValue(device.ID),
			SerialNumber:  types.StringValue(device.SerialNumber),
			Model:         types.StringValue(device.Model),
			Description:   types.StringValue(device.Description),
			AssetTag:      types.StringValue(device.AssetTag),
			Color:         types.StringValue(device.Color),
			BlueprintID:   types.StringValue(device.BlueprintID),
			UserID:        types.StringValue(device.UserID),
			DEPAccount:    types.StringValue(device.DEPAccount),
			DeviceFamily:  types.StringValue(device.DeviceFamily),
			OS:            types.StringValue(device.OS),
			ProfileStatus: types.StringValue(device.ProfileStatus),
			IsEnrolled:    types.BoolValue(device.IsEnrolled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
