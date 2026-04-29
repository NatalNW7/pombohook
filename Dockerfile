# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /pombohook-server ./cmd/server/

# Runtime stage
FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /pombohook-server /usr/local/bin/pombohook-server
ENTRYPOINT ["pombohook-server"]
