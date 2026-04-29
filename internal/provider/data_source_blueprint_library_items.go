package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &blueprintLibraryItemsDataSource{}

func NewBlueprintLibraryItemsDataSource() datasource.DataSource {
	return &blueprintLibraryItemsDataSource{}
}

type blueprintLibraryItemsDataSource struct {
	client *client.Client
}

type blueprintLibraryItemsDataSourceModel struct {
	BlueprintID  types.String                `tfsdk:"blueprint_id"`
	LibraryItems []blueprintLibraryItemModel `tfsdk:"library_items"`
}

type blueprintLibraryItemModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *blueprintLibraryItemsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_library_items"
}

func (d *blueprintLibraryItemsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all library items assigned to a specific blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Blueprint.",
			},
			"library_items": schema.ListNestedAttribute{
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

func (d *blueprintLibraryItemsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *blueprintLibraryItemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data blueprintLibraryItemsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var response struct {
		Results []client.BlueprintLibraryItem `json:"results"`
	}
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/blueprints/%s/list-library-items", data.BlueprintID.ValueString()), nil, &response)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read blueprint library items, got error: %s", err))
		return
	}

	for _, item := range response.Results {
		data.LibraryItems = append(data.LibraryItems, blueprintLibraryItemModel{
			ID:   types.StringValue(item.ID),
			Name: types.StringValue(item.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
