package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceSecretsEphemeralResource(t *testing.T) {
	// Requires a valid device ID
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "iru_device_secrets" "test" {
  device_id = "PLACEHOLDER_DEVICE_ID"
}

output "albc" {
  value = ephemeral.iru_device_secrets.test.device_based_albc
  sensitive = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("iru_device_secrets.test", "device_based_albc"),
				),
			},
		},
	})
}
