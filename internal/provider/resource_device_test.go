package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceResource(t *testing.T) {
	// Since devices cannot be created, we test Import
	// This requires an existing device ID in the environment or a mock.
	// For acceptance tests, we usually expect certain env vars.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "iru_device" "test" {
  # This is just to satisfy the schema for import testing
}
`,
				// ImportState testing
				ResourceName:      "iru_device.test",
				ImportState:       true,
				ImportStateId:     "PLACEHOLDER_DEVICE_ID", // User should replace this or we use an env var
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// Ignore fields that might change frequently
				},
			},
		},
	})
}
