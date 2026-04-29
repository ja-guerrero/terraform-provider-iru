package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInHouseAppResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "iru_in_house_app" "test" {
  name     = "Acc Test In-House App"
  file_key = "apps/test.ipa"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_in_house_app.test", "name", "Acc Test In-House App"),
					resource.TestCheckResourceAttrSet("iru_in_house_app.test", "id"),
				),
			},
		},
	})
}
