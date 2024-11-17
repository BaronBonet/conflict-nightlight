#!/usr/bin/env bash
set -euo pipefail

DIR_PATH=$(cd "$(dirname "${BASH_SOURCE:-$0}")" && pwd)

. "${DIR_PATH}/../../build_dependencies_versions"

PLATFORM=$(uname)
ARCH=$(uname -m)

DEST_DIR="${DIR_PATH}/../../.local/bin"

MOCKERY_URL="https://github.com/vektra/mockery/releases/download/v${MOCKERY_VERSION}/mockery_${MOCKERY_VERSION}_${PLATFORM}_${ARCH}.tar.gz"
echo "Downloading mockery $MOCKERY_VERSION from $MOCKERY_URL"
echo curl -sf -o "${DEST_DIR}/mockery.tar.gz" -L "$MOCKERY_URL"
if curl -sf -o "${DEST_DIR}/mockery.tar.gz" -L "$MOCKERY_URL"; then
  echo "Inspecting downloaded mockery file..."
  if file "${DEST_DIR}/mockery.tar.gz" | grep -q "gzip compressed data"; then
    echo "Extracting mockery to $DEST_DIR"
    tar -xf "${DEST_DIR}/mockery.tar.gz" -C "${DEST_DIR}/" mockery || {
      echo "Error extracting file"
      exit 2
    }
    echo "Cleaning up mockery archive"
    rm "${DEST_DIR}/mockery.tar.gz"
  else
    echo "Downloaded file is not a valid gzip archive. Contents:"
    cat "${DEST_DIR}/mockery.tar.gz"
    exit 1
  fi
else
  echo "Error downloading mockery from $MOCKERY_URL"
  exit 1
fi

echo "Downloading stringer"
GOBIN=${DIR_PATH}/../../.local/bin/ go install "golang.org/x/tools/cmd/stringer@v$STRINGER_VERSION"
