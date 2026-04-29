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

## Melhorias de Segurança
- [ ] Rotação de tokens
- [ ] Whitelist de IPs permitidos para registro de rotas
- [ ] mTLS entre CLI e servidor

## DX (Developer Experience)
- [ ] `pombohook replay <webhook-id>` — Re-envia um webhook do log
- [ ] `pombohook inspect` — Mostra requests/responses em tempo real no terminal (como tcpdump)
- [ ] Notificação desktop quando webhook é recebido
- [ ] Arquivo de configuração `.pombohook.yml` para evitar flags repetitivas
