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

var _ datasource.DataSource = &blueprintsDataSource{}

func NewBlueprintsDataSource() datasource.DataSource {
	return &blueprintsDataSource{}
}

type blueprintsDataSource struct {
	client *client.Client
}

type blueprintsDataSourceModel struct {
	ID         types.String             `tfsdk:"id"`
	Limit      types.Int64              `tfsdk:"limit"`
	Offset     types.Int64              `tfsdk:"offset"`
	Name       types.String             `tfsdk:"name"`
	Blueprints []blueprintListItemModel `tfsdk:"blueprints"`
}

type blueprintListItemModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Icon           types.String `tfsdk:"icon"`
	Color          types.String `tfsdk:"color"`
	EnrollmentCode types.String `tfsdk:"enrollment_code"`
}

func (d *blueprintsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprints"
}

func (d *blueprintsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all blueprints in the Iru instance.",
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
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by name.",
			},
			"blueprints": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for the Blueprint.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the Blueprint.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The description of the Blueprint.",
						},
						"icon": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The icon of the Blueprint.",
						},
						"color": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The color of the Blueprint.",
						},
						"enrollment_code": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The enrollment code for the Blueprint.",
						},
					},
				},
			},
		},
	}
}

func (d *blueprintsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *blueprintsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data blueprintsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allBlueprints []client.Blueprint
	offset := 0
	if !data.Offset.IsNull() {
		offset = int(data.Offset.ValueInt64())
	}
	limit := 300
	if !data.Limit.IsNull() {
		limit = int(data.Limit.ValueInt64())
	}

	for {
		params := url.Values{}
		params.Add("limit", fmt.Sprintf("%d", limit))
		params.Add("offset", fmt.Sprintf("%d", offset))

		if !data.Name.IsNull() {
			params.Add("name", data.Name.ValueString())
		}

		path := "/api/v1/blueprints?" + params.Encode()
		type listBlueprintsResponse struct {
			Results []client.Blueprint `json:"results"`
		}
		var listResp listBlueprintsResponse

		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read blueprints, got error: %s", err))
			return
		}

		allBlueprints = append(allBlueprints, listResp.Results...)

		if !data.Limit.IsNull() && len(allBlueprints) >= limit {
			allBlueprints = allBlueprints[:limit]
			break
		}

		if len(listResp.Results) < limit {
			break
		}
		offset += len(listResp.Results)
	}

	data.ID = types.StringValue("blueprints")
	data.Blueprints = make([]blueprintListItemModel, 0, len(allBlueprints))
	for _, blueprint := range allBlueprints {
		data.Blueprints = append(data.Blueprints, blueprintListItemModel{
			ID:             types.StringValue(blueprint.ID),
			Name:           types.StringValue(blueprint.Name),
			Description:    types.StringValue(blueprint.Description),
			Icon:           types.StringValue(blueprint.Icon),
			Color:          types.StringValue(blueprint.Color),
			EnrollmentCode: types.StringValue(blueprint.EnrollmentCode.Code),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
