package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

// deviceSimpleActionConfig holds the static configuration that distinguishes
// one zero-payload device action from another.
type deviceSimpleActionConfig struct {
	// actionName is the suffix appended to the Terraform type name, e.g. "blank_push".
	actionName string
	// actionPath is the API path segment, e.g. "blank-push".
	actionPath string
	// description is the MarkdownDescription shown in the schema.
	description string
}

// deviceSimpleAction implements action.Action for any device action that simply
// POSTs to /api/v1/devices/{device_id}/action/{actionPath} with no body.
type deviceSimpleAction struct {
	config deviceSimpleActionConfig
	client *client.Client
}

type deviceSimpleActionModel struct {
	DeviceID types.String `tfsdk:"device_id"`
}

var _ action.Action = &deviceSimpleAction{}

// newDeviceSimpleAction returns a factory function that creates a new
// deviceSimpleAction for the given config.
func newDeviceSimpleAction(cfg deviceSimpleActionConfig) func() action.Action {
	return func() action.Action {
		return &deviceSimpleAction{config: cfg}
	}
}

func (a *deviceSimpleAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_" + a.config.actionName
}

func (a *deviceSimpleAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: a.config.description,
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
		},
	}
}

func (a *deviceSimpleAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceSimpleAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceSimpleActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	err := a.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/action/%s", deviceID, a.config.actionPath), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to invoke %s, got error: %s", a.config.actionName, err))
		return
	}
}

// ---------------------------------------------------------------------------
// Constructor functions — one per simple action, preserving the names that
// provider.go already references.
// ---------------------------------------------------------------------------

func NewDeviceBlankPushAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "blank_push",
		actionPath:  "blank-push",
		description: "Sends a blank push to a specific device. This is an imperative action used to wake up a device and prompt it to check in with the MDM server.",
	})()
}

func NewDeviceClearPasscodeAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "clear_passcode",
		actionPath:  "clear-passcode",
		description: "Clears the passcode for a specific device. This is an imperative action.",
	})()
}

func NewDeviceDailyCheckinAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "daily_checkin",
		actionPath:  "dailycheckin",
		description: "Initiates a daily check-in for a device.",
	})()
}

func NewDeviceDisableLostModeAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "disable_lost_mode",
		actionPath:  "disablelostmode",
		description: "Disables Lost Mode on a specific device. This is the **standard MDM command** used to unlock a healthy device that is currently in Lost Mode. Use this when a user has recovered their device and you want to return it to a normal state. If the command is already pending, the API will indicate it is already in progress.",
	})()
}

func NewDeviceEnableRemoteDesktopAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "enable_remote_desktop",
		actionPath:  "enable-remote-desktop",
		description: "Enables Remote Desktop on a specific device. This is an imperative action.",
	})()
}

func NewDeviceForceCheckInAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "force_check_in",
		actionPath:  "force-check-in",
		description: "Forces a check-in for a specific device. This is an imperative action.",
	})()
}

func NewDeviceLockAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "lock",
		actionPath:  "lock",
		description: "Locks a specific device. This is an imperative action. For macOS, a PIN should be provided. For iOS/iPadOS, the device is locked to the lock screen.",
	})()
}

func NewDevicePlayLostModeSoundAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "play_lost_mode_sound",
		actionPath:  "playlostmodesound",
		description: "Plays the lost mode sound on a specific device.",
	})()
}

func NewDeviceRefreshCellularPlansAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "refresh_cellular_plans",
		actionPath:  "refreshcellularplans",
		description: "Refreshes cellular plans on a specific device.",
	})()
}

func NewDeviceReinstallAgentAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "reinstall_agent",
		actionPath:  "reinstallagent",
		description: "Reinstalls the Iru Agent on macOS devices.",
	})()
}

func NewDeviceRenewMDMProfileAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "renew_mdm_profile",
		actionPath:  "renewmdmprofile",
		description: "Renews the MDM profile on a specific device.",
	})()
}

func NewDeviceRestartAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "restart",
		actionPath:  "restart",
		description: "Restarts a specific device. This is an imperative action.",
	})()
}

func NewDeviceShutdownAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "shutdown",
		actionPath:  "shutdown",
		description: "Shuts down a specific device. This is an imperative action.",
	})()
}

func NewDeviceUpdateInventoryAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "update_inventory",
		actionPath:  "updateinventory",
		description: "Updates inventory for a specific device.",
	})()
}

func NewDeviceUpdateLocationAction() action.Action {
	return newDeviceSimpleAction(deviceSimpleActionConfig{
		actionName:  "update_location",
		actionPath:  "updatelocation",
		description: "Updates the location of a specific device.",
	})()
}
