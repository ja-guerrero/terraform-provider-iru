package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBlueprintResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
resource "iru_blueprint" "test" {
  name        = "Terraform Acceptance Test"
  description = "Created by Terraform Acceptance Test"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_blueprint.test", "name", "Terraform Acceptance Test"),
					resource.TestCheckResourceAttr("iru_blueprint.test", "description", "Created by Terraform Acceptance Test"),
					resource.TestCheckResourceAttrSet("iru_blueprint.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: `
resource "iru_blueprint" "test" {
  name        = "Terraform Acceptance Test Updated"
  description = "Updated by Terraform Acceptance Test"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("iru_blueprint.test", "name", "Terraform Acceptance Test Updated"),
					resource.TestCheckResourceAttr("iru_blueprint.test", "description", "Updated by Terraform Acceptance Test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
