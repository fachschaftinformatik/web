FROM --platform=$BUILDPLATFORM nixos/nix:2.35.3@sha256:ff480dbb23511aa1c10710aa068ceea5b93fed9bfce5f5379f3a6b07282cbc7b AS assets
WORKDIR /src
COPY flake.nix flake.lock go.mod ./
COPY view/styles ./view/styles
COPY view/*.ts ./view/
RUN nix --extra-experimental-features "nix-command flakes" build .#assets --out-link /assets && \
    cp -RL /assets /generated-assets

FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine@sha256:f2c872bcee949b7be9c490efd02b48a8a2fbf947cdf533f1c58525ec9d8170fd AS builder
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

FROM alpine:3.24.3@sha256:b3413572228e34bd1c9245f5565e6cd127005655e8a7899aef33d8e83255c9e6 AS app
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
