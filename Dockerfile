# Build stage
FROM golang:1.24-alpine AS builder

# Cria um usuário non-root para ser usado no runtime
RUN adduser -D -g '' -u 10001 pombohook

# Garante a existência dos certificados CA
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Compila estaticamente (CGO_ENABLED=0) e reduz o tamanho (-w -s)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /pombohook-server ./cmd/server/

# Runtime stage
FROM scratch

# Importa o usuário criado no builder
COPY --from=builder /etc/passwd /etc/passwd

# Importa os certificados SSL/TLS da Alpine do builder (necessário para o servidor fazer requisições HTTPS e se conectar a coisas se necessário)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copia o binário compilado
COPY --from=builder /pombohook-server /usr/local/bin/pombohook-server

# Força o contêiner a rodar com o usuário sem privilégios (pombohook/10001)
USER 10001

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pombohook-server"]
