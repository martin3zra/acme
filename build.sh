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

# Go app build
echo "Building Golang application..."
GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" \
go build -tags prod -o "bin/acme-${TARGET_OS}-${TARGET_ARCH}${EXT}"

echo "=============================="
echo " Build complete!"
echo " Output: bin/acme-${TARGET_OS}-${TARGET_ARCH}${EXT}"
echo "=============================="
