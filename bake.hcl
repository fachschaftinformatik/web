variable "IMAGE_PREFIX" {
  default = "ghcr.io/fachschaftinformatik/web"
}

variable "RELEASE_TAG" {
  default = "dev"
}

group "default" {
  targets = ["api", "website"]
}

target "common" {
  context   = "."
  platforms = ["linux/amd64"]
  labels = {
    "org.opencontainers.image.version" = RELEASE_TAG
  }
}

target "api" {
  inherits  = ["common"]
  dockerfile = "Dockerfile"
  args = {
    VERSION = RELEASE_TAG
  }
  tags       = ["${IMAGE_PREFIX}/api:${RELEASE_TAG}"]
  cache-from = ["type=gha,scope=api-amd64"]
  cache-to   = ["type=gha,mode=max,scope=api-amd64"]
}

target "website" {
  inherits   = ["common"]
  dockerfile = "website/Dockerfile"
  tags       = ["${IMAGE_PREFIX}/website:${RELEASE_TAG}"]
  cache-from = ["type=gha,scope=website-amd64"]
  cache-to   = ["type=gha,mode=max,scope=website-amd64"]
}
