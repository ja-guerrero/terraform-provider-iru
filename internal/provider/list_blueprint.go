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

var _ list.ListResource = &blueprintListResource{}
var _ list.ListResourceWithConfigure = &blueprintListResource{}

func NewBlueprintListResource() list.ListResource {
	return &blueprintListResource{}
}

type blueprintListResource struct {
	client *client.Client
}

func (r *blueprintListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (r *blueprintListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{MarkdownDescription: "Lists Iru Blueprint resources."}
}

func (r *blueprintListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *blueprintListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var response struct {
		Results []client.Blueprint `json:"results"`
	}
	err := r.client.DoRequest(ctx, "GET", "/api/v1/blueprints", nil, &response)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list blueprints, got error: %v", err)),
		})
		return
	}

	results := make([]list.ListResult, 0, len(response.Results))
	for _, blueprint := range response.Results {
		result := req.NewListResult(ctx)

		identity := blueprintResourceIdentityModel{
			ID: types.StringValue(blueprint.ID),
		}
		result.Diagnostics.Append(result.Identity.Set(ctx, &identity)...)

		if req.IncludeResource {
			resourceModel := blueprintResourceModel{
				ID:             types.StringValue(blueprint.ID),
				Name:           types.StringValue(blueprint.Name),
				Description:    types.StringValue(blueprint.Description),
				Icon:           types.StringValue(blueprint.Icon),
				Color:          types.StringValue(blueprint.Color),
				EnrollmentCode: types.StringValue(blueprint.EnrollmentCode.Code),
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, &resourceModel)...)
		}

		result.DisplayName = blueprint.Name
		results = append(results, result)
	}

	resp.Results = slices.Values(results)
}
