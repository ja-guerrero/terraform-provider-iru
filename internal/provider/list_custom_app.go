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

var _ list.ListResource = &customAppListResource{}
var _ list.ListResourceWithConfigure = &customAppListResource{}

func NewCustomAppListResource() list.ListResource {
	return &customAppListResource{}
}

type customAppListResource struct {
	client *client.Client
}

func (r *customAppListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_app"
}

func (r *customAppListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{MarkdownDescription: "Lists Iru Custom App resources."}
}

func (r *customAppListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customAppListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var response struct {
		Results []client.CustomApp `json:"results"`
	}
	err := r.client.DoRequest(ctx, "GET", "/api/v1/library/custom-apps", nil, &response)
	if err != nil {
		resp.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Client Error", fmt.Sprintf("Unable to list custom apps, got error: %v", err)),
		})
		return
	}

	results := make([]list.ListResult, 0, len(response.Results))
	for _, app := range response.Results {
		result := req.NewListResult(ctx)

		identity := customAppResourceIdentityModel{
			ID: types.StringValue(app.ID),
		}
		result.Diagnostics.Append(result.Identity.Set(ctx, &identity)...)

		if req.IncludeResource {
			resourceModel := customAppResourceModel{
				ID:                     types.StringValue(app.ID),
				Name:                   types.StringValue(app.Name),
				FileKey:                types.StringValue(app.FileKey),
				InstallType:            types.StringValue(app.InstallType),
				InstallEnforcement:     types.StringValue(app.InstallEnforcement),
				UnzipLocation:          types.StringValue(app.UnzipLocation),
				AuditScript:            types.StringValue(app.AuditScript),
				PreinstallScript:       types.StringValue(app.PreinstallScript),
				PostinstallScript:      types.StringValue(app.PostinstallScript),
				ShowInSelfService:      types.BoolValue(app.ShowInSelfService),
				SelfServiceCategoryID:  types.StringValue(app.SelfServiceCategoryID),
				SelfServiceRecommended: types.BoolValue(app.SelfServiceRecommended),
				Active:                 types.BoolValue(app.Active),
				Restart:                types.BoolValue(app.Restart),
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, &resourceModel)...)
		}

		result.DisplayName = app.Name
		results = append(results, result)
	}

	resp.Results = slices.Values(results)
}
