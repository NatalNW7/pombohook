# PomboHook — Future Features

## Dashboard de Logs
- [ ] `GET /logs` — Retorna os últimos N webhooks recebidos
- [ ] Informações por log: path, method, status code, timestamp, response time
- [ ] Filtro por path (ex: `/logs?path=/webhooks/mercadopago`)
- [ ] UI web simples para visualização (opcional)

## Endpoint /status
- [ ] `GET /status` — Retorna estado atual do serviço
  - Rotas registradas e seus targets (porta local)
  - Status da conexão WebSocket do CLI (connected/disconnected)
  - Uptime do servidor
  - Quantidade de webhooks na fila (pendentes)
  - Quantidade total de webhooks processados desde o boot

## Melhorias de Persistência
- [ ] SQLite para persistir webhooks em disco (sobrevive a restart)
- [ ] Volume persistente no fly.io para SQLite

## Multi-Developer Support
- [ ] Suporte a múltiplos CLIs conectados simultaneamente
- [ ] Isolamento de rotas por token/usuário
- [ ] Rate limiting por conexão

## Segurança

### Vulnerabilidades Identificadas (CodeQL)

#### 🔴 Alta — `io.ReadAll` sem limite de tamanho (`proxy/handler.go`)
- [ ] Limitar leitura do body com `http.MaxBytesReader` ou `io.LimitReader`
- [ ] Retornar `413 Request Entity Too Large` quando o limite for excedido
- **CWE-400** — Uncontrolled Resource Consumption
- **Risco:** DoS via payload arbitrariamente grande em endpoint público (`/`)
- **Referência:** `internal/proxy/handler.go:61`

#### 🟡 Média — Comparação de token vulnerável a timing attack (`auth/middleware.go`)
- [ ] Substituir comparação direta (`!=`) por `crypto/subtle.ConstantTimeCompare`
- **CWE-208** — Observable Timing Discrepancy
- **Risco:** Side-channel permite adivinhar o token byte-a-byte
- **Referência:** `internal/auth/middleware.go:30`

#### 🟡 Média — WebSocket `CheckOrigin` permissivo (`server/handlers.go`)
- [ ] Implementar validação de origem no `websocket.Upgrader`
- **CWE-346** — Origin Validation Error
- **Risco:** Cross-Site WebSocket Hijacking (CSWSH)
- **Mitigação parcial:** endpoint `/ws` protegido por `TokenMiddleware`
- **Referência:** `internal/server/handlers.go:29`

#### 🟠 Baixa — Erro silenciado no `io.ReadAll` (`proxy/handler.go`)
- [ ] Tratar erro retornado por `io.ReadAll` em vez de descartá-lo com `_`
- **Risco:** Falhas de leitura passam despercebidas, body pode estar incompleto
- **Referência:** `internal/proxy/handler.go:61`

#### 🟢 Baixa — SSRF potencial no Forwarder (`forward/forwarder.go`)
- [ ] Validar/sanitizar `frame.Path` antes de construir a URL de destino
- **CWE-918** — Server-Side Request Forgery
- **Mitigação parcial:** paths validados pelo `RouteRegistry` (whitelist)
- **Referência:** `internal/forward/forwarder.go:46`

### Melhorias Futuras
- [ ] Rotação de tokens
- [ ] Whitelist de IPs permitidos para registro de rotas
- [ ] mTLS entre CLI e servidor

## DX (Developer Experience)
- [ ] `pombohook replay <webhook-id>` — Re-envia um webhook do log
- [ ] `pombohook inspect` — Mostra requests/responses em tempo real no terminal (como tcpdump)
- [ ] Notificação desktop quando webhook é recebido
- [ ] Arquivo de configuração `.pombohook.yml` para evitar flags repetitivas
