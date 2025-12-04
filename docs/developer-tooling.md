# Developer Tooling

- **Skaffold**: `skaffold dev` to build/push local images and deploy Helm chart with dev values.
- **Telepresence**: `telepresence connect` then `telepresence intercept offgridflow-api --port 8080:8080` to route traffic to local API.
- **Local Kubernetes**: use Kind/Minikube, then `helm upgrade --install offgridflow infra/helm/offgridflow -f infra/helm/offgridflow/values-dev.yaml`.
- **Pre-commit**: install hooks `pre-commit install`; runs yaml format, golangci-lint, checkov.
- **Dependency updates**: run `npm audit fix && go get -u ./... && go mod tidy`; consider enabling Renovate/Dependabot.
- **Devcontainer**: open repo in VS Code with `devcontainer` to get Go/Node/Docker toolchain.
