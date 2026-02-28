#!/usr/bin/env bash
set -e

# Default values
TARGET_OS=${1:-darwin}      # macOS by default
TARGET_ARCH=${2:-amd64}     # default architecture

echo "=============================="
echo " Building for: $TARGET_OS / $TARGET_ARCH"
echo "=============================="

# Validate supported OS's
case "$TARGET_OS" in
  darwin|linux|windows)
    ;;
  *)
    echo "❌ Unsupported OS: $TARGET_OS"
    echo "Allowed: darwin, linux, windows"
    exit 1
    ;;
esac

# React build
echo "Building React application..."
npm run build

echo "Checking manifest.json..."
if ! test -f public/build/manifest.json; then
  echo "Manifest not found — copying vite build manifest..."
  cp public/build/.vite/manifest.json public/build/manifest.json
  echo "manifest.json copied!"
fi

# For Windows, add `.exe` automatically
EXT=""
[ "$TARGET_OS" = "windows" ] && EXT=".exe"

export GOOS="$TARGET_OS"
export GOARCH="$TARGET_ARCH"
export CGO_ENABLED=0

# Go app build
echo "Building Golang application..."
go build -tags prod -ldflags="-s -w" -o "bin/acme-${TARGET_OS}-${TARGET_ARCH}${EXT}" .
go build -tags prod -ldflags="-s -w" -o "bin/acme-cli-${TARGET_OS}-${TARGET_ARCH}${EXT}" ./cmd/cli

echo "=============================="
echo " Build complete!"
echo " Output: bin/acme-${TARGET_OS}-${TARGET_ARCH}${EXT}"
echo " Output: bin/acme-cli-${TARGET_OS}-${TARGET_ARCH}${EXT}"
echo "=============================="
