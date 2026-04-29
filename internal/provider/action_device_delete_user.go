package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ action.Action = &deviceDeleteUserAction{}

func NewDeviceDeleteUserAction() action.Action {
	return &deviceDeleteUserAction{}
}

type deviceDeleteUserAction struct {
	client *client.Client
}

type deviceDeleteUserActionModel struct {
	DeviceID       types.String `tfsdk:"device_id"`
	DeleteAllUsers types.Bool   `tfsdk:"delete_all_users"`
	ForceDeletion  types.Bool   `tfsdk:"force_deletion"`
	UserName       types.String `tfsdk:"user_name"`
}

func (a *deviceDeleteUserAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_action_delete_user"
}

func (a *deviceDeleteUserAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deletes a user from a specific device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"delete_all_users": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "If true, deletes all users.",
			},
			"force_deletion": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "If true, forces deletion.",
			},
			"user_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The username to delete.",
			},
		},
	}
}

func (a *deviceDeleteUserAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a.client = req.ProviderData.(*client.Client)
}

func (a *deviceDeleteUserAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data deviceDeleteUserActionModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()
	payload := map[string]interface{}{}

	if !data.DeleteAllUsers.IsNull() {
		payload["DeleteAllUsers"] = data.DeleteAllUsers.ValueBool()
	} else {
		payload["DeleteAllUsers"] = false
	}

	if !data.ForceDeletion.IsNull() {
		payload["ForceDeletion"] = data.ForceDeletion.ValueBool()
	} else {
		payload["ForceDeletion"] = false
	}

	if !data.UserName.IsNull() {
		payload["UserName"] = data.UserName.ValueString()
	}

	err := a.client.DoRequest(ctx, "POST", fmt.Sprintf("/api/v1/devices/%s/action/deleteuser", deviceID), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}
