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

var _ list.ListResource = &adeIntegrationListResource{}
var _ list.ListResourceWithConfigure = &adeIntegrationListResource{}

func NewADEIntegrationListResource() list.ListResource {
	return &adeIntegrationListResource{}
}

type adeIntegrationListResource struct {
	client *client.Client
}

func (r *adeIntegrationListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ade_integration"
}

func (r *adeIntegrationListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{MarkdownDescription: "Lists Iru Ade Integration resources."}
	resp.Schema = listschema.Schema{
		MarkdownDescription: "Lists Iru ADE Integrations.",
	}
}

func (r *adeIntegrationListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *adeIntegrationListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var response struct {
		Results []client.ADEIntegration `json:"results"`
	}
	err := r.client.DoRequest(ctx, "GET", "/api/v1/integrations/apple/ade/", nil, &response)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list ADE integrations, got error: %v", err)),
		})
		return
	}

	results := make([]list.ListResult, 0, len(response.Results))
	for _, integration := range response.Results {
		result := req.NewListResult(ctx)

		identity := adeIntegrationResourceIdentityModel{
			ID: types.StringValue(integration.ID),
		}
		result.Diagnostics.Append(result.Identity.Set(ctx, &identity)...)

		if req.IncludeResource {
			phone := integration.Phone
			if phone == "" {
				phone = integration.Defaults.Phone
			}
			email := integration.Email
			if email == "" {
				email = integration.Defaults.Email
			}

			resourceModel := adeIntegrationResourceModel{
				ID:                  types.StringValue(integration.ID),
				Phone:               types.StringValue(phone),
				Email:               types.StringValue(email),
				UseBlueprintRouting: types.BoolValue(integration.UseBlueprintRouting),
			}
			if integration.Blueprint != nil {
				resourceModel.BlueprintID = types.StringValue(integration.Blueprint.ID)
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, &resourceModel)...)
		}

		display := integration.Email
		if display == "" {
			display = integration.Defaults.Email
		}
		result.DisplayName = display
		results = append(results, result)
	}

	resp.Results = slices.Values(results)
}
