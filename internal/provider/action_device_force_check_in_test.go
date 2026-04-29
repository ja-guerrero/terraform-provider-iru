package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceForceCheckInAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
action "iru_device_action_force_check_in" "test" {
  device_id = "PLACEHOLDER"
}
`,
			},
		},
	})
}
