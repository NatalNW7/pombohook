#!/usr/bin/env bash
set -e

echo "🕊️  Installing PomboHook CLI..."

# Detect OS and Arch
OS="$(uname -s)"
ARCH="$(uname -m)"

# Map OS
case "$OS" in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  *) echo "❌ Unsupported OS: $OS"; exit 1 ;;
esac

# Map Arch
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version from GitHub API
echo "🔍 Fetching latest release version..."
LATEST_TAG=$(curl -s https://api.github.com/repos/NatalNW7/pombohook/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
  echo "❌ Failed to fetch the latest release version."
  exit 1
fi

# GoReleaser uses version without 'v' prefix
VERSION=${LATEST_TAG#v}

TARBALL="pombo_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/NatalNW7/pombohook/releases/latest/download/$TARBALL"

echo "⬇️  Downloading $TARBALL..."

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

if ! curl -sSL -f -o "$TMP_DIR/$TARBALL" "$DOWNLOAD_URL"; then
  echo "❌ Failed to download the binary. Please check your internet connection and verify the release exists."
  exit 1
fi

echo "📦 Extracting..."
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

INSTALL_DIR="/usr/local/bin"
echo "🔧 Installing pombo to $INSTALL_DIR (requires sudo)..."
sudo mv "$TMP_DIR/pombo" "$INSTALL_DIR/pombo"
sudo chmod +x "$INSTALL_DIR/pombo"

echo ""
echo "🎉 PomboHook CLI installed successfully!"
echo "You can now run 'pombo' from anywhere in your terminal."
echo ""
echo "🚀 Getting started:"
echo "  1. Authenticate:  pombo ping --server \"ws://your-server:8080\" --token \"your-token\""
echo "  2. Setup route:   pombo route --path=\"/webhook\" --port=3000"
echo "  3. Start proxy:   pombo go"
echo ""
echo "📂 Local configurations are saved in: ~/.pombohook"
echo ""
