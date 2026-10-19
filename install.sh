#!/usr/bin/env bash
# Kernel Unix One-Line Installer (Linux & macOS)
set -e

echo -e "\033[1;36m\n⚡ Installing Kernel AI Agent...\033[0m"

REPO="404kernel/kernel"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY_NAME="kernel-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}"
INSTALL_DIR="${HOME}/.local/bin"

mkdir -p "${INSTALL_DIR}"
DEST_PATH="${INSTALL_DIR}/kernel"

echo -e "\033[1;33m• Downloading latest binary (${OS}/${ARCH})...\033[0m"
if command -v curl >/dev/null 2>&1; then
  curl -sSL "${DOWNLOAD_URL}" -o "${DEST_PATH}" || {
    echo "Prebuilt binary download failed, checking Go..."
    if command -v go >/dev/null 2>&1; then
      go install github.com/404kernel/kernel@latest
      exit 0
    fi
    exit 1
  }
elif command -v wget >/dev/null 2>&1; then
  wget -q "${DOWNLOAD_URL}" -O "${DEST_PATH}"
fi

chmod +x "${DEST_PATH}"

# Add to PATH if needed
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "export PATH=\"\$PATH:${INSTALL_DIR}\"" >> "${HOME}/.bashrc" 2>/dev/null || true
     echo "export PATH=\"\$PATH:${INSTALL_DIR}\"" >> "${HOME}/.zshrc" 2>/dev/null || true ;;
esac

echo -e "\033[1;32m\n🎉 Installation complete! Sparky is ready.\033[0m"
echo -e "\033[1;36mRun 'kernel' in your terminal.\033[0m\n"
