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

var _ list.ListResource = &customProfileListResource{}
var _ list.ListResourceWithConfigure = &customProfileListResource{}

func NewCustomProfileListResource() list.ListResource {
	return &customProfileListResource{}
}

type customProfileListResource struct {
	client *client.Client
}

func (r *customProfileListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_profile"
}

func (r *customProfileListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{MarkdownDescription: "Lists Iru Custom Profile resources."}
}

func (r *customProfileListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customProfileListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var response struct {
		Results []client.CustomProfile `json:"results"`
	}
	err := r.client.DoRequest(ctx, "GET", "/api/v1/library/custom-profiles", nil, &response)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list custom profiles, got error: %v", err)),
		})
		return
	}

	results := make([]list.ListResult, 0, len(response.Results))
	for _, profile := range response.Results {
		result := req.NewListResult(ctx)

		identity := customProfileResourceIdentityModel{
			ID: types.StringValue(profile.ID),
		}
		result.Diagnostics.Append(result.Identity.Set(ctx, &identity)...)

		if req.IncludeResource {
			resourceModel := customProfileResourceModel{
				ID:            types.StringValue(profile.ID),
				Name:          types.StringValue(profile.Name),
				Active:        types.BoolValue(profile.Active),
				MDMIdentifier: types.StringValue(profile.MDMIdentifier),
				RunsOnMac:     types.BoolValue(profile.RunsOnMac),
				RunsOnIPhone:  types.BoolValue(profile.RunsOnIPhone),
				RunsOnIPad:    types.BoolValue(profile.RunsOnIPad),
				RunsOnTV:      types.BoolValue(profile.RunsOnTV),
				RunsOnVision:  types.BoolValue(profile.RunsOnVision),
			}
			// ProfileFile is not returned in list usually, so we don't set it here to avoid empty string
			result.Diagnostics.Append(result.Resource.Set(ctx, &resourceModel)...)
		}

		result.DisplayName = profile.Name
		results = append(results, result)
	}

	resp.Results = slices.Values(results)
}
