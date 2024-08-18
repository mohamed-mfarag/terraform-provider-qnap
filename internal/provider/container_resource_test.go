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
					resource "qnap_container" "test_container_1" {
						name = "terraform_test_container_1"
						image = "nginx:latest"
						network = "eth0"
						status = "running"
						networktype = "bridge"
						type = "docker"
						removeanonvolumes = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("qnap_container.test_container_1", "name", "terraform_test_container_1"),
					resource.TestCheckResourceAttr("qnap_container.test_container_1", "image", "nginx:latest"),
					resource.TestCheckResourceAttr("qnap_container.test_container_1", "network", "eth0"),
					resource.TestCheckResourceAttr("qnap_container.test_container_1", "status", "running"),
					resource.TestCheckResourceAttr("qnap_container.test_container_1", "networktype", "bridge"),
					resource.TestCheckResourceAttr("qnap_container.test_container_1", "type", "docker"),
					resource.TestCheckResourceAttr("qnap_container.test_container_1", "removeanonvolumes", "true"),
				),
			},
			// Existing test case 2
			{
				Config: `
					resource "qnap_container" "test_container_2" {
						name = "terraform_test_container_2"
						image = "nginx:latest"
						network = "host"
						status = "stopped"
						networktype = "default"
						type = "docker"
						removeanonvolumes = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("qnap_container.test_container_2", "name", "terraform_test_container_2"),
					resource.TestCheckResourceAttr("qnap_container.test_container_2", "image", "nginx:latest"),
					resource.TestCheckResourceAttr("qnap_container.test_container_2", "network", "host"),
					resource.TestCheckResourceAttr("qnap_container.test_container_2", "status", "stopped"),
					resource.TestCheckResourceAttr("qnap_container.test_container_2", "networktype", "default"),
					resource.TestCheckResourceAttr("qnap_container.test_container_2", "type", "docker"),
					resource.TestCheckResourceAttr("qnap_container.test_container_2", "removeanonvolumes", "false"),
				),
			},
			{
				Config: `
					resource "qnap_container" "full_coverage" {
						status             = "running"
						removeanonvolumes  = true
						type               = "docker"
						name               = "terrafrom_testing_demo"
						image              = "nginx:latest"
						network            = "bridge"
						networktype        = "default"
						hostname           = "my-hostname"
						privileged         = false
					
						portbindings {
							[
								host      = 49123
								container = 80
								protocol  = "tcp"
								hostip    = "0.0.0.0"
								containerip = "10.0.0.100"
							]
						}
					
						restartpolicy {
						name             = "on-failure"
						maximumretrycount = 5
						}
					
						autoremove   = false
						cmd          = ["nginx", "-g", "daemon off;"]
						entrypoint   = ["/bin/sh"]
						tty          = true
						openstdin    = true
					
						runtime     = "runc"
									
						dns = ["8.8.8.8", "8.8.4.4"]
						env = {
						"ENV_VAR1" = "value1"
						"ENV_VAR2" = "value2"
						}
						labels = {
						"app" = "my-app"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "status", "running"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "removeanonvolumes", "true"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "type", "docker"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "name", "terraform_full_coverage"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "image", "nginx:latest"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "network", "bridge"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "networktype", "default"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "hostname", "my-hostname"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "privileged", "false"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "portbindings.0.host", "49123"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "portbindings.0.container", "80"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "portbindings.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "portbindings.0.hostip", "0.0.0.0"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "portbindings.0.containerip", "10.0.0.100"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "restartpolicy.name", "on-failure"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "restartpolicy.maximumretrycount", "5"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "autoremove", "false"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "cmd.0", "nginx"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "cmd.1", "-g"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "cmd.2", "daemon off;"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "entrypoint.0", "/bin/sh"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "tty", "true"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "openstdin", "true"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "runtime", "runc"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "dns.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "dns.1", "8.8.4.4"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "env.ENV_VAR1", "value1"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "env.ENV_VAR2", "value2"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage", "labels.app", "my-app"),
				),
			},
		},
	})
}

// func TestAccContainerResourceMinimal(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccContainerResourceMinimalConfig,
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("qnap_container_resource.minimal", "status", "running"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.minimal", "removeanonvolumes", "true"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.minimal", "type", "docker"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.minimal", "name", "minimal-container"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.minimal", "image", "nginx:latest"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.minimal", "network", "bridge"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.minimal", "networktype", "bridge"),
// 				),
// 			},
// 		},
// 	})
// }

// func TestAccContainerResourceInvalidStatusType(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccContainerResourceInvalidStatusTypeConfig,
// 				ExpectError: regexp.MustCompile(`Invalid status or type`),
// 			},
// 		},
// 	})
// }

// func TestAccContainerResourceInvalidNetwork(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccContainerResourceInvalidNetworkConfig,
// 				ExpectError: regexp.MustCompile(`Invalid network or networktype`),
// 			},
// 		},
// 	})
// }

// func TestAccContainerResourceEmptyOptionalFields(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccContainerResourceEmptyOptionalFieldsConfig,
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("qnap_container_resource.empty_optional_fields", "status", "running"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.empty_optional_fields", "removeanonvolumes", "true"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.empty_optional_fields", "type", "docker"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.empty_optional_fields", "name", "empty-optional-fields"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.empty_optional_fields", "image", "nginx:latest"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.empty_optional_fields", "network", "bridge"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.empty_optional_fields", "networktype", "bridge"),
// 				),
// 			},
// 		},
// 	})
// }

// func TestAccContainerResourcePortBindingsEdge(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccContainerResourcePortBindingsEdgeConfig,
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("qnap_container_resource.port_bindings_edge", "portbindings.0.host", "0"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.port_bindings_edge", "portbindings.0.container", "65535"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.port_bindings_edge", "portbindings.0.protocol", "tcp"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.port_bindings_edge", "portbindings.0.hostip", "0.0.0.0"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.port_bindings_edge", "portbindings.0.containerip", "255.255.255.255"),
// 				),
// 			},
// 		},
// 	})
// }

// func TestAccContainerResourceDnsEnv(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccContainerResourceDnsEnvConfig,
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("qnap_container_resource.dns_env", "dns.#", "2"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.dns_env", "env.KEY1", "value1"),
// 					resource.TestCheckResourceAttr("qnap_container_resource.dns_env", "env.KEY2", "value2"),
// 				),
// 			},
// 		},
// 	})
// }

// Add additional test cases as needed.
