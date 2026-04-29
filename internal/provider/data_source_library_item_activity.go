package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &libraryItemActivityDataSource{}

func NewLibraryItemActivityDataSource() datasource.DataSource {
	return &libraryItemActivityDataSource{}
}

type libraryItemActivityDataSource struct {
	client *client.Client
}

type libraryItemActivityDataSourceModel struct {
	LibraryItemID types.String               `tfsdk:"library_item_id"`
	Activity      []libraryItemActivityModel `tfsdk:"activity"`
}

type libraryItemActivityModel struct {
	DeviceID     types.String `tfsdk:"device_id"`
	DeviceName   types.String `tfsdk:"device_name"`
	Status       types.String `tfsdk:"status"`
	ActivityTime types.String `tfsdk:"activity_time"`
}

func (d *libraryItemActivityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_library_item_activity"
}

func (d *libraryItemActivityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get activity for a specific library item.",
		Attributes: map[string]schema.Attribute{
			"library_item_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Library Item.",
			},
			"activity": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"device_id":     schema.StringAttribute{Computed: true},
						"device_name":   schema.StringAttribute{Computed: true},
						"status":        schema.StringAttribute{Computed: true},
						"activity_time": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *libraryItemActivityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *libraryItemActivityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data libraryItemActivityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var activity []client.LibraryItemActivity
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/library/library-items/%s/activity", data.LibraryItemID.ValueString()), nil, &activity)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read library item activity, got error: %s", err))
		return
	}

	for _, a := range activity {
		data.Activity = append(data.Activity, libraryItemActivityModel{
			DeviceID:     types.StringValue(a.DeviceID),
			DeviceName:   types.StringValue(a.DeviceName),
			Status:       types.StringValue(a.Status),
			ActivityTime: types.StringValue(a.ActivityTime),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
