package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceAppsDataSource{}

func NewDeviceAppsDataSource() datasource.DataSource {
	return &deviceAppsDataSource{}
}

type deviceAppsDataSource struct {
	client *client.Client
}

type deviceAppsDataSourceModel struct {
	DeviceID types.String      `tfsdk:"device_id"`
	Apps     []deviceAppsModel `tfsdk:"apps"`
}

type deviceAppsModel struct {
	Name     types.String `tfsdk:"name"`
	Version  types.String `tfsdk:"version"`
	BundleID types.String `tfsdk:"bundle_id"`
}

func (d *deviceAppsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_apps"
}

func (d *deviceAppsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List apps installed on a specific device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"apps": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":      schema.StringAttribute{Computed: true},
						"version":   schema.StringAttribute{Computed: true},
						"bundle_id": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *deviceAppsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceAppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceAppsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apps []struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		BundleID string `json:"bundle_id"`
	}
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/apps", data.DeviceID.ValueString()), nil, &apps)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device apps, got error: %s", err))
		return
	}

	for _, app := range apps {
		data.Apps = append(data.Apps, deviceAppsModel{
			Name:     types.StringValue(app.Name),
			Version:  types.StringValue(app.Version),
			BundleID: types.StringValue(app.BundleID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
