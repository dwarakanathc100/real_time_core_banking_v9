# Stage 1 - Build
FROM golang:1.24.3-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
RUN go mod download
COPY . .
COPY .env .env
# Build the app binary
RUN CGO_ENABLED=0 GOOS=linux go build -o rtcb ./cmd/api/main.go

# Stage 2 - Runtime
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app
COPY --from=builder /app/rtcb .
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

EXPOSE 8080

# no hardcoded envs here; compose injects them at runtime
CMD ["./rtcb"]
