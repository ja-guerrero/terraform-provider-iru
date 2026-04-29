package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ action.Action = &deviceEraseAction{}

func NewDeviceEraseAction() action.Action {
	return &deviceEraseAction{}
}

type deviceEraseAction struct {
	client *client.Client
}

type deviceEraseActionModel struct {
	DeviceID               types.String `tfsdk:"device_id"`
	PIN                    types.String `tfsdk:"pin"`
	PreserveDataPlan       types.Bool   `tfsdk:"preserve_data_plan"`
	DisallowProximitySetup types.Bool   `tfsdk:"disallow_proximity_setup"`
	EraseMode              types.String `tfsdk:"erase_mode"`
	EraseFlags             types.String `tfsdk:"erase_flags"`
	ReturnToServiceEnabled types.Bool   `tfsdk:"return_to_service_enabled"`
	ReturnToServiceProfile types.String `tfsdk:"return_to_service_profile"`
}

func (a *deviceEraseAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_erase"
}

func (a *deviceEraseAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Erases a specific device. This is a **HIGHLY DESTRUCTIVE** imperative action. Behavior varies by platform: macOS uses the PIN for Find My; Windows and Android support specific wipe modes and flags; and supported Apple devices can utilize Return to Service (RTS) for automated WiFi profile association after the wipe.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"pin": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The six-character PIN for Find My (macOS only).",
			},
			"preserve_data_plan": schema.BoolAttribute{
				Optional: true,
			},
			"disallow_proximity_setup": schema.BoolAttribute{
				Optional: true,
			},
			"erase_mode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "For Windows devices: WIPE, WIPE_CLOUD, WIPE_PROTECTED.",
			},
			"erase_flags": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "For Android devices: WIPE_EXTERNAL_STORAGE, WIPE_ESIMS.",
			},
			"return_to_service_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to enable Return to Service.",
			},
			"return_to_service_profile": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The WiFi profile ID for Return to Service.",
			},
		},
	}
}

func (a *deviceEraseAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceEraseAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceEraseActionModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	payload := map[string]interface{}{}
	if !data.PIN.IsNull() {
		payload["PIN"] = data.PIN.ValueString()
	}
	if !data.PreserveDataPlan.IsNull() {
		payload["PreserveDataPlan"] = data.PreserveDataPlan.ValueBool()
	}
	if !data.DisallowProximitySetup.IsNull() {
		payload["DisallowProximitySetup"] = data.DisallowProximitySetup.ValueBool()
	}
	if !data.EraseMode.IsNull() {
		payload["erase_mode"] = data.EraseMode.ValueString()
	}
	if !data.EraseFlags.IsNull() {
		payload["erase_flags"] = data.EraseFlags.ValueString()
	}

	if !data.ReturnToServiceEnabled.IsNull() {
		rts := map[string]interface{}{
			"Enabled": data.ReturnToServiceEnabled.ValueBool(),
		}
		if !data.ReturnToServiceProfile.IsNull() {
			rts["ProfileId"] = data.ReturnToServiceProfile.ValueString()
		}
		payload["ReturnToService"] = rts
	}

	err := a.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/action/erase", deviceID), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to invoke erase, got error: %s", err))
		return
	}
}
