package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ ephemeral.EphemeralResource = &blueprintOTAProfileEphemeralResource{}

func NewBlueprintOTAProfileEphemeralResource() ephemeral.EphemeralResource {
	return &blueprintOTAProfileEphemeralResource{}
}

type blueprintOTAProfileEphemeralResource struct {
	client *client.Client
}

type blueprintOTAProfileEphemeralResourceModel struct {
	BlueprintID types.String `tfsdk:"blueprint_id"`
	ProfileXML  types.String `tfsdk:"profile_xml"`
}

func (r *blueprintOTAProfileEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_ota_profile"
}

func (r *blueprintOTAProfileEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch the manual Over-the-Air (OTA) enrollment profile (.mobileconfig) for a specific blueprint. This is an ephemeral resource; the value is not stored in state.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The UUID of the blueprint.",
			},
			"profile_xml": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The XML content of the manual enrollment profile.",
			},
		},
	}
}

func (r *blueprintOTAProfileEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *blueprintOTAProfileEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data blueprintOTAProfileEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqRaw, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/blueprints/%s/ota-enrollment-profile", r.client.APIURL, data.BlueprintID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}
	reqRaw.Header.Set("Authorization", "Bearer "+r.client.APIToken)

	respRaw, err := r.client.HTTPClient.Do(reqRaw)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to perform request, got error: %s", err))
		return
	}
	defer respRaw.Body.Close()

	bodyBytes, _ := io.ReadAll(respRaw.Body)
	data.ProfileXML = types.StringValue(string(bodyBytes))

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
