#!/usr/bin/env bash

echo "Building React application."
npm run build

echo "check if the manifest file exists, if not, rename it."

if ! test -f public/build/manifest.json; then
  echo "laravel-vite-plugin not running in dev mode, use build manifest file "
  cp public/build/.vite/manifest.json public/build/manifest.json

  echo "manifest.json file was copy to public/build "
fi

echo "Building Golang application."
go build -tags prod -o bin/acme
