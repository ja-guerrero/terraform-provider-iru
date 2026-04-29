package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ action.Action = &deviceSetDataRoamingAction{}

func NewDeviceSetDataRoamingAction() action.Action {
	return &deviceSetDataRoamingAction{}
}

type deviceSetDataRoamingAction struct {
	client *client.Client
}

type deviceSetDataRoamingActionModel struct {
	DeviceID types.String `tfsdk:"device_id"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

func (a *deviceSetDataRoamingAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_set_data_roaming"
}

func (a *deviceSetDataRoamingAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets data roaming settings for an Apple device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether data roaming should be enabled.",
			},
		},
	}
}

func (a *deviceSetDataRoamingAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceSetDataRoamingAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceSetDataRoamingActionModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	payload := map[string]bool{
		"Enabled": data.Enabled.ValueBool(),
	}

	err := a.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/action/toggledataroaming", deviceID), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set data roaming, got error: %s", err))
		return
	}
}
