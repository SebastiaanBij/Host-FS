#!/bin/sh

# Configuration
set -eu

# Variables
REPO_URL="https://github.com/SebastiaanBij/Host-FS.git"
REPO_BRANCH="master"
WORK_DIR="/tmp/host-fs"
PLUGIN_NAME="host-fs"
IMAGE_TAG="host-fs:rootfs"
BUILDER_NAME="host-fs-builder"
LOG_FILE="/tmp/host-fs-install.log"

# Functions
log() {
    printf '[host-fs] %s\n' "$1"
}

die() {
    printf '[host-fs] ERROR: %s\n' "$1" >&2
    printf '[host-fs] Output:\n' >&2
    cat "${LOG_FILE}" >&2
    exit 1
}

require() {
    command -v "$1" > /dev/null 2>&1 || die "'$1' is required but not installed"
}

run() {
    "$@" >> "${LOG_FILE}" 2>&1 || die "Command failed: $*"
}

# Validation
touch "${LOG_FILE}"
[ "$(id -u)" = "0" ] || die "This script must be run as root"

require git
require tar
require docker

# Clone Repository
log "Cloning repository..."
rm -rf "${WORK_DIR}"
run git clone -q "${REPO_URL}" -b "${REPO_BRANCH}" "${WORK_DIR}"
cd "${WORK_DIR}"

# Build Root Filesystem
log "Building root filesystem..."
rm -rf ./rootfs
mkdir -p ./rootfs

run docker build \
    --tag "${IMAGE_TAG}" \
    --file ./.docker/Dockerfile \
    .

run docker create --name "${BUILDER_NAME}" "${IMAGE_TAG}"
docker export "${BUILDER_NAME}" | tar -x -C ./rootfs/ >> "${LOG_FILE}" 2>&1 || die "Command failed: docker export"
run docker rm "${BUILDER_NAME}"
run docker rmi "${IMAGE_TAG}"

# Install Plugin
log "Installing plugin..."
docker plugin disable "${PLUGIN_NAME}" >> "${LOG_FILE}" 2>&1 || true
docker plugin rm "${PLUGIN_NAME}" >> "${LOG_FILE}" 2>&1 || true
run docker plugin create "${PLUGIN_NAME}" .
run docker plugin enable "${PLUGIN_NAME}:latest"

# Cleanup
log "Cleaning up..."
cd /
rm -rf "${WORK_DIR}"
rm -f "${LOG_FILE}"

log "Done! Plugin installed successfully."
