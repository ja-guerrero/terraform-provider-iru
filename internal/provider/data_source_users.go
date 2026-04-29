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

var _ datasource.DataSource = &usersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

type usersDataSource struct {
	client *client.Client
}

type usersDataSourceModel struct {
	ID     types.String        `tfsdk:"id"`
	Limit  types.Int64         `tfsdk:"limit"`
	Offset types.Int64         `tfsdk:"offset"`
	Name   types.String        `tfsdk:"name"`
	Email  types.String        `tfsdk:"email"`
	Users  []userListItemModel `tfsdk:"users"`
}

type userListItemModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Email      types.String `tfsdk:"email"`
	IsArchived types.Bool   `tfsdk:"is_archived"`
}

func (d *usersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all users in the Iru instance.",
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
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by email.",
			},
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier for the User.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the User.",
						},
						"email": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The email of the User.",
						},
						"is_archived": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the user is archived.",
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data usersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var allUsers []client.User
	offset := 0
	if !data.Offset.IsNull() {
		offset = int(data.Offset.ValueInt64())
	}
	limit := 300
	if !data.Limit.IsNull() {
		limit = int(data.Limit.ValueInt64())
	}

	for {
		type listUsersResponse struct {
			Results []client.User `json:"results"`
		}
		var listResp listUsersResponse

		params := url.Values{}
		params.Add("limit", fmt.Sprintf("%d", limit))
		params.Add("offset", fmt.Sprintf("%d", offset))

		if !data.Name.IsNull() {
			params.Add("name", data.Name.ValueString())
		}
		if !data.Email.IsNull() {
			params.Add("email", data.Email.ValueString())
		}

		path := "/api/v1/users?" + params.Encode()
		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read users, got error: %s", err))
			return
		}

		allUsers = append(allUsers, listResp.Results...)

		if !data.Limit.IsNull() && len(allUsers) >= limit {
			allUsers = allUsers[:limit]
			break
		}

		if len(listResp.Results) < limit {
			break
		}
		offset += len(listResp.Results)
	}

	data.ID = types.StringValue("users")
	data.Users = make([]userListItemModel, 0, len(allUsers))
	for _, user := range allUsers {
		data.Users = append(data.Users, userListItemModel{
			ID:         types.StringValue(user.ID),
			Name:       types.StringValue(user.Name),
			Email:      types.StringValue(user.Email),
			IsArchived: types.BoolValue(user.IsArchived),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
