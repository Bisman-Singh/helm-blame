#!/usr/bin/env bash
set -euo pipefail

PROJECT_NAME="helm-blame"
REPO="Bisman-Singh/helm-blame"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Get latest release tag
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
  echo "Error: could not determine latest release version" >&2
  exit 1
fi

# Build download URL
EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${PROJECT_NAME}_${VERSION}_${OS}_${ARCH}.${EXT}"

echo "Installing ${PROJECT_NAME} v${VERSION} (${OS}/${ARCH})..."

# Download and extract
INSTALL_DIR="${HELM_PLUGIN_DIR}/bin"
mkdir -p "$INSTALL_DIR"

cd "$INSTALL_DIR"
if [ "$EXT" = "zip" ]; then
  curl -sL "$URL" -o archive.zip
  unzip -o archive.zip
  rm -f archive.zip
else
  curl -sL "$URL" | tar xz
fi

echo "${PROJECT_NAME} v${VERSION} installed."
