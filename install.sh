#!/usr/bin/env bash
set -e

# MarinateMD installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/glueckkanja/marinatemd/main/install.sh | bash

REPO="glueckkanja/marinatemd"
BINARY_NAME="marinate"
PROJECT_NAME="marinatemd"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get latest release version
echo "Fetching latest release..."
LATEST_VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
SEMVER_VERSION=${LATEST_VERSION#v}

if [ -z "$LATEST_VERSION" ]; then
    echo "Failed to fetch latest version"
    exit 1
fi

echo "Latest version: $LATEST_VERSION"

# Construct download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/${PROJECT_NAME}_${OS}_${ARCH}.tar.gz"

echo "Downloading from: $DOWNLOAD_URL"

# Download and extract
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

curl -fsSL "$DOWNLOAD_URL" -o "${BINARY_NAME}.tar.gz"
tar -xzf "${BINARY_NAME}.tar.gz"

# Install binary
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
echo "Installing to $INSTALL_DIR..."

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
else
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Cleanup
cd -
rm -rf "$TMP_DIR"

echo "✓ $BINARY_NAME $LATEST_VERSION installed successfully!"
echo "Run '$BINARY_NAME version' to verify installation"
