FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /caddy-sidecar .
RUN go build -o /caddy ./cmd/caddy

FROM alpine:3.24
RUN apk add --no-cache procps   # provides pgrep
COPY --from=builder /caddy-sidecar /usr/local/bin/caddy-sidecar
COPY --from=builder /caddy /usr/local/bin/caddy
EXPOSE 8080
ENTRYPOINT ["caddy-sidecar"]