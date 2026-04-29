package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &threatsDataSource{}

func NewThreatsDataSource() datasource.DataSource {
	return &threatsDataSource{}
}

type threatsDataSource struct {
	client *client.Client
}

type threatsDataSourceModel struct {
	ID      types.String  `tfsdk:"id"`
	Limit   types.Int64   `tfsdk:"limit"`
	Offset  types.Int64   `tfsdk:"offset"`
	Results []threatModel `tfsdk:"results"`
}

type threatModel struct {
	ThreatName         types.String `tfsdk:"threat_name"`
	Classification     types.String `tfsdk:"classification"`
	Status             types.String `tfsdk:"status"`
	DeviceName         types.String `tfsdk:"device_name"`
	DeviceID           types.String `tfsdk:"device_id"`
	DetectionDate      types.String `tfsdk:"detection_date"`
	FilePath           types.String `tfsdk:"file_path"`
	FileHash           types.String `tfsdk:"file_hash"`
	DeviceSerialNumber types.String `tfsdk:"device_serial_number"`
}

func (d *threatsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_threats"
}

func (d *threatsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List detected threats.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of results to return.",
			},
			"offset": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of results to skip.",
			},
			"results": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"threat_name":          schema.StringAttribute{Computed: true},
						"classification":       schema.StringAttribute{Computed: true},
						"status":               schema.StringAttribute{Computed: true},
						"device_name":          schema.StringAttribute{Computed: true},
						"device_id":            schema.StringAttribute{Computed: true},
						"detection_date":       schema.StringAttribute{Computed: true},
						"file_path":            schema.StringAttribute{Computed: true},
						"file_hash":            schema.StringAttribute{Computed: true},
						"device_serial_number": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *threatsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *threatsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data threatsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var all []client.Threat
	offset := 0
	if !data.Offset.IsNull() {
		offset = int(data.Offset.ValueInt64())
	}
	limit := 300
	if !data.Limit.IsNull() {
		limit = int(data.Limit.ValueInt64())
	}

	for {
		type threatResponse struct {
			Results []client.Threat `json:"results"`
		}
		var listResp threatResponse

		path := fmt.Sprintf("/api/v1/threat-details?limit=%d&offset=%d", limit, offset)
		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read threats, got error: %s", err))
			return
		}

		all = append(all, listResp.Results...)

		if !data.Limit.IsNull() && len(all) >= limit {
			all = all[:limit]
			break
		}

		if len(listResp.Results) < limit {
			break
		}
		offset += len(listResp.Results)
	}

	for _, item := range all {
		data.Results = append(data.Results, threatModel{
			ThreatName:         types.StringValue(item.ThreatName),
			Classification:     types.StringValue(item.Classification),
			Status:             types.StringValue(item.Status),
			DeviceName:         types.StringValue(item.DeviceName),
			DeviceID:           types.StringValue(item.DeviceID),
			DetectionDate:      types.StringValue(item.DetectionDate),
			FilePath:           types.StringValue(item.FilePath),
			FileHash:           types.StringValue(item.FileHash),
			DeviceSerialNumber: types.StringValue(item.DeviceSerialNumber),
		})
	}

	data.ID = types.StringValue("threats")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
