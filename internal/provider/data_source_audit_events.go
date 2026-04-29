package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &auditEventsDataSource{}

func NewAuditEventsDataSource() datasource.DataSource {
	return &auditEventsDataSource{}
}

type auditEventsDataSource struct {
	client *client.Client
}

type auditEventsDataSourceModel struct {
	ID        types.String      `tfsdk:"id"`
	Limit     types.Int64       `tfsdk:"limit"`
	SortBy    types.String      `tfsdk:"sort_by"`
	StartDate types.String      `tfsdk:"start_date"`
	EndDate   types.String      `tfsdk:"end_date"`
	Results   []auditEventModel `tfsdk:"results"`
}

type auditEventModel struct {
	ID              types.String `tfsdk:"id"`
	Action          types.String `tfsdk:"action"`
	OccurredAt      types.String `tfsdk:"occurred_at"`
	ActorID         types.String `tfsdk:"actor_id"`
	ActorType       types.String `tfsdk:"actor_type"`
	TargetID        types.String `tfsdk:"target_id"`
	TargetType      types.String `tfsdk:"target_type"`
	TargetComponent types.String `tfsdk:"target_component"`
}

func (d *auditEventsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_events"
}

func (d *auditEventsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List audit log events from the Activity module. This data source provides a global view of actions performed in the Iru instance, including blueprint changes, device enrollment, and administrative actions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "A max upper limit is set at 500 records returned per request.",
			},
			"sort_by": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Sort results by occurred_at, id either ascending (default behavior) or descending(-) order.",
			},
			"start_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by start date in datetime or year-month-day (2024-11-26) formats.",
			},
			"end_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by start date in datetime or year-month-day (2024-12-06) formats.",
			},
			"results": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"action":           schema.StringAttribute{Computed: true},
						"occurred_at":      schema.StringAttribute{Computed: true},
						"actor_id":         schema.StringAttribute{Computed: true},
						"actor_type":       schema.StringAttribute{Computed: true},
						"target_id":        schema.StringAttribute{Computed: true},
						"target_type":      schema.StringAttribute{Computed: true},
						"target_component": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *auditEventsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *auditEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data auditEventsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var all []client.AuditEvent
	cursor := ""
	limit := 500
	if !data.Limit.IsNull() {
		limit = int(data.Limit.ValueInt64())
	}

	for {
		type auditResponse struct {
			Results []client.AuditEvent `json:"results"`
			Next    string              `json:"next"`
		}
		var listResp auditResponse

		params := url.Values{}
		params.Add("limit", fmt.Sprintf("%d", limit))
		if cursor != "" {
			params.Add("cursor", cursor)
		}
		if !data.SortBy.IsNull() {
			params.Add("sort_by", data.SortBy.ValueString())
		}
		if !data.StartDate.IsNull() {
			params.Add("start_date", data.StartDate.ValueString())
		}
		if !data.EndDate.IsNull() {
			params.Add("end_date", data.EndDate.ValueString())
		}

		path := "/api/v1/audit/events?" + params.Encode()

		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read audit events, got error: %s", err))
			return
		}

		all = append(all, listResp.Results...)

		// If user specified a limit and we reached it, stop.
		if !data.Limit.IsNull() && len(all) >= limit {
			all = all[:limit]
			break
		}

		if listResp.Next == "" || len(listResp.Results) == 0 {
			break
		}

		// Extract cursor from Next URL
		nextURL, err := url.Parse(listResp.Next)
		if err != nil {
			break
		}
		cursor = nextURL.Query().Get("cursor")
		if cursor == "" {
			break
		}
	}

	data.ID = types.StringValue("audit_events")
	data.Results = make([]auditEventModel, 0, len(all))
	for _, item := range all {
		data.Results = append(data.Results, auditEventModel{
			ID:              types.StringValue(item.ID),
			Action:          types.StringValue(item.Action),
			OccurredAt:      types.StringValue(item.OccurredAt),
			ActorID:         types.StringValue(item.ActorID),
			ActorType:       types.StringValue(item.ActorType),
			TargetID:        types.StringValue(item.TargetID),
			TargetType:      types.StringValue(item.TargetType),
			TargetComponent: types.StringValue(item.TargetComponent),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
