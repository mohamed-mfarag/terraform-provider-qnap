terraform {
  required_providers {
    qnap = {
      source = "hashicorp.com/edu/qnap"
    }
  }
}

provider "qnap" {
  host     = "https://server.venus.home"
  username = "mohamed"
  password = "Moh@Med1"
}

resource "qnap_container" "bazarr10" {
  name              = "bazarr-10"
  image             = "linuxserver/bazarr:latest"
  type              = "docker"
  network           = "bridge"
  networktype       = "default"
  removeanonvolumes = true
  restartpolicy = {
    name : "always",
    maximumretrycount : 0,
  }
  cpupin = {
    cpuids : "",
    type : "",
  }
  portbindings = [
    {
    host      = "49116",
    container = "6767",
    protocol  = "TCP",
    hostip    = "0.0.0.0",
    containerip = "",
  },
      {
    host      = "49119",
    container = "6766",
    protocol  = "TCP",
    hostip    = "0.0.0.0",
    containerip = "",
  },

  ]
    volumes = [
    {
      type        = "volume"
      name        = "volume_1"
      container   = "/config"
      source      = "/ZFS530_DATA/.qpkg/container-station/docker/volumes/volume_1/_data"
      destination = "/config"
      permission  = "writable"
    },
  ]
}

output "bazarr10" {
  value = qnap_container.bazarr10
}