# OffGridFlow - Project Structure

This document describes the professional organization of the OffGridFlow project.

## Root Directory Overview

```
OffGridFlow/
├── .devcontainer/       # Development container configuration
├── .github/             # GitHub Actions workflows and templates
├── .vscode/             # VS Code workspace settings
├── bin/                 # Compiled binaries (gitignored)
├── build/               # Build artifacts and executables
├── cmd/                 # Go application entry points
├── config/              # Configuration files
├── deployments/         # Deployment configurations (dev, staging, prod)
├── docs/                # 📚 ALL PROJECT DOCUMENTATION
├── evaluation/          # Evaluation and testing resources
├── examples/            # Example code and usage samples
├── frontend/            # Frontend application code
├── infra/               # Infrastructure as Code (Terraform, K8s, Helm)
├── internal/            # Private Go packages
├── memory-bank/         # Project knowledge and context
├── scripts/             # 🔧 Automation scripts (organized by purpose)
├── temp/                # Temporary files (gitignored)
├── web/                 # Web assets and public files
├── docker-compose.yml   # Docker composition for local development
├── Dockerfile           # Container image definition
├── go.mod              # Go module definition
├── go.sum              # Go dependencies checksum
├── Makefile            # Build automation
├── package.json        # Node.js dependencies
├── skaffold.yaml       # Skaffold configuration for K8s dev
└── README.md           # Main project README
```

## Key Directories Explained

### 📚 Documentation (`docs/`)
**All documentation is now centralized here:**

- **architecture/** - System design, patterns, data flows
- **audits/** - Code and production audit reports
- **guides/** - Quick start guides and references (AWS, SAP, etc.)
- **implementation/** - Feature implementation reports and readiness docs
- **phase-reports/** - Project phase completion tracking
- **reports/** - General analysis reports

See [docs/README.md](docs/README.md) for complete documentation index.

### 🔧 Scripts (`scripts/`)
**Organized by purpose:**

- **deployment/** - Production and staging deployment automation
- **development/** - Local development environment setup
- **migration/** - Database migrations and data seeding
- **python-utils/** - Python utility scripts for data processing
- **testing/** - Test execution and verification scripts

See [scripts/README.md](scripts/README.md) for usage details.

### 🏗️ Infrastructure (`infra/`)
Infrastructure as Code and configuration:

- **terraform/** - Cloud infrastructure provisioning
- **k8s/** - Kubernetes manifests
- **helm/** - Helm charts
- **gitops/** - GitOps configurations
- **db/** - Database schemas and migrations
- **grafana/** - Grafana dashboards
- **prometheus.yml** - Prometheus configuration
- **otel-collector-config.yaml** - OpenTelemetry collector

### 🚀 Deployments (`deployments/`)
Environment-specific configurations:

- **dev/** - Development environment
- **staging/** - Staging environment
- **prod/** - Production environment
- **grafana/** - Grafana configs per environment
- **prometheus/** - Prometheus configs per environment

### 💻 Application Code

#### Backend (Go)
- **cmd/** - Application entry points (main packages)
- **internal/** - Private application packages
  - handlers, services, repositories, models, middleware, etc.
- **config/** - Application configuration management

#### Frontend
- **frontend/** - React/Next.js frontend application
- **web/** - Static web assets

### 🛠️ Development Tools

- **.devcontainer/** - VS Code dev container for consistent environment
- **.github/** - CI/CD workflows, issue templates, PR templates
- **.vscode/** - Shared editor settings and extensions
- **examples/** - Code examples and usage demonstrations
- **evaluation/** - Evaluation framework and test data

### 📦 Build & Deploy

- **build/** - Compiled executables (gitignored)
- **bin/** - Binary outputs (gitignored)
- **temp/** - Temporary files during development (gitignored)
- **Makefile** - Build commands and automation
- **docker-compose.yml** - Multi-container local setup
- **Dockerfile** - Production container image
- **skaffold.yaml** - Kubernetes development workflow

## Configuration Files

### Environment
- `.env.example` - Example environment variables (COMMIT THIS)
- `.env.production.template` - Production environment template
- `.env.staging` - Staging environment variables
- `.env` - Local environment (NEVER COMMIT)

### Code Quality
- `.pre-commit-config.yaml` - Pre-commit hooks
- `.hintrc` - Linting hints configuration
- `.markdownlint.json` - Markdown linting rules
- `.gitignore` - Git ignore patterns
- `.dockerignore` - Docker ignore patterns

### Dependencies
- `go.mod` / `go.sum` - Go module dependencies
- `package.json` / `package-lock.json` - Node.js dependencies

## Organization Principles

### ✅ What We Did

1. **Centralized Documentation** - All `.md` reports and guides moved to `docs/` with clear categorization
2. **Organized Scripts** - Scripts categorized by purpose (deployment, testing, development, etc.)
3. **Clean Root** - Only essential config files and directories in root
4. **Clear Separation** - Source code, docs, scripts, infra, and deployments clearly separated
5. **Gitignore Compliance** - Build artifacts and temp files properly ignored
6. **README Files** - Added README.md in docs/ and scripts/ for navigation

### 📋 Best Practices

- **Documentation**: Always update relevant docs when making changes
- **Scripts**: Add new scripts to appropriate subdirectory with clear naming
- **Environment Variables**: Never commit `.env` files with secrets
- **Build Artifacts**: Always in `build/` or `bin/`, never committed
- **Configuration**: Use templates for sensitive configs (`.template` suffix)

## Quick Navigation

- [Main README](README.md) - Start here
- [Documentation Index](docs/README.md) - All project documentation
- [Scripts Guide](scripts/README.md) - Available automation scripts
- [Infrastructure README](infra/README.md) - Infrastructure documentation
- [Deployment Guide](docs/implementation/PRODUCTION_DEPLOYMENT_GUIDE.md)
- [Quick Start](docs/guides/QUICKSTART.md)

## Contributing

When adding new files:
1. Place them in the appropriate directory
2. Update relevant README files
3. Follow naming conventions (kebab-case for files, lowercase for directories)
4. Document in appropriate location

---

**Last Updated**: December 2025
**Organization Status**: ✅ Production-Ready
