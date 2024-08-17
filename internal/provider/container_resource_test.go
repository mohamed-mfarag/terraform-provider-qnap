package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccContainerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Existing test case 1
			{
				Config: `
					resource "qnap_container" "test" {
						name = "test-container"
						image = "nginx:latest"
						network = "eth0"
						status = "running"
						networktype = "bridge"
						type = "docker"
						removeanonvolumes = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("qnap_container.test", "name", "test-container"),
					resource.TestCheckResourceAttr("qnap_container.test", "image", "nginx:latest"),
					resource.TestCheckResourceAttr("qnap_container.test", "network", "eth0"),
					resource.TestCheckResourceAttr("qnap_container.test", "status", "running"),
					resource.TestCheckResourceAttr("qnap_container.test", "networktype", "bridge"),
					resource.TestCheckResourceAttr("qnap_container.test", "type", "docker"),
					resource.TestCheckResourceAttr("qnap_container.test", "removeanonvolumes", "true"),
				),
			},
			// Existing test case 2
			{
				Config: `
					resource "qnap_container" "test" {
						name = "test-container"
						image = "nginx:latest"
						network = "host"
						status = "stopped"
						networktype = "default"
						type = "docker"
						removeanonvolumes = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("qnap_container.test", "name", "test-container"),
					resource.TestCheckResourceAttr("qnap_container.test", "image", "nginx:latest"),
					resource.TestCheckResourceAttr("qnap_container.test", "network", "host"),
					resource.TestCheckResourceAttr("qnap_container.test", "status", "stopped"),
					resource.TestCheckResourceAttr("qnap_container.test", "networktype", "default"),
					resource.TestCheckResourceAttr("qnap_container.test", "type", "docker"),
					resource.TestCheckResourceAttr("qnap_container.test", "removeanonvolumes", "false"),
				),
			},
		},
	})
}
