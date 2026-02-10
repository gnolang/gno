#########################################
############### VARIABLES ###############
#########################################

variable "PROJECT_NAME" {
  default = "gno"
}

#########################################
################ GROUPS #################
#########################################

group "default" {
  targets = [
    "gno",
    "gnoland",
    "gnokey",
    "gnoweb", 
    "gnofaucet"
  ]
}

group "contribs" {
  targets = [
    "gnodev",
    "gnocontribs"
  ]
}

group "misc" {
  targets = ["portalloopd", "autocounterd"]
}

group "gnocore" {
    targets = ["default", "contribs"]
}

group "_all" {
    targets = ["gnocore", "misc"]
}

#########################################
############### TARGETS #################
#########################################

target "docker-metadata-action" {}

target "common" {
  inherits = ["docker-metadata-action"]
  attest = [
    "type=provenance,mode=max",
    "type=sbom",
  ]
  context = "../../../"
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ] 
  output = ["type=image"] # ,push=true -> Pushes to registry
}

target "gnoland" {
  inherits = ["common"]
  target = "gnoland"
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnoland"
  }
}

target "gnokey" {
  inherits = ["common"]
  target = "gnokey"
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnokey"
  }
}

target "gnoweb" {
  inherits = ["common"]
  target = "gnoweb"
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnoweb"
  }
}

target "gnofaucet" {
  inherits = ["common"]
  target = "gnofaucet"
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnofaucet"
  }
}

target "gno" {
  inherits = ["common"]
  target = "gno"
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gno"
  }
}

target "gnodev" {
  inherits = ["common"]
  target = "gnodev"
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnodev"
  }
}

target "gnocontribs" {
  inherits = ["common"]
  target = "gnocontribs"
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnocontribs"
  }
}

target "portalloopd" {
  inherits = ["common"]
  target = "portalloopd"
  tags = [
    "ghcr.io/gnolang/gno/portalloopd"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/portalloopd"
  }
}

target "autocounterd" {
  inherits = ["common"]
  target = "autocounterd"
  tags = [
    "ghcr.io/gnolang/gno/autocounterd"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/autocounterd"
  }
}
