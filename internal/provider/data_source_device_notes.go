package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &deviceNotesDataSource{}

func NewDeviceNotesDataSource() datasource.DataSource {
	return &deviceNotesDataSource{}
}

type deviceNotesDataSource struct {
	client *client.Client
}

type deviceNotesDataSourceModel struct {
	DeviceID types.String      `tfsdk:"device_id"`
	Notes    []deviceNoteModel `tfsdk:"notes"`
}

type deviceNoteModel struct {
	ID        types.String `tfsdk:"id"`
	Content   types.String `tfsdk:"content"`
	Author    types.String `tfsdk:"author"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *deviceNotesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_notes"
}

func (d *deviceNotesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all notes associated with a specific device. Each note includes content, author, and timestamp information.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier for the Device.",
			},
			"notes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"content":    schema.StringAttribute{Computed: true},
						"author":     schema.StringAttribute{Computed: true},
						"created_at": schema.StringAttribute{Computed: true},
						"updated_at": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *deviceNotesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *deviceNotesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data deviceNotesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var notes []client.DeviceNote
	err := d.client.DoRequest(ctx, "GET", fmt.Sprintf("/api/v1/devices/%s/notes", data.DeviceID.ValueString()), nil, &notes)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device notes, got error: %s", err))
		return
	}

	for _, n := range notes {
		data.Notes = append(data.Notes, deviceNoteModel{
			ID:        types.StringValue(n.ID),
			Content:   types.StringValue(n.Content),
			Author:    types.StringValue(n.Author),
			CreatedAt: types.StringValue(n.CreatedAt),
			UpdatedAt: types.StringValue(n.UpdatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
