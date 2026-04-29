package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomScriptResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
resource "iru_custom_script" "test" {
  name                = "Acc Test Script"
  active              = true
  execution_frequency = "once"
  script              = "#!/bin/zsh
echo 'hello'"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_custom_script.test", "name", "Acc Test Script"),
					resource.TestCheckResourceAttr("iru_custom_script.test", "active", "true"),
					resource.TestCheckResourceAttr("iru_custom_script.test", "execution_frequency", "once"),
					resource.TestCheckResourceAttrSet("iru_custom_script.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: `
resource "iru_custom_script" "test" {
  name                = "Acc Test Script Updated"
  active              = false
  execution_frequency = "every_day"
  script              = "#!/bin/zsh
echo 'updated'"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_custom_script.test", "name", "Acc Test Script Updated"),
					resource.TestCheckResourceAttr("iru_custom_script.test", "active", "false"),
					resource.TestCheckResourceAttr("iru_custom_script.test", "execution_frequency", "every_day"),
				),
			},
		},
	})
}
