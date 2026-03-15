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

# Resolve git tag (prefer exact tag on HEAD, then latest tag, else commit SHA)
GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -z "$GIT_TAG" ]; then
  # Fall back to short commit SHA when no suitable tag is found
  GIT_TAG=$(git rev-parse --short HEAD 2>/dev/null || echo "untagged")
fi
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

# Optional overrides:
#  - BUILD_TAG_OVERRIDE=v1.4.0 ./build.sh darwin amd64
#  - BUILD_TAG_BUMP_NEXT_PATCH=1 ./build.sh linux arm64
if [ -n "$BUILD_TAG_OVERRIDE" ]; then
  NAME_TAG="$BUILD_TAG_OVERRIDE"
elif [ -n "$BUILD_TAG_BUMP_NEXT_PATCH" ]; then
  NAME_TAG="$(bump_patch "$LAST_TAG")"
else
  # Default: use the resolved git tag or commit SHA for artifact naming
  NAME_TAG="$GIT_TAG"
fi
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
echo " Git ref/tag used for build: $GIT_TAG"
echo " Artifact name tag: $NAME_TAG"
echo " Output: $APP_BIN"
echo " Output: $CLI_BIN"
echo " Output: $ZIP_FILE"
echo " Last semantic git tag: $LAST_TAG"
echo " (Note: Name tag is for filenames only; no git tag is created by this script)"
echo "=============================="
