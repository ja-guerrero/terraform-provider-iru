package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDevicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "iru_devices" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// We can't guarantee there are devices, but we check if the attribute exists
					resource.TestCheckResourceAttrSet("data.iru_devices.test", "devices.#"),
				),
			},
		},
	})
}
