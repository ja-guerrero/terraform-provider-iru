package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &selfServiceCategoriesDataSource{}

func NewSelfServiceCategoriesDataSource() datasource.DataSource {
	return &selfServiceCategoriesDataSource{}
}

type selfServiceCategoriesDataSource struct {
	client *client.Client
}

type selfServiceCategoriesDataSourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Name    types.String               `tfsdk:"name"`
	Results []selfServiceCategoryModel `tfsdk:"results"`
}

type selfServiceCategoryModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *selfServiceCategoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_self_service_categories"
}

func (d *selfServiceCategoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List self service categories.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by name.",
			},
			"results": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true},
						"name": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *selfServiceCategoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *selfServiceCategoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data selfServiceCategoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var listResp []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	err := d.client.DoRequest(ctx, "GET", "/api/v1/library/self-service/categories", nil, &listResp)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read self service categories, got error: %s", err))
		return
	}

	data.ID = types.StringValue("self_service_categories")
	for _, item := range listResp {
		if !data.Name.IsNull() && item.Name != data.Name.ValueString() {
			continue
		}
		data.Results = append(data.Results, selfServiceCategoryModel{
			ID:   types.StringValue(item.ID),
			Name: types.StringValue(item.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
