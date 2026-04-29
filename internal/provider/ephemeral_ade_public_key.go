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

var _ ephemeral.EphemeralResource = &adePublicKeyEphemeralResource{}

func NewADEPublicKeyEphemeralResource() ephemeral.EphemeralResource {
	return &adePublicKeyEphemeralResource{}
}

type adePublicKeyEphemeralResource struct {
	client *client.Client
}

type adePublicKeyEphemeralResourceModel struct {
	PublicKey types.String `tfsdk:"public_key"`
}

func (r *adePublicKeyEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ade_public_key"
}

func (r *adePublicKeyEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch the ADE public key required to create an MDM server connection in Apple Business Manager. This is an ephemeral resource; the value is not stored in state.",
		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The PEM-encoded public key.",
			},
		},
	}
}

func (r *adePublicKeyEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *adePublicKeyEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data adePublicKeyEphemeralResourceModel

	// Need a custom GET that returns raw string
	reqRaw, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/integrations/apple/ade/public_key/", r.client.APIURL), nil)
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
	data.PublicKey = types.StringValue(string(bodyBytes))

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
