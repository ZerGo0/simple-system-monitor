#!/bin/bash
set -e

echo "Claude Code GitHub Bot Setup Script"
echo "===================================="
echo ""

GREEN='\x1b[0;32m'
YELLOW='\x1b[1;33m'
RED='\x1b[0;31m'
NC='\x1b[0m'

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

install_tool() {
    local tool_name="$1"
    local check_command="$2"
    local install_commands="$3"

    if command_exists "$check_command"; then
        echo -e "${GREEN}[OK]${NC} $tool_name already installed: $($check_command --version 2>&1 | head -n1)"
    else
        echo -e "${YELLOW}[INSTALL]${NC} Installing $tool_name..."
        eval "$install_commands"
        if command_exists "$check_command"; then
            echo -e "${GREEN}[OK]${NC} $tool_name installed successfully"
        else
            echo -e "${RED}[FAIL]${NC} Failed to install $tool_name"
            exit 1
        fi
    fi
}

version_ge() {
    [ "$(printf '%s\n' "$1" "$2" | sort -V | head -n1)" = "$2" ]
}

install_go() {
    local required_version="$1"
    local current_version=""

    if command_exists "go"; then
        current_version="$(go version | awk '{print $3}' | sed 's/^go//')"
    fi

    if [ -z "$required_version" ]; then
        required_version="1.21.0"
    fi

    if [ -n "$current_version" ] && version_ge "$current_version" "$required_version"; then
        echo -e "${GREEN}[OK]${NC} Go already installed: $(go version)"
        return
    fi

    echo -e "${YELLOW}[INSTALL]${NC} Installing Go ${required_version}..."

    local os
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    if [ "$os" != "linux" ]; then
        echo -e "${RED}[FAIL]${NC} Unsupported OS: $os"
        exit 1
    fi

    local arch
    arch="$(uname -m)"
    local go_arch=""
    if [ "$arch" = "x86_64" ] || [ "$arch" = "amd64" ]; then
        go_arch="amd64"
    elif [ "$arch" = "aarch64" ] || [ "$arch" = "arm64" ]; then
        go_arch="arm64"
    else
        echo -e "${RED}[FAIL]${NC} Unsupported architecture: $arch"
        exit 1
    fi

    local tarball="go${required_version}.${os}-${go_arch}.tar.gz"
    local url="https://go.dev/dl/${tarball}"

    curl -fsSL -o "/tmp/${tarball}" "$url"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "/tmp/${tarball}"
    rm -f "/tmp/${tarball}"

    if [ -n "${GITHUB_PATH:-}" ]; then
        echo "/usr/local/go/bin" >> "$GITHUB_PATH"
    fi

    export PATH="/usr/local/go/bin:$PATH"

    if command_exists "go"; then
        echo -e "${GREEN}[OK]${NC} Go installed successfully: $(go version)"
    else
        echo -e "${RED}[FAIL]${NC} Failed to install Go"
        exit 1
    fi
}

echo "Installing core dependencies..."
echo ""

echo "Checking git..."
if command_exists "git"; then
    echo -e "${GREEN}[OK]${NC} git already installed: $(git --version)"
else
    echo -e "${YELLOW}[INSTALL]${NC} Installing git..."
    sudo apt-get update
    sudo apt-get install -y git
fi

echo "Checking ripgrep (rg)..."
install_tool "Ripgrep" "rg" "sudo apt-get update && sudo apt-get install -y ripgrep"

install_tool "Curl" "curl" "sudo apt-get update && sudo apt-get install -y curl"

echo "Checking Go..."
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
GO_MOD_VERSION=""
if [ -f "$SCRIPT_DIR/go.mod" ]; then
    GO_MOD_VERSION="$(awk '/^go /{print $2; exit}' "$SCRIPT_DIR/go.mod")"
fi
install_go "$GO_MOD_VERSION"

echo ""
echo "Setup completed successfully."
echo ""

echo "Verifying installation..."
for cmd in git rg curl go; do
    if command_exists "$cmd"; then
        echo -e "${GREEN}[OK]${NC} $cmd is available"
    else
        echo -e "${RED}[FAIL]${NC} $cmd is missing"
        exit 1
    fi
done

echo ""
echo "Environment is ready for Claude Code."
