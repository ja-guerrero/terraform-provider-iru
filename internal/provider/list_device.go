package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ list.ListResource = &deviceListResource{}
var _ list.ListResourceWithConfigure = &deviceListResource{}

func NewDeviceListResource() list.ListResource {
	return &deviceListResource{}
}

type deviceListResource struct {
	client *client.Client
}

func (r *deviceListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *deviceListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{MarkdownDescription: "Lists Iru Device resources."}
}

func (r *deviceListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *deviceListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var devices []client.Device
	err := r.client.DoRequest(ctx, "GET", "/api/v1/devices", nil, &devices)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list devices, got error: %v", err)),
		})
		return
	}

	results := make([]list.ListResult, 0, len(devices))
	for _, device := range devices {
		result := req.NewListResult(ctx)

		// API returns device_id for list
		id := device.ID
		identity := deviceResourceIdentityModel{
			ID: types.StringValue(id),
		}
		result.Diagnostics.Append(result.Identity.Set(ctx, &identity)...)

		if req.IncludeResource {
			resourceModel := deviceResourceModel{
				ID:           types.StringValue(id),
				DeviceName:   types.StringValue(device.DeviceName),
				AssetTag:     types.StringValue(device.AssetTag),
				BlueprintID:  types.StringValue(device.BlueprintID),
				UserID:       types.StringValue(device.UserID),
				SerialNumber: types.StringValue(device.SerialNumber),
				Model:        types.StringValue(device.Model),
				OSVersion:    types.StringValue(device.OSVersion),
				Platform:     types.StringValue(device.Platform),
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, &resourceModel)...)
		}

		result.DisplayName = fmt.Sprintf("%s (%s)", device.DeviceName, device.SerialNumber)
		results = append(results, result)
	}

	resp.Results = slices.Values(results)
}
