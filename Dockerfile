FROM --platform=$BUILDPLATFORM nixos/nix:2.34.9@sha256:53b5ec518b696373829ea305c3c07a8c870113e44648c88ec484ef078550fd4f AS assets
WORKDIR /src
COPY flake.nix flake.lock go.mod ./
COPY view/styles ./view/styles
COPY view/*.ts ./view/
RUN nix --extra-experimental-features "nix-command flakes" build .#assets --out-link /assets && \
    cp -RL /assets /generated-assets

FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go application.go ./
COPY internal ./internal
COPY model ./model
COPY site ./site
COPY storage ./storage
COPY view ./view
COPY --from=assets /generated-assets/ ./view/assets/
ARG TARGETOS TARGETARCH
ARG VERSION=dev
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags="-s -w -X main.Version=${VERSION}" -o /out/web .

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS app
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S -g 1001 web && adduser -S -u 1001 -G web web && \
    mkdir -p /data && chown web:web /data
WORKDIR /app
COPY --from=builder /out/web ./web
USER web:web
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/app/web", "-check-health", "http://127.0.0.1:8080/healthz"]
ENTRYPOINT ["/app/web"]
CMD ["-addr", "0.0.0.0:8080", "-database", "/data/sqlite.db"]
