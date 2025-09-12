#!/bin/sh

# This script downloads and installs the latest version of the 'span' tool.
# It is designed to be run via curl:
#   curl -sSfL https://raw.githubusercontent.com/gregory-chatelier/span/main/install.sh | sh

set -e

# --- Configuration ---
# The GitHub repository to fetch the tool from.
REPO="gregory-chatelier/span"

# The name of the binary.
APP_NAME="span"

# The directory to install the binary to.
INSTALL_DIR="/usr/local/bin"

# --- Helper Functions ---

echo_err() {
    echo "Error: $1" >&2
    exit 1
}

# --- Main Logic ---

# 1. Get the latest version from GitHub API
get_latest_version() {
    # Fetches the latest tag name (e.g., "v0.1.0") from the GitHub API.
    # We use curl to fetch the releases and jq to parse the JSON.
    # If jq is not available, we fall back to a grep/sed method.
    if command -v jq >/dev/null 2>&1; then
        curl -s "https://api.github.com/repos/$REPO/releases/latest" | jq -r .tag_name
    else
        curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    fi
}

# 2. Detect OS and Architecture
get_os_arch() {
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$os" in
        linux) os="linux" ;;
        darwin) os="darwin" ;;
        mingw* | msys*) os="windows" ;;
        *) echo_err "Unsupported OS: $os" ;;
    esac

    case "$arch" in
        x86_64 | amd64) arch="amd64" ;;
        aarch64 | arm64) arch="arm64" ;;
        *) echo_err "Unsupported architecture: $arch" ;;
    esac

    echo "$os-$arch"
}

# --- Execution ---

echo "Installing $APP_NAME..."

# Get the latest version and target platform
VERSION=$(get_latest_version)
if [ -z "$VERSION" ]; then
    echo_err "Could not determine the latest version. Check the repository URL."
fi

PLATFORM=$(get_os_arch)

# Construct the download URL
FILENAME="$APP_NAME-$PLATFORM"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"

# For Windows, the binary has a .exe extension
if [ "$(uname -s | cut -c 1-5)" = "MINGW" ] || [ "$(uname -s | cut -c 1-4)" = "MSYS" ]; then
    FILENAME+=".exe"
    DOWNLOAD_URL+=".exe"
fi

# Download the binary to a temporary location
TMP_FILE=$(mktemp)
echo "Downloading from $DOWNLOAD_URL..."
curl -sSfL "$DOWNLOAD_URL" -o "$TMP_FILE"

# Install the binary
echo "Installing to $INSTALL_DIR..."

# Use sudo if the install directory is not writable by the current user
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "$INSTALL_DIR/$APP_NAME"
    chmod +x "$INSTALL_DIR/$APP_NAME"
else
    echo "Sudo privileges are required to install to $INSTALLDIR."
    sudo mv "$TMP_FILE" "$INSTALL_DIR/$APP_NAME"
    sudo chmod +x "$INSTALL_DIR/$APP_NAME"
fi

echo "$APP_NAME version $VERSION has been installed successfully!"
