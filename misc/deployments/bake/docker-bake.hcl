#########################################
############### VARIABLES ###############
#########################################

variable "TAG" {
  default = "chain-test6"
}

variable "PROJECT_NAME" {
  default = "gno"
}

variable "DATE" {
  default = "2025-03-31"
}

variable "FULL_COMMIT" {
  default = "959e90374dfea8afa791df43352ec6dbac3f30ed"
}

variable "VERSION" {
  default = "1.0.0"
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

target "common" {
  attest = [
    "type=provenance,mode=max",
    "type=sbom",
  ]
  context = "../../../"
  
  labels = {
    "org.opencontainers.image.created" = "${DATE}"
    "org.opencontainers.image.title" = "${PROJECT_NAME}"
    "org.opencontainers.image.revision" = "${FULL_COMMIT}"
    "org.opencontainers.image.version" = "${VERSION}"
    "org.opencontainers.image.authors" = "Gno Core Team"
  }
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
    "ghcr.io/gnolang/gno/gnoweb:gnoland-home"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnoweb"
  }
}

target "gnofaucet" {
  inherits = ["common"]
  target = "gnofaucet"
  tags = [
    "ghcr.io/gnolang/gno/mygnofaucet"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnofaucet"
  }
}

target "gno" {
  inherits = ["common"]
  target = "gno"
  tags = [
    "ghcr.io/gnolang/gno"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gno"
  }
}

target "gnodev" {
  inherits = ["common"]
  target = "gnodev"
  tags = [
    "ghcr.io/gnolang/gno/gnodev"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnodev"
  }
}

target "gnocontribs" {
  inherits = ["common"]
  target = "gnocontribs"
  tags = [
    "ghcr.io/gnolang/gno/gnocontribs"
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
