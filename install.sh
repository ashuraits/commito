#!/bin/sh
set -e

REPO="ashuraits/commito"
BIN="commito"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64)  arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch"; exit 1 ;;
esac

latest=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
if [ -z "$latest" ]; then
  echo "Could not fetch latest release"; exit 1
fi

url="https://github.com/${REPO}/releases/download/${latest}/${BIN}-${os}-${arch}"
echo "Installing commito ${latest} (${os}/${arch})..."

tmp=$(mktemp)
curl -fsSL "$url" -o "$tmp"
chmod +x "$tmp"

# prefer ~/.local/bin (no sudo needed), fall back to /usr/local/bin
if [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
  INSTALL_DIR="$HOME/.local/bin"
elif [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="/usr/local/bin"
  sudo mv "$tmp" "${INSTALL_DIR}/${BIN}"
  echo "Installed to ${INSTALL_DIR}/${BIN}"
  exit 0
fi

mv "$tmp" "${INSTALL_DIR}/${BIN}"
echo "Installed to ${INSTALL_DIR}/${BIN}"

# remind if ~/.local/bin is not in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "Add to your shell: export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
esac
