#!/usr/bin/env bash
set -euo pipefail

role_raw="${OFFGRIDFLOW_SERVICE_ROLE:-${RAILWAY_SERVICE_NAME:-${SERVICE_NAME:-}}}"
role="$(echo "${role_raw}" | tr '[:upper:]' '[:lower:]')"

if [[ -z "${role}" ]]; then
  echo "OFFGRIDFLOW_SERVICE_ROLE not set. Set to 'web' or 'api' in Railway."
  exit 1
fi

if [[ "${role}" == *"web"* || "${role}" == *"frontend"* ]]; then
  echo "Railway start: web service"
  cd web
  exec npm start
fi

if [[ "${role}" == *"api"* || "${role}" == *"backend"* ]]; then
  echo "Railway start: api service"
  if [[ ! -x bin/offgridflow-api ]]; then
    echo "Missing bin/offgridflow-api. Ensure the build step completed."
    exit 1
  fi
  exec ./bin/offgridflow-api
fi

echo "Unknown OFFGRIDFLOW_SERVICE_ROLE: ${role_raw}. Use 'web' or 'api'."
exit 1
