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
  status            = "stopped"
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
      protocol    = "tcp",
      hostip      = "0.0.0.0",
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
resource "qnap_app" "demo_app" {
  status            = "stopped"
  name              = "postgresql-test"
  removeanonvolumes = true
  yml               = "version: '3'\nservices:\n  postgres:\n    image: postgres:15.1\n    restart: always\n    ports:\n      - 127.0.0.1:5432:5432\n    volumes:\n      - postgres_db:/var/lib/postgresql/data\n    environment:\n      POSTGRES_USER: postgres_qnap_user\n      POSTGRES_PASSWORD: postgres_qnap_pwd\n\n  phppgadmin:\n    image: qnapsystem/phppgadmin:7.13.0-1\n    restart: on-failure\n    ports:\n      - 7070:80\n    depends_on:\n      - postgres\n    environment:\n      PHP_PG_ADMIN_SERVER_HOST: postgres\n      PHP_PG_ADMIN_SERVER_PORT: 5432\n\nvolumes:\n  postgres_db:\n"
}

output "bazarr10" {
  value = qnap_container.bazarr10
}
output "demo_app" {
  value = qnap_app.demo_app
}

