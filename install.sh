#!/bin/sh
set -e

REPO="ashuraits/commito"
BIN="commito"
INSTALL_DIR="/usr/local/bin"

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

if [ -w "$INSTALL_DIR" ]; then
  mv "$tmp" "${INSTALL_DIR}/${BIN}"
else
  sudo mv "$tmp" "${INSTALL_DIR}/${BIN}"
fi

echo "Installed to ${INSTALL_DIR}/${BIN}"
echo "Run: commito"
