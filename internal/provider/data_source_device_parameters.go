package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceParametersDataSource{}

func NewDeviceParametersDataSource() datasource.DataSource {
	return &deviceParametersDataSource{}
}

type deviceParametersDataSource struct {
	client *client.Client
}

type deviceParametersDataSourceModel struct {
	DeviceID   types.String            `tfsdk:"device_id"`
	Parameters []deviceParametersModel `tfsdk:"parameters"`
}

type deviceParametersModel struct {
	ID     types.String `tfsdk:"id"`
	Status types.String `tfsdk:"status"`
}

func (d *deviceParametersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_parameters"
}

func (d *deviceParametersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List parameters and their status for a specific device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
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

func (d *deviceParametersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceParametersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceParametersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var params []struct {
		ID     string `json:"parameter_id"`
		Status string `json:"status"`
	}
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/parameters", data.DeviceID.ValueString()), nil, &params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device parameters, got error: %s", err))
		return
	}

	for _, p := range params {
		data.Parameters = append(data.Parameters, deviceParametersModel{
			ID:     types.StringValue(p.ID),
			Status: types.StringValue(p.Status),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
