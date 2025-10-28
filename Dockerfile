FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY database/migrations ./database/migrations
RUN mkdir /data

EXPOSE 8080
CMD ["/app/server"]
