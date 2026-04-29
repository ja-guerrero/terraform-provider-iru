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

var _ list.ListResource = &adeDeviceListResource{}
var _ list.ListResourceWithConfigure = &adeDeviceListResource{}

func NewADEDeviceListResource() list.ListResource {
	return &adeDeviceListResource{}
}

type adeDeviceListResource struct {
	client *client.Client
}

func (r *adeDeviceListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ade_device"
}

func (r *adeDeviceListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "Lists Iru ADE Device resources.",
	}
}

func (r *adeDeviceListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *adeDeviceListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var allDevices []client.ADEDevice
	page := 1

	for {
		path := fmt.Sprintf("/api/v1/integrations/apple/ade/devices?page=%d", page)

		type adeDevicesResponse struct {
			Results []client.ADEDevice `json:"results"`
			Next    string             `json:"next"`
		}
		var listResp adeDevicesResponse
		err := r.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
				diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list ADE devices, got error: %v", err)),
			})
			return
		}

		allDevices = append(allDevices, listResp.Results...)
		if listResp.Next == "" || len(listResp.Results) == 0 {
			break
		}
		page++
	}

	results := make([]list.ListResult, 0, len(allDevices))
	for _, device := range allDevices {
		result := req.NewListResult(ctx)

		identity := adeDeviceResourceIdentityModel{
			ID: types.StringValue(device.ID),
		}
		result.Diagnostics.Append(result.Identity.Set(ctx, &identity)...)

		if req.IncludeResource {
			resourceModel := adeDeviceResourceModel{
				ID:                  types.StringValue(device.ID),
				SerialNumber:        types.StringValue(device.SerialNumber),
				Model:               types.StringValue(device.Model),
				Description:         types.StringValue(device.Description),
				AssetTag:            types.StringValue(device.AssetTag),
				Color:               types.StringValue(device.Color),
				BlueprintID:         types.StringValue(device.BlueprintID),
				UserID:              types.StringValue(device.UserID),
				DEPAccount:          types.StringValue(device.DEPAccount),
				DeviceFamily:        types.StringValue(device.DeviceFamily),
				OS:                  types.StringValue(device.OS),
				ProfileStatus:       types.StringValue(device.ProfileStatus),
				IsEnrolled:          types.BoolValue(device.IsEnrolled),
				UseBlueprintRouting: types.BoolValue(device.UseBlueprintRouting),
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, &resourceModel)...)
		}

		result.DisplayName = fmt.Sprintf("%s (%s)", device.Model, device.SerialNumber)
		results = append(results, result)
	}

	resp.Results = slices.Values(results)
}
