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
			// test case 1
			{
				Config: `
					resource "qnap_container" "min_coverage" {
						name = "terraform_test_min_coverage"
						image = "nginx:latest"
						network = "eth0"
						status = "running"
						networktype = "bridge"
						type = "docker"
						removeanonvolumes = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("qnap_container.min_coverage", "name", "terraform_test_min_coverage"),
					resource.TestCheckResourceAttr("qnap_container.min_coverage", "image", "nginx:latest"),
					resource.TestCheckResourceAttr("qnap_container.min_coverage", "network", "eth0"),
					resource.TestCheckResourceAttr("qnap_container.min_coverage", "status", "running"),
					resource.TestCheckResourceAttr("qnap_container.min_coverage", "networktype", "bridge"),
					resource.TestCheckResourceAttr("qnap_container.min_coverage", "type", "docker"),
					resource.TestCheckResourceAttr("qnap_container.min_coverage", "removeanonvolumes", "true"),
				),
			},
			// test case 2
			{
				Config: `
					resource "qnap_container" "full_coverage_1" {
						status             = "running"
						removeanonvolumes  = true
						type               = "docker"
						name               = "terraform_test_full_coverage_1"
						image              = "nginx:1.26.2"
						network            = "bridge"
						networktype        = "default"
						hostname           = "my-hostname"
						privileged         = false
						portbindings = [
							{
								host      = 49123,
								container = 80,
								protocol  = "tcp",
								hostip    = "0.0.0.0",
							}
						]
						restartpolicy = {
							name              = "onFailure"
							maximumretrycount = 5
						}
						autoremove   = false
						cmd          = ["nginx", "-g", "daemon off;"]
						entrypoint   = ["/docker-entrypoint.sh"]
						tty          = true
						openstdin    = true
						runtime     = "runc"
						volumes	 = [
								{
									type = "volume",
									name = "terraform_test_full_coverage_volume",
									destination = "/terraform_test_full_coverage_volume",
									permission = "writable",
									container = "",
									source = "/ZFS530_DATA/.qpkg/container-station/docker/volumes/terraform_test_full_coverage_volume/_data",
								},
								{
									type = "host",
									source = "/Container/container-station-data/volumes/demo",
									destination = "/demo",
									permission = "writable",
									container = "",
									name = "",
								},
							]
									
						dns = ["8.8.8.8", "8.8.4.4"]
						env = {
							"NGINX_VERSION" = "1.26.2"
							"NJS_VERSION" = "0.8.5"
							"NJS_RELEASE" = "1~bookworm"
							"PKG_RELEASE" = "1~bookworm"
							"DYNPKG_RELEASE" = "2~bookworm"
							"PATH" = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "status", "running"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "removeanonvolumes", "true"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "type", "docker"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "name", "terraform_test_full_coverage_1"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "image", "nginx:1.26.2"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "network", "bridge"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "networktype", "default"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "hostname", "my-hostname"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "privileged", "false"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "portbindings.0.host", "49123"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "portbindings.0.container", "80"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "portbindings.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "portbindings.0.hostip", "0.0.0.0"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "restartpolicy.name", "onFailure"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "restartpolicy.maximumretrycount", "5"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "autoremove", "false"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "cmd.0", "nginx"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "cmd.1", "-g"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "cmd.2", "daemon off;"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "entrypoint.0", "/docker-entrypoint.sh"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "tty", "true"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "openstdin", "true"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "runtime", "runc"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "dns.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "dns.1", "8.8.4.4"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "env.DYNPKG_RELEASE", "2~bookworm"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "env.PKG_RELEASE", "1~bookworm"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "env.NGINX_VERSION", "1.26.2"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "env.NJS_RELEASE", "1~bookworm"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "env.NJS_VERSION", "0.8.5"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "env.PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.#", "2"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.0.type", "volume"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.0.name", "terraform_test_full_coverage_volume"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.0.destination", "/terraform_test_full_coverage_volume"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.0.permission", "writable"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.0.source", "/ZFS530_DATA/.qpkg/container-station/docker/volumes/terraform_test_full_coverage_volume/_data"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.0.container", ""),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.1.type", "host"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.1.source", "/Container/container-station-data/volumes/demo"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.1.destination", "/demo"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.1.permission", "writable"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.1.container", ""),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_1", "volumes.1.name", ""),
				),
			},
			// test case 3
			{
				Config: `
					resource "qnap_container" "full_coverage_2" {
						status             = "running"
						removeanonvolumes  = true
						type               = "docker"
						name               = "terraform_test_full_coverage_2"
						image              = "nginx:1.26.2"
						network            = "eth0"
						networktype        = "bridge"
						hostname           = "my-hostname"
						privileged         = false
						ipaddress 		   = "192.168.178.233"
						portbindings = [
							{
								host      = 49123,
								container = 80,
								protocol  = "tcp",
								hostip    = "0.0.0.0",
							}
						]
						autoremove   = true
						cmd          = ["nginx", "-g", "daemon off;"]
						entrypoint   = ["/docker-entrypoint.sh"]
						tty          = false
						openstdin    = false
						runtime     = "runc"
						volumes	 = [
								{
									type = "volume",
									name = "terraform_test_full_coverage_volume",
									destination = "/terraform_test_full_coverage_volume",
									permission = "writable",
									container = "",
									source = "/ZFS530_DATA/.qpkg/container-station/docker/volumes/terraform_test_full_coverage_volume/_data",
								},
								{
									type = "host",
									source = "/Container/container-station-data/volumes/demo",
									destination = "/demo",
									permission = "writable",
									container = "",
									name = "",
								},
							]	
						dns = []
						env = {
							"NGINX_VERSION" = "1.26.2"
							"NJS_VERSION" = "0.8.5"
							"NJS_RELEASE" = "1~bookworm"
							"PKG_RELEASE" = "1~bookworm"
							"DYNPKG_RELEASE" = "2~bookworm"
							"PATH" = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "status", "running"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "removeanonvolumes", "true"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "type", "docker"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "name", "terraform_test_full_coverage_2"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "image", "nginx:1.26.2"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "network", "eth0"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "networktype", "bridge"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "ipaddress", "192.168.178.233"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "hostname", "my-hostname"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "privileged", "false"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "portbindings.0.host", "49123"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "portbindings.0.container", "80"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "portbindings.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "portbindings.0.hostip", "0.0.0.0"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "autoremove", "true"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "cmd.0", "nginx"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "cmd.1", "-g"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "cmd.2", "daemon off;"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "entrypoint.0", "/docker-entrypoint.sh"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "tty", "false"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "openstdin", "false"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "runtime", "runc"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "dns.#", "0"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "env.DYNPKG_RELEASE", "2~bookworm"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "env.PKG_RELEASE", "1~bookworm"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "env.NGINX_VERSION", "1.26.2"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "env.NJS_RELEASE", "1~bookworm"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "env.NJS_VERSION", "0.8.5"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "env.PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.0.type", "volume"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.0.name", "terraform_test_full_coverage_volume"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.0.destination", "/terraform_test_full_coverage_volume"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.0.permission", "writable"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.0.source", "/ZFS530_DATA/.qpkg/container-station/docker/volumes/terraform_test_full_coverage_volume/_data"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.0.container", ""),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.1.type", "host"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.1.source", "/Container/container-station-data/volumes/demo"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.1.destination", "/demo"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.1.permission", "writable"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.1.container", ""),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "volumes.1.name", ""),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "networks.0.ipaddress", "192.168.178.233"),
					resource.TestCheckResourceAttr("qnap_container.full_coverage_2", "networks.0.isstaticip", "true"),
				),
			},
		},
	})
}
