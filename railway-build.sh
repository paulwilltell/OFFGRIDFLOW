#!/bin/sh
set -eu

role_raw=${OFFGRIDFLOW_SERVICE_ROLE:-${RAILWAY_SERVICE_NAME:-${SERVICE_NAME:-}}}
role=$(printf '%s' "$role_raw" | tr '[:upper:]' '[:lower:]')

if [ -z "$role" ]; then
  echo "OFFGRIDFLOW_SERVICE_ROLE not set. Set to 'web' or 'api' in Railway."
  exit 1
fi

case "$role" in
  *web*|*frontend*)
    echo "Railway build: web service"
    cd web
    npm ci
    npm run build
    ;;
  *api*|*backend*)
    echo "Railway build: api service"
    go mod download
    mkdir -p bin
    go build -o bin/offgridflow-api ./cmd/api
    ;;
  *)
    echo "Unknown OFFGRIDFLOW_SERVICE_ROLE: ${role_raw}. Use 'web' or 'api'."
    exit 1
    ;;
esac
