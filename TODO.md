# PomboHook — Future Features

---

## Code Review — Issues Menores (Prioridade Baixa)

_Identificadas na revisão de código em 2026-05-02._

### Cobertura & Testes
- [ ] Melhorar cobertura de `RunSleep` (11.5%) — considerar mock de `os.Process` para testar caminhos de SIGTERM/SIGKILL/timeout
- [ ] Usar `t.Setenv()` em vez de `os.Setenv()` nos testes de `internal/config/config_test.go` para evitar leaks entre testes paralelos

### Limpeza de Código
- [ ] Remover/simplificar `cli.Dispatch()` em `internal/cli/commands.go` — duplica routing que já existe em `cmd/pombo/main.go`
- [ ] Considerar extrair `extractPort()` helper dos testes (`forwarder_test.go` e `e2e_setup_test.go`) para um `testutil` compartilhado

### Otimizações
- [ ] `TunnelManager.Send()` em `internal/tunnel/server.go` — usar `RLock` para o check de `isOnline` e `Lock` apenas para o write (micro-otimização, baixo impacto)

### Divergências do Spec (Intencionais, Documentar)
- [ ] `cmd/server/main.go`: Queue capacity hardcoded como `200` mas o plano original diz `20` — documentar que é intencional para produção
- [ ] `NewServer` aceita `authMiddleware` como parâmetro explícito (melhor DI) — diverge do spec original que mostra `auth` como campo. Manter design atual, atualizar docs
- [ ] Storage tests usam `package storage` (white-box) — considerar adicionar testes black-box em `package storage_test` para validar a superfície pública

### Documentação
- [ ] Criar `README.md` com instruções de uso, setup e exemplos — listado no plano original mas ausente

---

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

## Melhorias de Segurança
- [ ] Rotação de tokens
- [ ] Whitelist de IPs permitidos para registro de rotas
- [ ] mTLS entre CLI e servidor

## DX (Developer Experience)
- [ ] `pombohook replay <webhook-id>` — Re-envia um webhook do log
- [ ] `pombohook inspect` — Mostra requests/responses em tempo real no terminal (como tcpdump)
- [ ] Notificação desktop quando webhook é recebido
- [ ] Arquivo de configuração `.pombohook.yml` para evitar flags repetitivas
