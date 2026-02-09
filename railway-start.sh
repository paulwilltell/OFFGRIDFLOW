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
    echo "Railway start: web service"
    exec npm run start --workspace web
    ;;
  *api*|*backend*)
    echo "Railway start: api service"
    if [ -x ./bin/offgridflow-api ]; then
      exec ./bin/offgridflow-api
    fi
    if [ -x /app/offgridflow-api ]; then
      exec /app/offgridflow-api
    fi
    echo "Missing offgridflow-api binary. Ensure the build step completed."
    exit 1
    ;;
  *)
    echo "Unknown OFFGRIDFLOW_SERVICE_ROLE: ${role_raw}. Use 'web' or 'api'."
    exit 1
    ;;
esac
