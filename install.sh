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

# --- Helper Functions ---

echo_err() {
    echo "Error: $1" >&2
    exit 1
}

# 1. Get the latest version from GitHub API
get_latest_version() {
    # Fetches the latest tag name (e.g., "v0.1.0") from the GitHub API.
    # We use curl to fetch the releases and jq to parse the JSON.
    # If jq is not available, we fall back to a grep/sed method.
    if command -v jq >/dev/null 2>&1;
    then
        curl -s "https://api.github.com/repos/$REPO/releases/latest" | jq -r .tag_name
    else
        curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    fi
}

# 2. Detect OS, Architecture, and determine install directory
get_os_arch_install_dir() {
    os_name=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch_name=$(uname -m)
    install_dir="/usr/local/bin" # Default for Linux/macOS
    is_windows=false

    # Detect OS
    case "$os_name" in
        linux) os_name="linux" ;; 
        darwin) os_name="darwin" ;; 
        # Handle Git Bash, MSYS, MINGW, and potentially native CMD/PowerShell
        mingw* | msys* | cygwin* | nt) 
            os_name="windows"
            is_windows=true
            install_dir="$HOME/bin" # Default to user's home bin for Windows
            ;;
        *) echo_err "Unsupported OS: $os_name" ;; 
    esac

    # Detect Architecture
    case "$arch_name" in
        x86_64 | amd64) arch_name="amd64" ;; 
        aarch64 | arm64) arch_name="arm64" ;; 
        *) echo_err "Unsupported architecture: $arch_name" ;; 
    esac

    echo "$os_name-$arch_name $install_dir $is_windows"
}

# --- Execution ---

echo "Installing $APP_NAME..."

# Detect platform and install directory
eval $(get_os_arch_install_dir | { read PLATFORM_VAR INSTALL_DIR_VAR IS_WINDOWS_ENV_VAR; echo "PLATFORM=\"$PLATFORM_VAR\" INSTALL_DIR=\"$INSTALL_DIR_VAR\" IS_WINDOWS_ENV=\"$IS_WINDOWS_ENV_VAR\""; })

# Check if running as root and if sudo is needed
# This needs to be done after INSTALL_DIR is determined
if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ "$(id -u)" -ne 0 ]; then
    echo "Attempting to install to a system directory ($INSTALL_DIR). Re-executing with sudo..."
    # Re-execute the current script with sudo
    # This handles the `curl | sh` case where sudo only applies to curl
    exec sudo sh -c "$(printf %q "$0")" "$@"
fi

# Get the latest version
VERSION=$(get_latest_version)
if [ -z "$VERSION" ]; then
    echo_err "Could not determine the latest version. Check the repository URL."
fi

# Construct the download URL
FILENAME="$APP_NAME-$PLATFORM"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"

# For Windows, the binary has a .exe extension
if [ "$IS_WINDOWS_ENV" = "true" ]; then
    FILENAME+=".exe"
    DOWNLOAD_URL+=".exe"
fi

# Download the binary to a temporary location
TMP_FILE=$(mktemp)
echo "Downloading from $DOWNLOAD_URL..."
curl -sSfL "$DOWNLOAD_URL" -o "$TMP_FILE"

# Install the binary
# Attempt to create the install directory if it doesn't exist
mkdir -p "$INSTALL_DIR" || echo_err "Failed to create installation directory: $INSTALL_DIR. Check permissions."

# Move the binary and make it executable
if mv "$TMP_FILE" "$INSTALL_DIR/$APP_NAME"; then
    chmod +x "$INSTALL_DIR/$APP_NAME"
    echo "$APP_NAME version $VERSION has been installed successfully to $INSTALL_DIR!"
    if [ "$IS_WINDOWS_ENV" = "true" ]; then
        echo "Please ensure $INSTALL_DIR is in your system\'s PATH."
        echo "You may need to restart your terminal or system for changes to take effect."
    fi
else
    echo_err "Failed to move $APP_NAME to $INSTALL_DIR. Check permissions."
fi
