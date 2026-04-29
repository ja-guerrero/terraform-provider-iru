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

var _ list.ListResource = &customScriptListResource{}
var _ list.ListResourceWithConfigure = &customScriptListResource{}

func NewCustomScriptListResource() list.ListResource {
	return &customScriptListResource{}
}

type customScriptListResource struct {
	client *client.Client
}

func (r *customScriptListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_script"
}

func (r *customScriptListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{MarkdownDescription: "Lists Iru Custom Script resources."}
}

func (r *customScriptListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customScriptListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var response struct {
		Results []client.CustomScript `json:"results"`
	}
	err := r.client.DoRequest(ctx, "GET", "/api/v1/library/custom-scripts", nil, &response)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list custom scripts, got error: %v", err)),
		})
		return
	}

	results := make([]list.ListResult, 0, len(response.Results))
	for _, script := range response.Results {
		result := req.NewListResult(ctx)

		identity := customScriptResourceIdentityModel{
			ID: types.StringValue(script.ID),
		}
		result.Diagnostics.Append(result.Identity.Set(ctx, &identity)...)

		if req.IncludeResource {
			resourceModel := customScriptResourceModel{
				ID:                 types.StringValue(script.ID),
				Name:               types.StringValue(script.Name),
				Active:             types.BoolValue(script.Active),
				ExecutionFrequency: types.StringValue(script.ExecutionFrequency),
				Restart:            types.BoolValue(script.Restart),
				Script:             types.StringValue(script.Script),
				RemediationScript:  types.StringValue(script.RemediationScript),
				ShowInSelfService:  types.BoolValue(script.ShowInSelfService),
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, &resourceModel)...)
		}

		result.DisplayName = script.Name
		results = append(results, result)
	}

	resp.Results = slices.Values(results)
}
