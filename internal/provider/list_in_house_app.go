package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ list.ListResource = &inHouseAppListResource{}
var _ list.ListResourceWithConfigure = &inHouseAppListResource{}

func NewInHouseAppListResource() list.ListResource {
	return &inHouseAppListResource{}
}

type inHouseAppListResource struct {
	client *client.Client
}

func (r *inHouseAppListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_in_house_app"
}

func (r *inHouseAppListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{MarkdownDescription: "Lists Iru In House App resources."}
}

func (r *inHouseAppListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *inHouseAppListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var response struct {
		Results []client.InHouseApp `json:"results"`
	}
	err := r.client.DoRequest(ctx, "GET", "/api/v1/library/ipa-apps", nil, &response)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list in-house apps, got error: %v", err)),
		})
		return
	}

	results := make([]list.ListResult, 0, len(response.Results))
	for _, app := range response.Results {
		result := req.NewListResult(ctx)

		identity := inHouseAppResourceIdentityModel{
			ID: types.StringValue(app.ID),
		}
		result.Diagnostics.Append(result.Identity.Set(ctx, &identity)...)

		if req.IncludeResource {
			resourceModel := inHouseAppResourceModel{
				ID:           types.StringValue(app.ID),
				Name:         types.StringValue(app.Name),
				FileKey:      types.StringValue(app.FileKey),
				RunsOnIPhone: types.BoolValue(app.RunsOnIPhone),
				RunsOnIPad:   types.BoolValue(app.RunsOnIPad),
				RunsOnTV:     types.BoolValue(app.RunsOnTV),
				Active:       types.BoolValue(app.Active),
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, &resourceModel)...)
		}

		result.DisplayName = app.Name
		results = append(results, result)
	}

	resp.Results = slices.Values(results)
}
