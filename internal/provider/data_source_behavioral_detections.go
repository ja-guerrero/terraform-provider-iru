package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &behavioralDetectionsDataSource{}

func NewBehavioralDetectionsDataSource() datasource.DataSource {
	return &behavioralDetectionsDataSource{}
}

type behavioralDetectionsDataSource struct {
	client *client.Client
}

type behavioralDetectionsDataSourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Page    types.Int64                `tfsdk:"page"`
	Results []behavioralDetectionModel `tfsdk:"results"`
}

type behavioralDetectionModel struct {
	ID             types.String `tfsdk:"id"`
	ThreatID       types.String `tfsdk:"threat_id"`
	Description    types.String `tfsdk:"description"`
	Classification types.String `tfsdk:"classification"`
	DetectionDate  types.String `tfsdk:"detection_date"`
	ThreatStatus   types.String `tfsdk:"threat_status"`
	DeviceID       types.String `tfsdk:"device_id"`
	DeviceName     types.String `tfsdk:"device_name"`
}

func (d *behavioralDetectionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_behavioral_detections"
}

func (d *behavioralDetectionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List behavioral detections.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"page": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Page number to fetch.",
			},
			"results": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"threat_id":      schema.StringAttribute{Computed: true},
						"description":    schema.StringAttribute{Computed: true},
						"classification": schema.StringAttribute{Computed: true},
						"detection_date": schema.StringAttribute{Computed: true},
						"threat_status":  schema.StringAttribute{Computed: true},
						"device_id":      schema.StringAttribute{Computed: true},
						"device_name":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *behavioralDetectionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *behavioralDetectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data behavioralDetectionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var all []client.BehavioralDetection
	page := 1
	if !data.Page.IsNull() {
		page = int(data.Page.ValueInt64())
	}

	for {
		type bdResponse struct {
			Results []client.BehavioralDetection `json:"results"`
			Next    string                       `json:"next"`
		}
		var listResp bdResponse

		path := fmt.Sprintf("/api/v1/behavioral-detections?page=%d", page)
		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read behavioral detections, got error: %s", err))
			return
		}

		all = append(all, listResp.Results...)

		if !data.Page.IsNull() {
			break
		}

		if listResp.Next == "" || len(listResp.Results) == 0 {
			break
		}
		page++
		if page > 10 {
			break
		} // Safety
	}

	data.ID = types.StringValue("behavioral_detections")
	data.Results = make([]behavioralDetectionModel, 0, len(all))
	for _, item := range all {
		data.Results = append(data.Results, behavioralDetectionModel{
			ID:             types.StringValue(item.ID),
			ThreatID:       types.StringValue(item.ThreatID),
			Description:    types.StringValue(item.Description),
			Classification: types.StringValue(item.Classification),
			DetectionDate:  types.StringValue(item.DetectionDate),
			ThreatStatus:   types.StringValue(item.ThreatStatus),
			DeviceID:       types.StringValue(item.DeviceInfo.ID),
			DeviceName:     types.StringValue(item.DeviceInfo.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
