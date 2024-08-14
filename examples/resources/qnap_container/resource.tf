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
      host        = "49116",
      container   = "6767",
      protocol    = "TCP",
      hostip      = "0.0.0.0",
      containerip = "",
    },

  ]
  volumes = [
    {
      type        = "volume"
      name        = "volume_1"
      container   = ""
      source      = "/ZFS530_DATA/.qpkg/container-station/docker/volumes/volume_1/_data"
      destination = "/config"
      permission  = "writable"
    },
  ]
}