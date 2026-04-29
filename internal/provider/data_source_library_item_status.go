package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &libraryItemStatusDataSource{}

func NewLibraryItemStatusDataSource() datasource.DataSource {
	return &libraryItemStatusDataSource{}
}

type libraryItemStatusDataSource struct {
	client *client.Client
}

type libraryItemStatusDataSourceModel struct {
	LibraryItemID types.String             `tfsdk:"library_item_id"`
	Statuses      []libraryItemStatusModel `tfsdk:"statuses"`
}

type libraryItemStatusModel struct {
	DeviceID   types.String `tfsdk:"device_id"`
	DeviceName types.String `tfsdk:"device_name"`
	Status     types.String `tfsdk:"status"`
}

func (d *libraryItemStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_library_item_status"
}

func (d *libraryItemStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get status for a specific library item across devices.",
		Attributes: map[string]schema.Attribute{
			"library_item_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Library Item.",
			},
			"statuses": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"device_id":   schema.StringAttribute{Computed: true},
						"device_name": schema.StringAttribute{Computed: true},
						"status":      schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *libraryItemStatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *libraryItemStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data libraryItemStatusDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var statuses []client.LibraryItemStatus
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/library/library-items/%s/status", data.LibraryItemID.ValueString()), nil, &statuses)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read library item status, got error: %s", err))
		return
	}

	for _, s := range statuses {
		data.Statuses = append(data.Statuses, libraryItemStatusModel{
			DeviceID:   types.StringValue(s.DeviceID),
			DeviceName: types.StringValue(s.DeviceName),
			Status:     types.StringValue(s.Status),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
