FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine AS api-builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
# Keep downloaded modules in an exported layer, rather than an ephemeral cache mount.
RUN go mod download
COPY main.go ./
COPY internal ./internal
COPY database ./database
COPY docs ./docs
ARG TARGETOS TARGETARCH
ARG VERSION=dev
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w -X 'github.com/fachschaftinformatik/web/internal/api.Version=${VERSION}'" -o /app/server .

FROM oven/bun:alpine AS bun-runtime
FROM node:26.12.0-alpine AS node-runtime
FROM bun-runtime AS website-deps
RUN apk add --no-cache libc6-compat git
WORKDIR /app
COPY website/package.json website/bun.lock ./
RUN bun install --frozen-lockfile

FROM node-runtime AS website-builder
WORKDIR /app
COPY --from=website-deps /app/node_modules ./node_modules
COPY website/ ./
COPY docs/ ../docs
RUN node node_modules/next/dist/bin/next build && \
    rm -rf .next/standalone/node_modules/@img/*-linux-x64 \
           .next/standalone/node_modules/@img/*-linux-arm64

FROM alpine:latest AS app
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=80
RUN apk add --no-cache ca-certificates tzdata libstdc++ libgcc libatomic s6 su-exec && \
    addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 -G nodejs nextjs && \
    mkdir -p /data /app/website/.next && \
    chown nextjs:nodejs /app/website/.next
COPY --from=node-runtime /usr/local/bin/node /usr/local/bin/node
COPY --from=api-builder /app/server ./server
COPY database/migrations ./database/migrations
COPY --from=website-builder --chown=nextjs:nodejs /app/public ./website/public
COPY --from=website-builder --chown=nextjs:nodejs /app/.next/standalone ./website/
COPY --from=website-builder --chown=nextjs:nodejs /app/.next/static ./website/.next/static
COPY --chmod=755 docker/entrypoint.sh /usr/local/bin/docker-entrypoint.sh
COPY --chmod=755 docker/healthcheck.sh /usr/local/bin/healthcheck
COPY --chmod=755 docker/services/ /etc/s6/
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 CMD ["/usr/local/bin/healthcheck"]
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["serve"]
