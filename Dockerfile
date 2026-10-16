FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY internal ./internal
ARG TARGETOS TARGETARCH
ARG VERSION=dev
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags="-s -w -X main.Version=${VERSION}" -o /out/web .

FROM alpine:latest AS app
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
CMD ["-addr", "0.0.0.0:8080"]
