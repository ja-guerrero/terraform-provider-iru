package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ ephemeral.EphemeralResource = &deviceSecretsEphemeralResource{}

func NewDeviceSecretsEphemeralResource() ephemeral.EphemeralResource {
	return &deviceSecretsEphemeralResource{}
}

type deviceSecretsEphemeralResource struct {
	client *client.Client
}

type deviceSecretsEphemeralResourceModel struct {
	DeviceID             types.String `tfsdk:"device_id"`
	UserBasedALBC        types.String `tfsdk:"user_based_albc"`
	DeviceBasedALBC      types.String `tfsdk:"device_based_albc"`
	FileVaultRecoveryKey types.String `tfsdk:"filevault_recovery_key"`
	UnlockPin            types.String `tfsdk:"unlock_pin"`
	RecoveryLockPassword types.String `tfsdk:"recovery_lock_password"`
}

func (r *deviceSecretsEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_secrets"
}

func (r *deviceSecretsEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch sensitive secrets for a specific device, including Activation Lock bypass codes, FileVault recovery keys, and unlock PINs. This is an ephemeral resource; these highly sensitive values are NOT stored in the Terraform state.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"user_based_albc": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "User-based Activation Lock Bypass Code.",
			},
			"device_based_albc": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Device-based Activation Lock Bypass Code.",
			},
			"filevault_recovery_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "FileVault Personal Recovery Key.",
			},
			"unlock_pin": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Device Unlock PIN.",
			},
			"recovery_lock_password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Recovery Lock Password.",
			},
		},
	}
}

func (r *deviceSecretsEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *deviceSecretsEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data deviceSecretsEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()

	// ALBC
	var albc client.DeviceSecretsALBC
	err := r.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/secrets/bypasscode", deviceID), nil, &albc)
	if err == nil {
		data.UserBasedALBC = types.StringValue(albc.UserBasedALBC)
		data.DeviceBasedALBC = types.StringValue(albc.DeviceBasedALBC)
	}

	// FileVault
	var fv client.DeviceSecretsFileVault
	err = r.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/secrets/filevaultkey", deviceID), nil, &fv)
	if err == nil {
		data.FileVaultRecoveryKey = types.StringValue(fv.Key)
	}

	// Unlock Pin
	var pin client.DeviceSecretsUnlockPin
	err = r.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/secrets/unlockpin", deviceID), nil, &pin)
	if err == nil {
		data.UnlockPin = types.StringValue(pin.Pin)
	}

	// Recovery Lock
	var rl client.DeviceSecretsRecoveryLock
	err = r.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/secrets/recoverypassword", deviceID), nil, &rl)
	if err == nil {
		data.RecoveryLockPassword = types.StringValue(rl.RecoveryPassword)
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
