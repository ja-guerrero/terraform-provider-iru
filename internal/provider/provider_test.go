package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be called for every Terraform
// CLI command executed to create a provider server to which the CLI can
// connect.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"iru": providerserver.NewProtocol6WithError(New("test")()),
}

func TestNormalizeAPIURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://api.iru.io",
			expected: "https://api.iru.io",
		},
		{
			input:    "api.iru.io",
			expected: "https://api.iru.io",
		},
		{
			input:    "http://api.iru.io",
			expected: "http://api.iru.io",
		},
		{
			input:    "api.iru.io/",
			expected: "https://api.iru.io/",
		},
		{
			input:    "localhost:8080",
			expected: "https://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := normalizeAPIURL(tt.input)
			if actual != tt.expected {
				t.Errorf("normalizeAPIURL(%s) = %s; want %s", tt.input, actual, tt.expected)
			}
		})
	}
}
