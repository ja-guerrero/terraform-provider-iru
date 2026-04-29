package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ action.Action = &deviceSetNameAction{}

func NewDeviceSetNameAction() action.Action {
	return &deviceSetNameAction{}
}

type deviceSetNameAction struct {
	client *client.Client
}

type deviceSetNameActionModel struct {
	DeviceID   types.String `tfsdk:"device_id"`
	DeviceName types.String `tfsdk:"device_name"`
}

func (a *deviceSetNameAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_set_name"
}

func (a *deviceSetNameAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets the display name for a specific device. This is an imperative action that updates the name in the Iru console and, where supported, on the device itself.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"device_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The new name for the device.",
			},
		},
	}
}

func (a *deviceSetNameAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceSetNameAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceSetNameActionModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	payload := map[string]string{
		"DeviceName": data.DeviceName.ValueString(),
	}

	err := a.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/action/setname", deviceID), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set device name, got error: %s", err))
		return
	}
}
