package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceCommandsDataSource{}

func NewDeviceCommandsDataSource() datasource.DataSource {
	return &deviceCommandsDataSource{}
}

type deviceCommandsDataSource struct {
	client *client.Client
}

type deviceCommandsDataSourceModel struct {
	DeviceID types.String         `tfsdk:"device_id"`
	Limit    types.Int64          `tfsdk:"limit"`
	Commands []deviceCommandModel `tfsdk:"commands"`
}

type deviceCommandModel struct {
	UUID          types.String `tfsdk:"uuid"`
	CommandType   types.String `tfsdk:"command_type"`
	Status        types.Int64  `tfsdk:"status"`
	DateRequested types.String `tfsdk:"date_requested"`
	DateCompleted types.String `tfsdk:"date_completed"`
}

func (d *deviceCommandsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_commands"
}

func (d *deviceCommandsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List commands sent to a device.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Limit results (max 300).",
			},
			"commands": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed: true,
						},
						"command_type": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.Int64Attribute{
							Computed: true,
						},
						"date_requested": schema.StringAttribute{
							Computed: true,
						},
						"date_completed": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *deviceCommandsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceCommandsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceCommandsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limit := int64(300)
	if !data.Limit.IsNull() {
		limit = data.Limit.ValueInt64()
	}

	var response client.DeviceCommandList
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/commands?limit=%d", data.DeviceID.ValueString(), limit), nil, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device commands, got error: %s", err))
		return
	}

	for _, cmd := range response.Commands.Results {
		data.Commands = append(data.Commands, deviceCommandModel{
			UUID:          types.StringValue(cmd.UUID),
			CommandType:   types.StringValue(cmd.CommandType),
			Status:        types.Int64Value(int64(cmd.Status)),
			DateRequested: types.StringValue(cmd.DateRequested),
			DateCompleted: types.StringValue(cmd.DateCompleted),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
