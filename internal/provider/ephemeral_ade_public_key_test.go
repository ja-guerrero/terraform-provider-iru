package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccADEPublicKeyEphemeralResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "iru_ade_public_key" "test" {}

output "key" {
  value = ephemeral.iru_ade_public_key.test.public_key
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("iru_ade_public_key.test", "public_key"),
				),
			},
		},
	})
}
