package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ action.Action = &deviceUnlockAccountAction{}

func NewDeviceUnlockAccountAction() action.Action {
	return &deviceUnlockAccountAction{}
}

type deviceUnlockAccountAction struct {
	client *client.Client
}

type deviceUnlockAccountActionModel struct {
	DeviceID types.String `tfsdk:"device_id"`
	UserName types.String `tfsdk:"username"`
}

func (a *deviceUnlockAccountAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_unlock_account"
}

func (a *deviceUnlockAccountAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Unlocks a specific user account on a device. Available for Mac.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The local username to unlock.",
			},
		},
	}
}

func (a *deviceUnlockAccountAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceUnlockAccountAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceUnlockAccountActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]string{
		"UserName": data.UserName.ValueString(),
	}
	err := a.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/action/unlockaccount", data.DeviceID.ValueString()), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to invoke unlock account, got error: %s", err))
		return
	}
}
