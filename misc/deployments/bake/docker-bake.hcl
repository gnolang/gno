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

group "default" {
  targets = ["gnoland", "gnokey", "gnoweb"]
}

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
  }
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ] 
  output = ["type=image,push=true"] # Push to registry
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
  # context = "https://github.com/alexiscolin/gno.git#refactor/gnoland-home"
  tags = [
    "ghcr.io/gnolang/gno/gnoweb:gnoland-home"
  ]
  labels = {
    "org.opencontainers.image.title" = "${PROJECT_NAME}/gnoweb"
  }
}
