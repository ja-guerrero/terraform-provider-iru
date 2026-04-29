package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

// identityGetOrNil retrieves the resource identity from the request as a pointer.
// Returns nil if identity is not set (e.g., state predates identity support).
// The caller should check for nil before accessing identity fields.
func identityGetOrNil[T any](ctx context.Context, identity interface {
	Get(context.Context, interface{}) diag.Diagnostics
}, diagnostics *diag.Diagnostics) *T {
	var result *T
	diagnostics.Append(identity.Get(ctx, &result)...)
	return result
}

// resolveID returns the resource ID from state data, falling back to the identity
// value if the state ID is empty. This handles the case where identity exists but
// the state ID was not yet populated (e.g., during import).
func resolveID(dataID types.String, identityID *types.String) string {
	id := dataID.ValueString()
	if id == "" && identityID != nil {
		id = identityID.ValueString()
	}
	return id
}

// paginatedGet fetches all pages from a paginated Kandji API endpoint.
// The endpoint must return JSON in the shape: {"data": [...], "results": [...]}
// or a plain array. The dataKey parameter specifies which JSON key holds the array
// (typically "data" for Prism endpoints, "results" for list endpoints).
func paginatedGet[T any](ctx context.Context, c *client.Client, basePath string, userLimit *int, userOffset *int) ([]T, error) {
	offset := 0
	if userOffset != nil {
		offset = *userOffset
	}
	pageSize := 300
	maxResults := 0
	if userLimit != nil {
		maxResults = *userLimit
	}

	var all []T
	for {
		params := url.Values{}
		params.Add("limit", fmt.Sprintf("%d", pageSize))
		params.Add("offset", fmt.Sprintf("%d", offset))

		path := basePath + "?" + params.Encode()

		type paginatedResponse struct {
			Data    []T `json:"data"`
			Results []T `json:"results"`
		}
		var resp paginatedResponse

		if err := c.DoRequest(ctx, "GET", path, nil, &resp); err != nil {
			return nil, err
		}

		// Use whichever key is populated
		page := resp.Data
		if len(page) == 0 {
			page = resp.Results
		}

		all = append(all, page...)

		if maxResults > 0 && len(all) >= maxResults {
			all = all[:maxResults]
			break
		}

		if len(page) < pageSize {
			break
		}
		offset += len(page)
	}

	return all, nil
}
