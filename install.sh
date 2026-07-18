#!/bin/sh

# Installer for aiContext
# https://github.com/TheRealShek/aiContext

set -eu

{

# --- Color Definitions & Icons ---
if [ -t 1 ]; then
    ESC=$(printf '\033')
    RED="${ESC}[0;31m"
    GREEN="${ESC}[0;32m"
    YELLOW="${ESC}[0;33m"
    CYAN="${ESC}[0;36m"
    BOLD="${ESC}[1m"
    NC="${ESC}[0m"
else
    RED=""
    GREEN=""
    YELLOW=""
    CYAN=""
    BOLD=""
    NC=""
fi

TICK="${GREEN}✓${NC}"
CROSS="${RED}✗${NC}"
# --- Helper Functions ---
log_info() {
    printf "${BOLD}${CYAN}==>${NC} %s\n" "$*"
}

log_success() {
    printf "${TICK} %s\n" "$*"
}

log_error() {
    printf "${CROSS} ${RED}Error:${NC} %s\n" "$*" >&2
}

log_warning() {
    printf "${YELLOW}Warning:${NC} %s\n" "$*"
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

calculate_sha256() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    else
        log_error "A SHA-256 utility is required (sha256sum or shasum)."
        exit 1
    fi
}

# --- Version Resolution ---
log_info "Resolving latest version..."
REPO="TheRealShek/aiContext"
LATEST_URL=$(curl -sSfL -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" || true)

if [ -z "$LATEST_URL" ]; then
    log_error "Could not fetch latest release. Please check your network or try again."
    exit 1
fi

TAG=${LATEST_URL##*/}
case "$TAG" in
    v*) ;;
    *)
        log_error "Could not resolve release tag."
        exit 1
        ;;
esac

VERSION="${TAG#v}"
log_success "Latest release is ${BOLD}${TAG}${NC}"

# --- Download & Extraction ---
TMP_DIR=$(mktemp -d)
cleanup() {
    [ -n "${TMP_DIR:-}" ] && rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

ARCHIVE_NAME="aiContext_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/TheRealShek/aiContext/releases/download/${TAG}/${ARCHIVE_NAME}"
CHECKSUM_URL="https://github.com/TheRealShek/aiContext/releases/download/${TAG}/checksums.txt"

log_info "Downloading ${BOLD}${URL}${NC}..."
if ! curl -L -f -o "${TMP_DIR}/${ARCHIVE_NAME}" "$URL"; then
    log_error "Failed to download binary from GitHub."
    exit 1
fi

log_info "Verifying archive checksum..."
if ! curl -L -f -o "${TMP_DIR}/checksums.txt" "$CHECKSUM_URL"; then
    log_error "Failed to download release checksums from GitHub."
    exit 1
fi
EXPECTED_SHA=$(awk -v archive="$ARCHIVE_NAME" '$2 == archive {print $1}' "${TMP_DIR}/checksums.txt")
if [ -z "$EXPECTED_SHA" ]; then
    log_error "Checksum for ${ARCHIVE_NAME} was not found in checksums.txt."
    exit 1
fi
ACTUAL_SHA=$(calculate_sha256 "${TMP_DIR}/${ARCHIVE_NAME}")
if [ "$ACTUAL_SHA" != "$EXPECTED_SHA" ]; then
    log_error "Checksum verification failed for ${ARCHIVE_NAME}."
    exit 1
fi
log_success "Checksum verified"

log_info "Extracting archive..."
if ! tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "${TMP_DIR}"; then
    log_error "Failed to extract archive."
    exit 1
fi

# --- Target Location Determination ---
BINARY="${TMP_DIR}/aiContext"
if [ ! -f "$BINARY" ]; then
    log_error "Binary not found in archive."
    exit 1
fi

# Ensure executable permissions before moving
chmod +x "$BINARY"

INSTALL_DIR="${AICONTEXT_INSTALL_DIR:-/usr/local/bin}"
TARGET="${INSTALL_DIR}/aiContext"

log_info "Installing to ${BOLD}${TARGET}${NC}..."

if [ -d "$INSTALL_DIR" ] && [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY" "$TARGET"
elif [ ! -e "$INSTALL_DIR" ] && mkdir -p "$INSTALL_DIR" 2>/dev/null; then
    mv "$BINARY" "$TARGET"
else
    log_info "Elevated permissions (sudo) required to install to ${INSTALL_DIR}."
    if command -v sudo >/dev/null 2>&1; then
        sudo mkdir -p "$INSTALL_DIR"
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

# --- Run Setup ---
log_info "Running initial setup..."
if "$TARGET" setup; then
    printf "\n${BOLD}${GREEN}Setup complete! aiContext is ready to use.${NC}\n"
    printf "Navigate to your project directory and run:\n"
    printf "  ${CYAN}aiContext init${NC}\n\n"
else
    log_error "Setup failed. You can retry manually with: aiContext setup"
    exit 1
fi
}
