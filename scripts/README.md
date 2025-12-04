# Scripts Directory

Organized scripts for development, deployment, testing, and maintenance.

## Directory Structure

### 🚀 Deployment (`deployment/`)
Production and staging deployment scripts:
- `deploy-complete.ps1` - Complete deployment (PowerShell)
- `deploy-staging.ps1` - Staging deployment (PowerShell)
- `deploy-staging.sh` - Staging deployment (Bash)
- `deployment-checklist.ps1` - Deployment checklist automation

### 💻 Development (`development/`)
Local development environment setup:
- `dev-start.ps1` - Start development environment (PowerShell)
- `dev-start.sh` - Start development environment (Bash)
- `dev_up.ps1` - Bring up development stack

### 🧪 Testing (`testing/`)
Test execution and validation scripts:
- `test-all.ps1` - Run all tests
- `test-integration.ps1` - Integration tests (PowerShell)
- `test-integration.sh` - Integration tests (Bash)
- `v1_check.ps1` - V1 verification checks
- `verify_connectors.sh` - Connector verification

### 🔄 Migration (`migration/`)
Database and data migration scripts:
- `migrate.ps1` - Database migrations (PowerShell)
- `migrate.sh` - Database migrations (Bash)
- `seed_demo_data.go` - Demo data seeding

### 🐍 Python Utilities (`python-utils/`)
Python utility scripts for data processing:
- `fix_emissions_complete.py` - Emissions data fix (complete)
- `fix_emissions_handler_final.py` - Emissions handler fixes
- `fix_handlers_robust.py` - Robust handler fixes
- `integrate_scope1_scope3_compliance.py` - Scope 1/3 compliance integration
- `integrate_scope1_scope3_emissions.py` - Scope 1/3 emissions integration

### 📦 Dependencies (`deps/`)
Dependency management scripts (if present)

## Usage

### Quick Start Development
```bash
# Bash
./development/dev-start.sh

# PowerShell
.\development\dev-start.ps1
```

### Run Tests
```bash
# All tests
.\testing\test-all.ps1

# Integration tests only
.\testing\test-integration.sh
```

### Deploy to Staging
```bash
.\deployment\deploy-staging.sh
```

### Run Migrations
```bash
.\migration\migrate.sh
```

## Platform Notes

- `.ps1` files are for PowerShell (Windows/Cross-platform)
- `.sh` files are for Bash (Linux/macOS/WSL)
- Python scripts require Python 3.8+
- Go scripts require Go 1.21+
