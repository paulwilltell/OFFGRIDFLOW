#!/usr/bin/env bash
set -euo pipefail

echo "Updating Go modules..."
go get -u ./...
go mod tidy

echo "Updating npm deps..."
pushd web >/dev/null
npm install
npm audit fix || true
popd >/dev/null

echo "Done."
