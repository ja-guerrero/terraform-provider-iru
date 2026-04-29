package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomProfileResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
resource "iru_custom_profile" "test" {
  name         = "Acc Test Profile"
  active       = true
  profile_file = <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>PayloadDisplayName</key>
	<string>Test</string>
</dict>
</plist>
EOF
  runs_on_mac  = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_custom_profile.test", "name", "Acc Test Profile"),
					resource.TestCheckResourceAttr("iru_custom_profile.test", "active", "true"),
					resource.TestCheckResourceAttrSet("iru_custom_profile.test", "id"),
					resource.TestCheckResourceAttrSet("iru_custom_profile.test", "mdm_identifier"),
				),
			},
		},
	})
}
