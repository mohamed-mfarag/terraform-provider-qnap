package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccContainersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "qnap_containers" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of containers returned
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.#", "1"),
					// Verify the first container to ensure all attributes are set
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.id", "1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.name", "Container 1"),
					resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.type", "docker"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.image", "Image 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.imageid", "Image ID 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.status", "Status 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.project", "Project 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.runtime", "Runtime 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.memorylimit", "1024"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.cpulimit", "2"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.cpupin", "1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.uuid", "UUID 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.usedbyinternalservice", "Internal Service 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.privileged", "true"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.cpu", "1.5"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.memory", "512"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.tx", "100"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.rx", "200"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.read", "50"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.write", "75"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.created", "2022-01-01"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.startedat", "2022-01-01"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.cmd", "Command 1"),
					// // Verify the first container's port bindings
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.#", "2"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.0.host", "8080"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.0.container", "80"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.0.protocol", "tcp"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.0.hostip", "127.0.0.1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.0.containerip", "192.168.0.1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.1.host", "8443"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.1.container", "443"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.1.protocol", "tcp"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.1.hostip", "127.0.0.1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.portbindings.1.containerip", "192.168.0.2"),
					// // Verify the first container's networks
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.#", "2"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.0.id", "1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.0.name", "Network 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.0.displayname", "Network 1 Display"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.0.ipaddress", "192.168.0.10"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.0.macaddress", "00:11:22:33:44:55"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.0.gateway", "192.168.0.1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.0.networktype", "Type 1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.0.isstaticip", "true"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.1.id", "2"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.1.name", "Network 2"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.1.displayname", "Network 2 Display"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.1.ipaddress", "192.168.0.20"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.1.macaddress", "00:11:22:33:44:66"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.1.gateway", "192.168.0.1"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.1.networktype", "Type 2"),
					// resource.TestCheckResourceAttr("data.qnap_containers.test", "containers.0.networks.1.isstaticip", "false"),
				),
			},
		},
	})
}
