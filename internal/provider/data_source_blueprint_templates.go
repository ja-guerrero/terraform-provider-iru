package provider

import (
	"context"

	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &blueprintTemplatesDataSource{}

func NewBlueprintTemplatesDataSource() datasource.DataSource {
	return &blueprintTemplatesDataSource{}
}

type blueprintTemplatesDataSource struct {
	client *client.Client
}

type blueprintTemplatesDataSourceModel struct {
	Templates []blueprintTemplateModel `tfsdk:"templates"`
}

type blueprintTemplateModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *blueprintTemplatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_templates"
}

func (d *blueprintTemplatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List available Blueprint templates.",
		Attributes: map[string]schema.Attribute{
			"templates": schema.ListNestedAttribute{
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

func (d *blueprintTemplatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *blueprintTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data blueprintTemplatesDataSourceModel

	var response struct {
		Results []client.BlueprintTemplate `json:"results"`
	}
	err := d.client.DoRequest(ctx, "GET", "/api/v1/blueprints/templates/", nil, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read blueprint templates")
		return
	}

	for _, t := range response.Results {
		data.Templates = append(data.Templates, blueprintTemplateModel{
			ID:   types.StringValue(fmt.Sprintf("%d", t.ID)),
			Name: types.StringValue(t.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
