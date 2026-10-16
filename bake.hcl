variable "IMAGE_PREFIX" {
  default = "ghcr.io/fachschaftinformatik/web"
}

variable "RELEASE_TAG" {
  default = "dev"
}

group "default" {
  targets = ["app"]
}

target "common" {
  context   = "."
  platforms = ["linux/amd64"]
  attest = ["type=sbom"]
  labels = {
    "org.opencontainers.image.version" = RELEASE_TAG
  }
}

target "app" {
  inherits  = ["common"]
  dockerfile = "Dockerfile"
  args = {
    VERSION = RELEASE_TAG
  }
  tags       = ["${IMAGE_PREFIX}:${RELEASE_TAG}"]
  cache-from = ["type=gha,scope=app-amd64"]
  cache-to   = ["type=gha,mode=max,scope=app-amd64"]
}
