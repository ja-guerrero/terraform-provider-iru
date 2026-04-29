package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ action.Action = &deviceEnableLostModeAction{}

func NewDeviceEnableLostModeAction() action.Action {
	return &deviceEnableLostModeAction{}
}

type deviceEnableLostModeAction struct {
	client *client.Client
}

type deviceEnableLostModeActionModel struct {
	DeviceID    types.String `tfsdk:"device_id"`
	Message     types.String `tfsdk:"message"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	Footnote    types.String `tfsdk:"footnote"`
}

func (a *deviceEnableLostModeAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_enable_lost_mode"
}

func (a *deviceEnableLostModeAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Enables Lost Mode on a specific device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"message": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Message to display on the lost device.",
			},
			"phone_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Phone number to display on the lost device.",
			},
			"footnote": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Footnote to display on the lost device.",
			},
		},
	}
}

func (a *deviceEnableLostModeAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceEnableLostModeAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceEnableLostModeActionModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	payload := map[string]string{}
	if !data.Message.IsNull() {
		payload["Message"] = data.Message.ValueString()
	}
	if !data.PhoneNumber.IsNull() {
		payload["PhoneNumber"] = data.PhoneNumber.ValueString()
	}
	if !data.Footnote.IsNull() {
		payload["Footnote"] = data.Footnote.ValueString()
	}

	err := a.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/action/enablelostmode", deviceID), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable lost mode, got error: %s", err))
		return
	}
}
