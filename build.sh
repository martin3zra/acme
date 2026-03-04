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

# Resolve git tag (prefer exact tag on HEAD, fallback to latest tag)
GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null || echo "untagged")
GIT_TAG_SAFE=$(printf "%s" "$GIT_TAG" | tr '/ ' '--')

# Compute "next" tag for artifact names only (does NOT create a git tag)
LAST_TAG=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n1)
LAST_TAG=${LAST_TAG:-v0.0.0}

bump_patch() {
  local t="${1#v}"
  IFS='.' read -r major minor patch <<< "$t"
  major=${major:-0}
  minor=${minor:-0}
  patch=${patch:-0}
  printf "v%s.%s.%s" "$major" "$minor" "$((patch + 1))"
}

# Optional override:
# BUILD_TAG_OVERRIDE=v1.4.0 ./build.sh darwin amd64
NAME_TAG="${BUILD_TAG_OVERRIDE:-$(bump_patch "$LAST_TAG")}"
NAME_TAG_SAFE=$(printf "%s" "$NAME_TAG" | tr '/ ' '--')

APP_BIN="bin/acme-${TARGET_OS}-${TARGET_ARCH}-${NAME_TAG_SAFE}${EXT}"
CLI_BIN="bin/acme-cli-${TARGET_OS}-${TARGET_ARCH}-${NAME_TAG_SAFE}${EXT}"

# Go app build
echo "Building Golang application..."
go build -tags prod -ldflags="-s -w" -o "$APP_BIN" .
go build -tags prod -ldflags="-s -w" -o "$CLI_BIN" ./cmd/cli

# Zip artifacts using tag as zip name
ZIP_FILE="bin/${NAME_TAG_SAFE}.zip"
rm -f "$ZIP_FILE"

if ! command -v zip >/dev/null 2>&1; then
  echo "❌ 'zip' command not found. Install zip and rerun."
  exit 1
fi

zip -j "$ZIP_FILE" "$APP_BIN" "$CLI_BIN"

echo "=============================="
echo " Build complete!"
echo " Tag: $GIT_TAG"
echo " Output: $APP_BIN"
echo " Output: $CLI_BIN"
echo " Output: $ZIP_FILE"
echo " Last git tag: $LAST_TAG"
echo " Name tag used: $NAME_TAG (filename only; git tag not created)"
echo "=============================="
