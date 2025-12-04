# OffGridFlow Configuration

This directory centralizes configuration for each environment and replaces ad-hoc `.env` usage in production. Application code reads environment variables (see `internal/config`), so the recommended pattern is:

- Store secrets in a managed vault (AWS Secrets Manager/SSM, GCP Secret Manager, Vault).
- Render an environment-specific file (e.g., `config/production.yaml`) at deploy time.
- Inject values into containers as environment variables via your orchestrator (Kubernetes Secrets + ConfigMaps, ECS task env, etc.).

Usage
-----
1) Copy `config.example.yaml` to `config/development.yaml` (local) or `config/production.yaml` (live).
2) Fill values; never commit real secrets.
3) Map the YAML (or your secret store) into environment variables before starting the service.
4) For Kubernetes, prefer Secrets + ConfigMaps and inject as env vars; do **not** bake secrets into images.

Key environment variables
-------------------------
- Server: `OFFGRIDFLOW_HTTP_PORT`, `OFFGRIDFLOW_APP_ENV`
- Database: `OFFGRIDFLOW_DB_DSN`
- Auth: `OFFGRIDFLOW_JWT_SECRET`, `OFFGRIDFLOW_API_KEY`
- OpenAI: `OFFGRIDFLOW_OPENAI_API_KEY`, `OFFGRIDFLOW_OPENAI_MODEL`, `OFFGRIDFLOW_OPENAI_BASE_URL`
- Stripe: `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_*`
- Feature flags: `OFFGRIDFLOW_ENABLE_AUDIT_LOG`, `OFFGRIDFLOW_ENABLE_METRICS`, `OFFGRIDFLOW_ENABLE_GRAPHQL`, `OFFGRIDFLOW_ENABLE_OFFLINE_AI`
- Ingestion connectors: `OFFGRIDFLOW_AWS_*`, `OFFGRIDFLOW_AZURE_*`, `OFFGRIDFLOW_GCP_*`, `OFFGRIDFLOW_SAP_*`, `OFFGRIDFLOW_UTILITY_*`

Keep `.env` files for local convenience only; exclude them from images and production repos. Add new config keys in `internal/config` and mirror them here to keep the contract clear.