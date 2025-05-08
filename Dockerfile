# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Ensure all paths are relative to the WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mynuma_server ./cmd/server/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mynuma_server .
# If migrations are to be run by the container, they need to be copied.
# For now, assuming migrations are run separately or as part of a deploy script.
# COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./mynuma_server"]

