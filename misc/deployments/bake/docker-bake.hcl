#########################################
############### VARIABLES ###############
#########################################

variable "TAG" {
  default = "chain-test6"
}

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

group "_gno" { # overlaps the gno single target
    targets = ["default", "contribs"]
}

group "_all" {
    targets = ["gno", "misc"]
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
  dockerfile = "Dockerfile.new"
  
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ] 
  output = ["type=image"] # ,push=true -> Pushes to registry
}

target "gnoland" {
  inherits = ["common"]
  target = "gnoland"
  tags = [
    "ghcr.io/gnolang/gno/gnoland:${TAG}"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnoland"
  }
}

target "gnokey" {
  inherits = ["common"]
  target = "gnokey"
  tags = [
    "ghcr.io/gnolang/gno/gnokey:${TAG}"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnokey"
  }
}

target "gnoweb" {
  inherits = ["common"]
  target = "gnoweb"
  tags = [
    "ghcr.io/gnolang/gno/gnoweb:${TAG}"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnoweb"
  }
}

target "gnofaucet" {
  inherits = ["common"]
  target = "gnofaucet"
  tags = [
    "ghcr.io/gnolang/gno/gnofaucet:${TAG}"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnofaucet"
  }
}

target "gno" {
  inherits = ["common"]
  target = "gno"
  tags = [
    "ghcr.io/gnolang/gno:${TAG}"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gno"
  }
}

target "gnodev" {
  inherits = ["common"]
  target = "gnodev"
  tags = [
    "ghcr.io/gnolang/gno/gnodev:${TAG}"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnodev"
  }
}

target "gnocontribs" {
  inherits = ["common"]
  target = "gnocontribs"
  tags = [
    "ghcr.io/gnolang/gno/gnocontribs:${TAG}"
  ]
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
