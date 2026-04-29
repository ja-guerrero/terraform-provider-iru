package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ action.Action = &deviceCancelLostModeAction{}

func NewDeviceCancelLostModeAction() action.Action {
	return &deviceCancelLostModeAction{}
}

type deviceCancelLostModeAction struct {
	client *client.Client
}

type deviceCancelLostModeActionModel struct {
	DeviceID types.String `tfsdk:"device_id"`
}

func (a *deviceCancelLostModeAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_cancel_lost_mode"
}

func (a *deviceCancelLostModeAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sends a cancelation request if Lost Mode is in an error state. This is an **error-recovery/administrative action** used when the Lost Mode state is 'stuck'. Instead of just telling the device to stop, it attempts to clear the record/request cycle that might be preventing the device from updating. Use this only if the standard `disable` command has failed or if the Iru console shows the device is in an error state regarding its Lost Mode status.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
		},
	}
}

func (a *deviceCancelLostModeAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceCancelLostModeAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceCancelLostModeActionModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	err := a.client.DoRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/devices/%s/details/lostmode", deviceID), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to cancel lost mode, got error: %s", err))
		return
	}
}
