package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceStatusDataSource{}

func NewDeviceStatusDataSource() datasource.DataSource {
	return &deviceStatusDataSource{}
}

type deviceStatusDataSource struct {
	client *client.Client
}

type deviceStatusDataSourceModel struct {
	DeviceID     types.String              `tfsdk:"device_id"`
	LibraryItems []deviceLibraryItemsModel `tfsdk:"library_items"`
	Parameters   []deviceParametersModel   `tfsdk:"parameters"`
}

func (d *deviceStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_status"
}

func (d *deviceStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get the full status (library items and parameters) for a specific device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"library_items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":     schema.StringAttribute{Computed: true},
						"name":   schema.StringAttribute{Computed: true},
						"status": schema.StringAttribute{Computed: true},
					},
				},
			},
			"parameters": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":     schema.StringAttribute{Computed: true},
						"status": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *deviceStatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceStatusDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var status struct {
		LibraryItems []struct {
			ID     string `json:"library_item_id"`
			Name   string `json:"library_item_name"`
			Status string `json:"status"`
		} `json:"library_items"`
		Parameters []struct {
			ID     string `json:"parameter_id"`
			Status string `json:"status"`
		} `json:"parameters"`
	}
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/status", data.DeviceID.ValueString()), nil, &status)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device status, got error: %s", err))
		return
	}

	for _, item := range status.LibraryItems {
		data.LibraryItems = append(data.LibraryItems, deviceLibraryItemsModel{
			ID:     types.StringValue(item.ID),
			Name:   types.StringValue(item.Name),
			Status: types.StringValue(item.Status),
		})
	}

	for _, p := range status.Parameters {
		data.Parameters = append(data.Parameters, deviceParametersModel{
			ID:     types.StringValue(p.ID),
			Status: types.StringValue(p.Status),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
