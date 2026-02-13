#!/bin/sh
set -e

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

if [ "$OS" = "darwin" ]; then
    INSTALL_DIR="/usr/local/bin"
    NEEDS_SUDO="sudo"
elif [ "$OS" = "linux" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    NEEDS_SUDO=""
    mkdir -p "$INSTALL_DIR"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

echo "Detected OS: $OS, Arch: $ARCH"
echo "Installing to $INSTALL_DIR..."

VERSION=${1:-latest}
if [ "$VERSION" = "latest" ]; then
    LATEST_URL="https://api.github.com/repos/made-with-future/cleat/releases/latest"
    # Simple grep to parse JSON for tag_name
    TAG=$(curl -s $LATEST_URL | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4)
    if [ -z "$TAG" ]; then
        echo "Could not find latest release tag."
        exit 1
    fi
    VERSION=$TAG
fi

echo "Version: $VERSION"

ASSET_NAME="cleat_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/made-with-future/cleat/releases/download/${VERSION}/${ASSET_NAME}"

echo "Downloading $DOWNLOAD_URL..."
curl -fsSL "$DOWNLOAD_URL" -o cleat.tar.gz

echo "Extracting..."
tar -xzf cleat.tar.gz

echo "Installing..."
if [ -n "$NEEDS_SUDO" ] && [ ! -w "$INSTALL_DIR" ]; then
    echo "Sudo permissions required to move binary to $INSTALL_DIR"
    $NEEDS_SUDO mv cleat "$INSTALL_DIR/cleat"
else
    mv cleat "$INSTALL_DIR/cleat"
fi

chmod +x "$INSTALL_DIR/cleat"
rm cleat.tar.gz

echo "Successfully installed cleat to $INSTALL_DIR/cleat"
