# Build stage
FROM golang:1.24-alpine AS builder

# Create a non-root user to be used in the runtime
RUN adduser -D -g '' -u 10001 pombohook

# Ensure CA certificates exist
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Compile statically (CGO_ENABLED=0) and reduce size (-w -s)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /pombohook-server ./cmd/server/

# Runtime stage
FROM scratch

# Import the user created in the builder
COPY --from=builder /etc/passwd /etc/passwd

# Import the SSL/TLS certificates from Alpine's builder (required for the server to make HTTPS requests and connect to external services if needed)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /pombohook-server /usr/local/bin/pombohook-server

# Force the container to run with the unprivileged user (pombohook/10001)
USER 10001

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pombohook-server"]
