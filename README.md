# 🕊️ PomboHook

PomboHook é uma ferramenta open-source leve e rápida, escrita em Go, para receber webhooks da internet diretamente no seu ambiente de desenvolvimento local (localhost). 

## 🤔 Por que "PomboHook"?
Historicamente, pombos-correio eram usados para entregar mensagens importantes de um ponto distante até um destino seguro de forma rápida e confiável. O **PomboHook** atua como o seu pombo-correio digital: ele captura as mensagens (webhooks) na nuvem e as entrega com segurança diretamente na porta do seu servidor local.

## 🎯 Intenção do Projeto
O desenvolvimento de integrações baseadas em webhooks (como Mercado Pago, Stripe, GitHub, etc.) geralmente exige a exposição da máquina local para a internet usando ferramentas pagas ou complexas como o ngrok. O PomboHook nasceu para ser uma alternativa **self-hosted**, minimalista e voltada para a experiência do desenvolvedor (DX). 

Ele é composto por duas partes:
1. **O Servidor:** Fica hospedado na nuvem (ex: fly.io) recebendo os webhooks reais.
2. **O CLI:** Roda na sua máquina local, conectando-se ao servidor via WebSocket e encaminhando os dados para a porta da sua aplicação (ex: `localhost:8080`).

## 🚀 Setup Inicial e Como Executar

### Pré-requisitos
- [Go](https://go.dev/) 1.21+ instalado.
- Make (opcional, mas recomendado).

### Compilando o projeto
Clone o repositório e compile os binários do Servidor e do CLI:
```bash
make build
```
Isso gerará dois executáveis na pasta `bin/`: `bin/pombohook-server` e `bin/pombo`.

### Passo 1: Subir o Servidor
Você pode rodar o servidor localmente para testes ou hospedá-lo na nuvem.
```bash
# O servidor usa variáveis de ambiente para configuração
export PORT=8080
export AUTH_TOKEN="meu-token-super-secreto"
export LOG_LEVEL="debug"

./pombohook-server
```

### Passo 2: Conectar o CLI (Pombo)
Na sua máquina local, autentique-se com o servidor:
```bash
# Ping inicial para salvar a configuração no seu ~/.pombohook
./pombo ping --server "ws://localhost:8080" --token "meu-token-super-secreto"
```

### Passo 3: Registrar uma Rota
Diga ao PomboHook para qual porta local ele deve mandar os webhooks de um determinado path:
```bash
./pombo route --path="/webhooks/pagamentos" --port=3000 # enviara todos webhooks que chegarem no path "/webhooks/pagamentos" do servidor para localhost:3000/webhooks/pagamentos
```

### Passo 4: Voar! (Iniciar o Forwarding)
Inicie a escuta em tempo real:
```bash
./pombo go
```
Se preferir rodar em background, use:
```bash
./pombo go --background
```
Para parar a execução em background:
```bash
./pombo sleep
```

## 📦 Resiliência e Fila de Webhooks (Offline Mode)

O que acontece se a sua internet cair, ou se você fechar o CLI local enquanto a integração (ex: Mercado Pago) tenta te mandar um webhook?

Para evitar perda de dados, o Servidor do PomboHook possui uma **fila em memória (Queue)**:
1. **Desconexão:** Quando o Servidor detecta que o CLI local não está conectado, ele intercepta o webhook recebido e o guarda na fila. O serviço externo que enviou o webhook receberá uma resposta de sucesso (`202 Accepted`), e não precisará fazer retentativas.
2. **Limite de Segurança:** Por padrão, a fila comporta até **20 webhooks simultâneos**. Se o limite for atingido, os webhooks mais antigos são descartados para dar espaço aos novos (comportamento de *buffer circular*). Isso impede vazamentos de memória na sua hospedagem em nuvem.
3. **Reconexão (Flush):** Assim que você ligar o CLI (`./pombo go`) novamente, o servidor detecta a conexão e imediatamente "descarrega" (flush) todos os webhooks acumulados na fila diretamente para a sua máquina local, na ordem em que chegaram.

## 📂 Organização de Pastas e Responsabilidades

O projeto segue a estrutura padrão de projetos Go (`Standard Go Project Layout`):

- `cmd/`
  - `server/main.go`: Ponto de entrada do Servidor. Faz injeção de dependências e sobe o servidor HTTP.
  - `pombo/main.go`: Ponto de entrada do CLI local. Processa os comandos (ping, route, go, sleep).
- `internal/` — Código privado e regras de negócio da aplicação:
  - `auth/`: Middlewares de autenticação (validação do `AUTH_TOKEN`).
  - `cli/`: Lógica principal dos comandos do CLI e gerenciamento de processos (daemon/background).
  - `config/`: Setup de variáveis de ambiente.
  - `forward/`: Forwarder HTTP local. Recebe os frames via WebSocket e dispara os requests para o seu `localhost`.
  - `proxy/`: Proxy reverso do Servidor. Intercepta os webhooks da web e os coloca na fila.
  - `queue/`: Fila em memória para gerenciar bursts de webhooks caso o CLI se desconecte temporariamente.
  - `router/`: Gerenciador de rotas. Mapeia paths (ex: `/webhook`) para portas locais.
  - `server/`: Estrutura base do servidor HTTP e rotas WebSocket.
  - `storage/`: Manipulação de arquivos locais do CLI (`config.json`, `routes.json`, `pombo.pid`).
  - `tunnel/`: Gerenciamento do WebSocket (TunnelManager) entre o Servidor e o CLI.
- `tests/` — Testes E2E (End-to-End) garantindo que todas as peças funcionam juntas.

## 🤝 Contribuindo

Nós adoramos contribuições! Se você deseja ajudar a melhorar o PomboHook, por favor, leia o nosso guia de contribuição antes de começar.

Veja como contribuir em: [CONTRIBUTING.md](CONTRIBUTING.md)
