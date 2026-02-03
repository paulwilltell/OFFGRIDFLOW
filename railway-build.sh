#!/usr/bin/env bash
set -euo pipefail

role_raw="${OFFGRIDFLOW_SERVICE_ROLE:-${RAILWAY_SERVICE_NAME:-${SERVICE_NAME:-}}}"
role="$(echo "${role_raw}" | tr '[:upper:]' '[:lower:]')"

if [[ -z "${role}" ]]; then
  echo "OFFGRIDFLOW_SERVICE_ROLE not set. Set to 'web' or 'api' in Railway."
  exit 1
fi

if [[ "${role}" == *"web"* || "${role}" == *"frontend"* ]]; then
  echo "Railway build: web service"
  cd web
  npm ci
  npm run build
  exit 0
fi

if [[ "${role}" == *"api"* || "${role}" == *"backend"* ]]; then
  echo "Railway build: api service"
  go mod download
  go build -o bin/offgridflow-api ./cmd/api
  exit 0
fi

echo "Unknown OFFGRIDFLOW_SERVICE_ROLE: ${role_raw}. Use 'web' or 'api'."
exit 1
