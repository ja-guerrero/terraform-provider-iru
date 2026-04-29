package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
resource "iru_tag" "test" {
  name = "tf-acc-test"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_tag.test", "name", "tf-acc-test"),
					resource.TestCheckResourceAttrSet("iru_tag.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
