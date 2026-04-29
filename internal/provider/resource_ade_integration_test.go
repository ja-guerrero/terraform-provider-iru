package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccADEIntegrationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "iru_ade_integration" "test" {
  blueprint_id          = "c0148e35-c734-4402-b2fb-1c61aab72550"
  phone                 = "555-555-5555"
  email                 = "test@example.com"
  mdm_server_token_file = "mock-token-content"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_ade_integration.test", "phone", "555-555-5555"),
					resource.TestCheckResourceAttr("iru_ade_integration.test", "email", "test@example.com"),
					resource.TestCheckResourceAttrSet("iru_ade_integration.test", "id"),
				),
			},
		},
	})
}
