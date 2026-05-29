#!/usr/bin/env bash

# Premium Installer for aiContext
# https://github.com/TheRealShek/aiContext

set -euo pipefail

{

# --- Color Definitions & Icons ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

TICK="${GREEN}✓${NC}"
CROSS="${RED}✗${NC}"
INFO="${BLUE}i${NC}"

# --- Helper Functions ---
log_info() {
    echo -e "${BOLD}${CYAN}==>${NC} $*"
}

log_success() {
    echo -e "${TICK} $*"
}

log_error() {
    echo -e "${CROSS} ${RED}Error:${NC} $*" >&2
}

log_warning() {
    echo -e "${YELLOW}Warning:${NC} $*"
}

# --- System Validation ---
log_info "Detecting system environment..."

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${OS}" in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    *)
        log_error "Unsupported Operating System: ${OS}"
        exit 1
        ;;
esac

case "${ARCH}" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        log_error "Unsupported architecture: ${ARCH}"
        exit 1
        ;;
esac

log_success "Detected ${BOLD}${OS}/${ARCH}${NC}"

# --- Dependencies Check ---
for cmd in curl tar; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "Missing required dependency: ${BOLD}$cmd${NC}"
        exit 1
    fi
done

# --- Version Resolution ---
log_info "Resolving latest version..."
REPO="TheRealShek/aiContext"
LATEST_JSON=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" || true)

if [ -z "$LATEST_JSON" ]; then
    log_error "Could not fetch latest release. Please check your network or try again."
    exit 1
fi

TAG=$(echo "$LATEST_JSON" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)
if [ -z "$TAG" ]; then
    log_error "Could not resolve release tag."
    exit 1
fi

VERSION="${TAG#v}"
log_success "Latest release is ${BOLD}${TAG}${NC}"

# --- Download & Extraction ---
TMP_DIR=$(mktemp -d)
cleanup() {
    [ -n "${TMP_DIR:-}" ] && rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

URL="https://github.com/TheRealShek/aiContext/releases/download/${TAG}/aiContext_${VERSION}_${OS}_${ARCH}.tar.gz"

log_info "Downloading ${BOLD}${URL}${NC}..."
if ! curl -L -f -o "${TMP_DIR}/aiContext.tar.gz" "$URL"; then
    log_error "Failed to download binary from GitHub."
    exit 1
fi

log_info "Extracting archive..."
if ! tar -xzf "${TMP_DIR}/aiContext.tar.gz" -C "${TMP_DIR}"; then
    log_error "Failed to extract archive."
    exit 1
fi

# --- Target Location Determination ---
BINARY=$(find "${TMP_DIR}" -type f -name "aiContext" -print -quit)
if [ -z "$BINARY" ]; then
    log_error "Binary not found in archive."
    exit 1
fi

# Ensure executable permissions before moving
chmod +x "$BINARY"

INSTALL_DIR="/usr/local/bin"
TARGET="${INSTALL_DIR}/aiContext"

log_info "Installing to ${BOLD}${TARGET}${NC}..."

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY" "$TARGET"
else
    log_info "Elevated permissions (sudo) required to install to ${INSTALL_DIR}."
    if command -v sudo >/dev/null 2>&1; then
        sudo mv "$BINARY" "$TARGET"
    else
        log_warning "No sudo access. Installing to ${HOME}/.local/bin instead..."
        INSTALL_DIR="${HOME}/.local/bin"
        TARGET="${INSTALL_DIR}/aiContext"
        mkdir -p "$INSTALL_DIR"
        mv "$BINARY" "$TARGET"

        # Verify if INSTALL_DIR is in PATH
        case ":$PATH:" in
            *:"$INSTALL_DIR":*) ;;
            *)
                log_warning "${INSTALL_DIR} is not in your PATH. You may need to add it to your shell config."
                ;;
        esac
    fi
fi
log_success "Successfully installed ${BOLD}aiContext${NC} to ${BOLD}${TARGET}${NC}!"

# --- Setup Recommendation ---
echo -e "\n${BOLD}${GREEN}aiContext is ready!${NC}"
echo -e "Run the setup command once to install the default templates:"
echo -e "  ${CYAN}aiContext setup${NC}\n"
}
