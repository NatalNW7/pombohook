# 🕊️ PomboHook

PomboHook is a lightweight, fast, open-source tool written in Go to receive webhooks from the internet directly into your local development environment (localhost).

## 🤔 Why "PomboHook"?
Historically, carrier pigeons (*pombos-correio* in Portuguese) were used to deliver important messages from a distant point to a safe destination quickly and reliably. **PomboHook** acts as your digital carrier pigeon: it catches messages (webhooks) in the cloud and safely delivers them right to the door of your local server.

## 🎯 Project Intent
Developing webhook-based integrations (like Mercado Pago, Stripe, GitHub, etc.) usually requires exposing your local machine to the internet using paid or complex tools like ngrok. PomboHook was born to be a **self-hosted**, minimalist alternative focused on developer experience (DX).

It consists of two parts:
1. **The Server:** Hosted in the cloud (e.g., fly.io), receiving the actual webhooks.
2. **The CLI:** Runs on your local machine, connecting to the server via WebSocket and forwarding the data to your application's port (e.g., `localhost:8080`).

## 🚀 Initial Setup and How to Run

### Prerequisites
- [Go](https://go.dev/) 1.21+ installed.
- Make (optional, but recommended).

### Installing the CLI (Pombo)

There are multiple ways to install the CLI so it's globally available in your system's `PATH`:

#### Option A: Automatic Install Script (Recommended)
You can use our automated installer which downloads the latest release and configures your system's `PATH`.

**For Linux / macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/NatalNW7/pombohook/main/scripts/install.sh | bash
```

**For Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/NatalNW7/pombohook/main/scripts/install.ps1 | iex
```

#### Option B: Using Go (For Go developers)
If you already have Go installed, you can simply run:
```bash
go install github.com/NatalNW7/pombohook/cmd/pombo@latest
```

#### Option C: Compiling manually
```bash
make build
```
This will generate two executables in the `bin/` folder: `bin/pombohook-server` and `bin/pombo`.

### Step 1: Start the Server
You can run the server locally for testing or host it in the cloud.

#### Option A: Running from Binary
```bash
# The server uses environment variables for configuration
export PORT=8080
export AUTH_TOKEN="my-super-secret-token"
export LOG_LEVEL="debug"

./bin/pombohook-server
```

#### Option B: Running with Docker (Recommended for Cloud Hosting)
You can use our official, highly secure, minimalist Docker image based on `scratch` (0 shells, minimal attack surface) published on Docker Hub.

```bash
docker run -d \
  -p 8080:8080 \
  -e PORT=8080 \
  -e AUTH_TOKEN="my-super-secret-token" \
  -e LOG_LEVEL="debug" \
  --name pombohook-server \
  natalnw7/pombohook-server:latest
```

### Step 2: Connect the CLI (Pombo)
On your local machine, authenticate with the server:
```bash
# Initial ping to save the configuration in your ~/.pombohook
./pombo ping --server "ws://localhost:8080" --token "my-super-secret-token"
```

### Step 3: Register a Route
Tell PomboHook to which local port it should send webhooks from a specific path:
```bash
./pombo route --path="/webhooks/payments" --port=3000 # will send all webhooks arriving at the server's "/webhooks/payments" path to localhost:3000/webhooks/payments
```

### Step 4: Fly! (Start Forwarding)
Start listening in real-time:
```bash
./pombo go
```
If you prefer to run it in the background, use:
```bash
./pombo go --background
```
To stop the background execution:
```bash
./pombo sleep
```

## 📦 Resilience and Webhook Queue (Offline Mode)

What happens if your internet goes down, or if you close the local CLI while an integration (e.g., Mercado Pago) tries to send you a webhook?

To prevent data loss, the PomboHook Server has an **in-memory queue**:
1. **Disconnection:** When the Server detects that the local CLI is not connected, it intercepts the incoming webhook and stores it in the queue. The external service that sent the webhook will receive a success response (`202 Accepted`) and will not need to make retries.
2. **Safety Limit:** By default, the queue holds up to **20 simultaneous webhooks**. If the limit is reached, the oldest webhooks are discarded to make room for new ones (*circular buffer* behavior). This prevents memory leaks in your cloud hosting.
3. **Reconnection (Flush):** As soon as you start the CLI (`./pombo go`) again, the server detects the connection and immediately flushes all accumulated webhooks from the queue directly to your local machine, in the order they arrived.

## 📂 Folder Organization and Responsibilities

The project follows the standard Go project structure (`Standard Go Project Layout`):

- `cmd/`
  - `server/main.go`: Server entry point. Performs dependency injection and starts the HTTP server.
  - `pombo/main.go`: Local CLI entry point. Processes commands (ping, route, go, sleep).
- `internal/` — Private code and application business rules:
  - `auth/`: Authentication middlewares (`AUTH_TOKEN` validation).
  - `cli/`: Core logic for CLI commands and process management (daemon/background).
  - `config/`: Environment variables setup.
  - `forward/`: Local HTTP forwarder. Receives frames via WebSocket and fires requests to your `localhost`.
  - `proxy/`: Server reverse proxy. Intercepts webhooks from the web and places them in the queue.
  - `queue/`: In-memory queue to manage webhook bursts in case the CLI temporarily disconnects.
  - `router/`: Route manager. Maps paths (e.g., `/webhook`) to local ports.
  - `server/`: Base structure of the HTTP server and WebSocket routes.
  - `storage/`: Local CLI file manipulation (`config.json`, `routes.json`, `pombo.pid`).
  - `tunnel/`: WebSocket management (TunnelManager) between the Server and the CLI.
- `tests/` — End-to-End (E2E) tests ensuring all pieces work together.

## 🤝 Contributing

We love contributions! If you wish to help improve PomboHook, please read our contributing guide before you start.

See how to contribute in: [CONTRIBUTING.md](CONTRIBUTING.md)
