FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/main.go

FROM alpine:latest

# ✅ Install signal reaper (dumb-init) for proper signal handling
RUN apk add --no-cache dumb-init

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 8080

# Use dumb-init as PID1 to handle signals properly
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["./app"]

EXPOSE 8080

STOPSIGNAL SIGTERM

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./app"]
