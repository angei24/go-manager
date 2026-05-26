#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== build =="
go build -o gm ./cmd/gm

echo "== unit tests =="
go test ./...

echo "== gm help =="
./gm --help >/dev/null

echo "== init smoke =="
TMP=$(mktemp -d)
./gm init "$TMP/demo" --module example.com/demo --no-git
test -f "$TMP/demo/go.mod"
test -f "$TMP/demo/main.go"
test -f "$TMP/demo/.gm-version"
rm -rf "$TMP"

echo "OK"
