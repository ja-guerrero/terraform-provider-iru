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

var _ datasource.DataSource = &customScriptsDataSource{}

func NewCustomScriptsDataSource() datasource.DataSource {
	return &customScriptsDataSource{}
}

type customScriptsDataSource struct {
	client *client.Client
}

type customScriptsDataSourceModel struct {
	ID      types.String                  `tfsdk:"id"`
	Limit   types.Int64                   `tfsdk:"limit"`
	Offset  types.Int64                   `tfsdk:"offset"`
	Name    types.String                  `tfsdk:"name"`
	Scripts []customScriptDataSourceModel `tfsdk:"scripts"`
}

type customScriptDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Active             types.Bool   `tfsdk:"active"`
	ExecutionFrequency types.String `tfsdk:"execution_frequency"`
	Restart            types.Bool   `tfsdk:"restart"`
	Script             types.String `tfsdk:"script"`
	RemediationScript  types.String `tfsdk:"remediation_script"`
	ShowInSelfService  types.Bool   `tfsdk:"show_in_self_service"`
}

func (d *customScriptsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_scripts"
}

func (d *customScriptsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all custom scripts in the Iru instance.",
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
			"scripts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for the Custom Script.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the Custom Script.",
						},
						"active": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether this library item is active.",
						},
						"execution_frequency": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The execution frequency.",
						},
						"restart": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether to restart the computer after script execution.",
						},
						"script": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The content of the script.",
						},
						"remediation_script": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The content of the remediation script.",
						},
						"show_in_self_service": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether to show the script in Self Service.",
						},
					},
				},
			},
		},
	}
}

func (d *customScriptsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *customScriptsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data customScriptsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allScripts []client.CustomScript
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

		path := "/api/v1/library/custom-scripts?" + params.Encode()
		type listScriptsResponse struct {
			Results []client.CustomScript `json:"results"`
		}
		var listResp listScriptsResponse

		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom scripts, got error: %s", err))
			return
		}

		allScripts = append(allScripts, listResp.Results...)

		if !data.Limit.IsNull() && len(allScripts) >= limit {
			allScripts = allScripts[:limit]
			break
		}

		if len(listResp.Results) < limit {
			break
		}
		offset += len(listResp.Results)
	}

	data.ID = types.StringValue("custom_scripts")
	data.Scripts = make([]customScriptDataSourceModel, 0, len(allScripts))
	for _, script := range allScripts {
		data.Scripts = append(data.Scripts, customScriptDataSourceModel{
			ID:                 types.StringValue(script.ID),
			Name:               types.StringValue(script.Name),
			Active:             types.BoolValue(script.Active),
			ExecutionFrequency: types.StringValue(script.ExecutionFrequency),
			Restart:            types.BoolValue(script.Restart),
			Script:             types.StringValue(script.Script),
			RemediationScript:  types.StringValue(script.RemediationScript),
			ShowInSelfService:  types.BoolValue(script.ShowInSelfService),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
