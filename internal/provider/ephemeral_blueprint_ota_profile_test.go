package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBlueprintOTAProfileEphemeralResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "iru_blueprint_ota_profile" "test" {
  blueprint_id = "c0148e35-c734-4402-b2fb-1c61aab72550"
}

output "profile" {
  value = ephemeral.iru_blueprint_ota_profile.test.profile_xml
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("iru_blueprint_ota_profile.test", "profile_xml"),
				),
			},
		},
	})
}
