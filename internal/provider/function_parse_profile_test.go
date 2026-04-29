package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestParseProfileXML is a standard Go unit test.
// It does NOT require TF_ACC=1 or an API Token.
func TestParseProfileXML(t *testing.T) {
	xml := "mock-xml"
	result, err := parseProfileXML(xml)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["identifier"] != "extracted-id" {
		t.Errorf("Expected identifier extracted-id, got %s", result["identifier"])
	}

	if result["uuid"] != "extracted-uuid" {
		t.Errorf("Expected uuid extracted-uuid, got %s", result["uuid"])
	}
}

// TestAccParseProfileFunction is a Terraform Acceptance Test.
// It requires TF_ACC=1 and a valid IRU_API_TOKEN.
func TestAccParseProfileFunction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
output "test" {
  value = provider::iru::parse_profile("mock-xml")
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "{identifier = \"extracted-id\", uuid = \"extracted-uuid\"}"),
				),
			},
		},
	})
}
