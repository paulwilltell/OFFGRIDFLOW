# Architecture

- Go services (API, worker, CLI) with modular internal packages.
- PostgreSQL primary datastore; event bus abstraction for async flows.
- Next.js frontend served separately or via edge/CDN.
- Infrastructure via Terraform + Kubernetes manifests for api/worker/web.
- Feature-flag-friendly compliance modules (CSRD, CBAM, California, optional SEC).
