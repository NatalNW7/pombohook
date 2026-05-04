# Builder stage just to grab certificates and create the unprivileged user
FROM alpine:3.19 AS builder

RUN adduser -D -g '' -u 10001 pombohook
RUN apk add --no-cache ca-certificates

# Runtime stage
FROM scratch

# Import the user created in the builder
COPY --from=builder /etc/passwd /etc/passwd

# Import the SSL/TLS certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary (GoReleaser puts this in the build context)
COPY pombohook-server /usr/local/bin/pombohook-server

# Force the container to run with the unprivileged user
USER 10001

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pombohook-server"]
