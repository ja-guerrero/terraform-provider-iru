package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceRenameAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
action "iru_device_action_rename" "test" {
  device_id = "PLACEHOLDER"
  new_name  = "Terraform Renamed"
}
`,
			},
		},
	})
}
