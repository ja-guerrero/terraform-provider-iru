package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ action.Action = &deviceSetPersonalHotspotAction{}

func NewDeviceSetPersonalHotspotAction() action.Action {
	return &deviceSetPersonalHotspotAction{}
}

type deviceSetPersonalHotspotAction struct {
	client *client.Client
}

type deviceSetPersonalHotspotActionModel struct {
	DeviceID types.String `tfsdk:"device_id"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

func (a *deviceSetPersonalHotspotAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_set_personal_hotspot"
}

func (a *deviceSetPersonalHotspotAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets personal hotspot settings for an Apple device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether personal hotspot should be enabled.",
			},
		},
	}
}

func (a *deviceSetPersonalHotspotAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceSetPersonalHotspotAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceSetPersonalHotspotActionModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	payload := map[string]bool{
		"Enabled": data.Enabled.ValueBool(),
	}

	err := a.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/action/togglepersonalhotspot", deviceID), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set personal hotspot, got error: %s", err))
		return
	}
}
