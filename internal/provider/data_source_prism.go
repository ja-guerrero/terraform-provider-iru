package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &prismDataSource{}

func NewPrismDataSource() datasource.DataSource {
	return &prismDataSource{}
}

type prismDataSource struct {
	client *client.Client
}

type prismDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Category types.String `tfsdk:"category"`
	Limit    types.Int64  `tfsdk:"limit"`
	Offset   types.Int64  `tfsdk:"offset"`
	Results  types.List   `tfsdk:"results"`
}

func (d *prismDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prism"
}

func (d *prismDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Query any Prism reporting category. Each result entry is returned as a map of string key-value pairs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"category": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Prism category to query.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"activation_lock",
						"application_firewall",
						"apps",
						"cellular",
						"certificates",
						"desktop_and_screensaver",
						"device_information",
						"filevault",
						"gatekeeper_and_xprotect",
						"installed_profiles",
						"kernel_extensions",
						"launch_agents_and_daemons",
						"local_users",
						"startup_settings",
						"system_extensions",
						"transparency_database",
					),
				},
			},
			"limit": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of results to return.",
			},
			"offset": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of results to skip.",
			},
			"results": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "List of result entries. Each entry is a map of string key-value pairs.",
				ElementType:         types.MapType{ElemType: types.StringType},
			},
		},
	}
}

func (d *prismDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *prismDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data prismDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	category := data.Category.ValueString()

	var all []client.PrismEntry
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

		path := "/api/v1/prism/" + category + "?" + params.Encode()
		type prismResponse struct {
			Data []client.PrismEntry `json:"data"`
		}
		var listResp prismResponse

		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read prism %s, got error: %s", category, err))
			return
		}

		all = append(all, listResp.Data...)

		if !data.Limit.IsNull() && len(all) >= limit {
			all = all[:limit]
			break
		}

		if len(listResp.Data) < limit {
			break
		}
		offset += len(listResp.Data)
	}

	data.ID = types.StringValue("prism_" + category)

	// Convert each PrismEntry (map[string]interface{}) to map[string]string.
	resultMaps := make([]map[string]string, 0, len(all))
	for _, entry := range all {
		m := make(map[string]string, len(entry))
		for k, v := range entry {
			m[k] = fmt.Sprintf("%v", v)
		}
		resultMaps = append(resultMaps, m)
	}

	resultsList, diags := types.ListValueFrom(ctx, types.MapType{ElemType: types.StringType}, resultMaps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Results = resultsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
