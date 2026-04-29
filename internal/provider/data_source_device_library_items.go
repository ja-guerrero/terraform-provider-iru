package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceLibraryItemsDataSource{}

func NewDeviceLibraryItemsDataSource() datasource.DataSource {
	return &deviceLibraryItemsDataSource{}
}

type deviceLibraryItemsDataSource struct {
	client *client.Client
}

type deviceLibraryItemsDataSourceModel struct {
	DeviceID     types.String              `tfsdk:"device_id"`
	LibraryItems []deviceLibraryItemsModel `tfsdk:"library_items"`
}

type deviceLibraryItemsModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
}

func (d *deviceLibraryItemsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_library_items"
}

func (d *deviceLibraryItemsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List library items and their status for a specific device.",
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
		},
	}
}

func (d *deviceLibraryItemsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceLibraryItemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceLibraryItemsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var items []struct {
		ID     string `json:"library_item_id"`
		Name   string `json:"library_item_name"`
		Status string `json:"status"`
	}
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/library-items", data.DeviceID.ValueString()), nil, &items)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device library items, got error: %s", err))
		return
	}

	for _, item := range items {
		data.LibraryItems = append(data.LibraryItems, deviceLibraryItemsModel{
			ID:     types.StringValue(item.ID),
			Name:   types.StringValue(item.Name),
			Status: types.StringValue(item.Status),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
