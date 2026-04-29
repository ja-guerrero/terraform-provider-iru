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

var _ datasource.DataSource = &customProfilesDataSource{}

func NewCustomProfilesDataSource() datasource.DataSource {
	return &customProfilesDataSource{}
}

type customProfilesDataSource struct {
	client *client.Client
}

type customProfilesDataSourceModel struct {
	ID       types.String                   `tfsdk:"id"`
	Limit    types.Int64                    `tfsdk:"limit"`
	Offset   types.Int64                    `tfsdk:"offset"`
	Name     types.String                   `tfsdk:"name"`
	Profiles []customProfileDataSourceModel `tfsdk:"profiles"`
}

type customProfileDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Active        types.Bool   `tfsdk:"active"`
	MDMIdentifier types.String `tfsdk:"mdm_identifier"`
	RunsOnMac     types.Bool   `tfsdk:"runs_on_mac"`
	RunsOnIPhone  types.Bool   `tfsdk:"runs_on_iphone"`
	RunsOnIPad    types.Bool   `tfsdk:"runs_on_ipad"`
	RunsOnTV      types.Bool   `tfsdk:"runs_on_tv"`
	RunsOnVision  types.Bool   `tfsdk:"runs_on_vision"`
}

func (d *customProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_profiles"
}

func (d *customProfilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all custom profiles in the Iru instance.",
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
			"profiles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for the Custom Profile.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the Custom Profile.",
						},
						"active": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether this library item is active.",
						},
						"mdm_identifier": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The MDM identifier of the profile.",
						},
						"runs_on_mac": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the profile runs on macOS.",
						},
						"runs_on_iphone": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the profile runs on iOS.",
						},
						"runs_on_ipad": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the profile runs on iPadOS.",
						},
						"runs_on_tv": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the profile runs on tvOS.",
						},
						"runs_on_vision": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the profile runs on visionOS.",
						},
					},
				},
			},
		},
	}
}

func (d *customProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *customProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data customProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allProfiles []client.CustomProfile
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

		path := "/api/v1/library/custom-profiles?" + params.Encode()
		type listProfilesResponse struct {
			Results []client.CustomProfile `json:"results"`
		}
		var listResp listProfilesResponse

		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom profiles, got error: %s", err))
			return
		}

		allProfiles = append(allProfiles, listResp.Results...)

		if !data.Limit.IsNull() && len(allProfiles) >= limit {
			allProfiles = allProfiles[:limit]
			break
		}

		if len(listResp.Results) < limit {
			break
		}
		offset += len(listResp.Results)
	}

	data.ID = types.StringValue("custom_profiles")
	data.Profiles = make([]customProfileDataSourceModel, 0, len(allProfiles))
	for _, p := range allProfiles {
		data.Profiles = append(data.Profiles, customProfileDataSourceModel{
			ID:            types.StringValue(p.ID),
			Name:          types.StringValue(p.Name),
			Active:        types.BoolValue(p.Active),
			MDMIdentifier: types.StringValue(p.MDMIdentifier),
			RunsOnMac:     types.BoolValue(p.RunsOnMac),
			RunsOnIPhone:  types.BoolValue(p.RunsOnIPhone),
			RunsOnIPad:    types.BoolValue(p.RunsOnIPad),
			RunsOnTV:      types.BoolValue(p.RunsOnTV),
			RunsOnVision:  types.BoolValue(p.RunsOnVision),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
