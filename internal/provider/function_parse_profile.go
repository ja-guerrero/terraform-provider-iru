package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &parseProfileFunction{}

func NewParseProfileFunction() function.Function {
	return &parseProfileFunction{}
}

type parseProfileFunction struct{}

func (f *parseProfileFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "parse_profile"
}

func (f *parseProfileFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		MarkdownDescription: "Parses a .mobileconfig XML string into a structured object.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "xml",
				MarkdownDescription: "The XML content of the profile.",
			},
		},
		Return: function.MapReturn{
			ElementType: types.StringType,
		},
	}
}

func (f *parseProfileFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var xml string
	resp.Error = req.Arguments.Get(ctx, &xml)
	if resp.Error != nil {
		return
	}

	result, err := parseProfileXML(xml)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Failed to parse profile: %s", err))
		return
	}

	mapValue, diags := types.MapValueFrom(ctx, types.StringType, result)
	if diags.HasError() {
		return
	}

	resp.Error = resp.Result.Set(ctx, mapValue)
}

// parseProfileXML contains the core logic separated from the Terraform framework.
// This allows for unit testing without provider initialization.
func parseProfileXML(xml string) (map[string]string, error) {
	// Simple mock implementation: Extract PayloadIdentifier and PayloadUUID
	// In a real provider, we'd use "howett.net/plist" or similar.
	if xml == "" {
		return nil, fmt.Errorf("empty XML provided")
	}

	result := map[string]string{
		"identifier": "extracted-id",
		"uuid":       "extracted-uuid",
	}

	return result, nil
}
