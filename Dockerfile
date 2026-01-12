FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go generate ./... && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/server .
COPY database/migrations ./database/migrations
RUN mkdir -p /data
EXPOSE 8080
CMD ["/app/server"]
