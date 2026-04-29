package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomAppResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "iru_custom_app" "test" {
  name                = "Acc Test App"
  file_key            = "apps/test-v1.pkg"
  install_type        = "package"
  install_enforcement = "install_once"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_custom_app.test", "name", "Acc Test App"),
					resource.TestCheckResourceAttrSet("iru_custom_app.test", "id"),
				),
			},
		},
	})
}
