#!/usr/bin/env bash
set -e

# Usage instructions
usage() {
    cat <<EOF
tskr Installer

Usage:
  install.sh [options]

Options:
  -b, --build-from-source   Force building from source using Go CLI
  -r, --release             Download pre-compiled binary release from GitHub
  -d, --dir <path>          Custom installation directory (default: \$HOME/.local/bin)
  -h, --help                Show this help message

Examples:
  ./install.sh
  ./install.sh --build-from-source
  ./install.sh --dir /usr/local/bin
EOF
    exit 0
}

BUILD_FROM_SOURCE=false
FORCE_RELEASE=false
INSTALL_DIR=""
REPO_OWNER="fibrasek"
REPO_NAME="tskr"

# Parse command line options
while [ $# -gt 0 ]; do
    case "$1" in
        -b|--build-from-source)
            BUILD_FROM_SOURCE=true
            shift
            ;;
        -r|--release)
            FORCE_RELEASE=true
            shift
            ;;
        -d|--dir)
            if [ -n "$2" ]; then
                INSTALL_DIR="$2"
                shift 2
            else
                echo "Error: --dir requires a path argument." >&2
                exit 1
            fi
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Error: Unknown option $1" >&2
            usage
            ;;
    esac
done

# Default installation directory if not specified
if [ -z "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
fi

# Operating system detection
OS="$(uname -s)"
case "$OS" in
    Linux*)     OS_NAME="linux" ;;
    Darwin*)    OS_NAME="darwin" ;;
    *)          echo "Error: Unsupported operating system: $OS" >&2; exit 1 ;;
esac

# Architecture detection
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)   ARCH_NAME="amd64" ;;
    aarch64|arm64)  ARCH_NAME="arm64" ;;
    *)              echo "Error: Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

echo "Target System: $OS_NAME/$ARCH_NAME"
echo "Installation Directory: $INSTALL_DIR"

TEMP_DIR=$(mktemp -d)
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Check if Go compiler is available
HAS_GO=false
if command -v go >/dev/null 2>&1; then
    HAS_GO=true
fi

# Determine build or download mode
if [ "$BUILD_FROM_SOURCE" = true ]; then
    if [ "$HAS_GO" = false ]; then
        echo "Error: Go is not installed on this system. Cannot build from source." >&2
        exit 1
    fi
    MODE="source"
elif [ "$FORCE_RELEASE" = true ]; then
    MODE="release"
elif [ -f "go.mod" ] && [ "$HAS_GO" = true ]; then
    MODE="source"
else
    MODE="release"
fi

if [ "$MODE" = "source" ]; then
    echo "Building tskr from source..."
    if [ -f "go.mod" ]; then
        go build -o "$TEMP_DIR/tskr" .
    else
        echo "Fetching repository source via Go..."
        TMP_SRC=$(mktemp -d)
        git clone --depth 1 "https://github.com/$REPO_OWNER/$REPO_NAME.git" "$TMP_SRC"
        (cd "$TMP_SRC" && go build -o "$TEMP_DIR/tskr" .)
        rm -rf "$TMP_SRC"
    fi
else
    echo "Downloading pre-compiled release for $OS_NAME/$ARCH_NAME..."
    RELEASE_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/latest/download/tskr-${OS_NAME}-${ARCH_NAME}.tar.gz"
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$RELEASE_URL" -o "$TEMP_DIR/tskr.tar.gz"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$TEMP_DIR/tskr.tar.gz" "$RELEASE_URL"
    else
        echo "Error: Neither curl nor wget is installed." >&2
        exit 1
    fi

    tar -xzf "$TEMP_DIR/tskr.tar.gz" -C "$TEMP_DIR"
    
    # Locate extracted binary
    if [ -f "$TEMP_DIR/tskr-${OS_NAME}-${ARCH_NAME}" ]; then
        mv "$TEMP_DIR/tskr-${OS_NAME}-${ARCH_NAME}" "$TEMP_DIR/tskr"
    fi
fi

if [ ! -f "$TEMP_DIR/tskr" ]; then
    echo "Error: Binary output not found." >&2
    exit 1
fi

chmod +x "$TEMP_DIR/tskr"

# Execute tskr install subcommand
echo "Installing binary..."
"$TEMP_DIR/tskr" install --dir "$INSTALL_DIR"

# Shell configuration guide
SHELL_NAME="$(basename "${SHELL:-bash}")"
case "$SHELL_NAME" in
    fish)
        PROFILE_FILE="$HOME/.config/fish/config.fish"
        EXPORT_CMD="set -Ua fish_user_paths $INSTALL_DIR"
        ;;
    zsh)
        PROFILE_FILE="$HOME/.zshrc"
        EXPORT_CMD="export PATH=\"$INSTALL_DIR:\$PATH\""
        ;;
    *)
        PROFILE_FILE="$HOME/.bashrc"
        EXPORT_CMD="export PATH=\"$INSTALL_DIR:\$PATH\""
        ;;
esac

case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
        echo ""
        echo "To finish setup, add $INSTALL_DIR to your PATH:"
        echo "  Run: echo '$EXPORT_CMD' >> $PROFILE_FILE"
        echo "  Or reload shell config: source $PROFILE_FILE"
        ;;
esac

echo "Installation complete!"
