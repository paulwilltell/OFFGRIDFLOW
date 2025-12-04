# OffGridFlow - Infrastructure & DevOps Guide

This guide covers everything needed to run OffGridFlow from local development to production deployment.

## Table of Contents

1. [Quick Start - Local Development](#quick-start---local-development)
2. [Docker Images](#docker-images)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Terraform Infrastructure](#terraform-infrastructure)
5. [CI/CD Pipeline](#cicd-pipeline)
6. [Troubleshooting](#troubleshooting)

---

## Quick Start - Local Development

### Prerequisites

- Docker Desktop (20.10+)
- Docker Compose (2.0+)
- 8GB RAM minimum
- 20GB disk space

### One-Command Start

**Linux/macOS:**
```bash
./scripts/dev-start.sh
```

**Windows:**
```powershell
.\scripts\dev-start.ps1
```

**Using Make:**
```bash
make dev
```

**Manual Docker Compose:**
```bash
docker-compose up
```

### What Gets Started

- **PostgreSQL** (port 5432) - Database with automatic migrations
- **Redis** (port 6379) - Cache and rate limiting
- **API** (port 8080) - Go backend with auto-migrations
- **Worker** - Background job processor
- **Web** (port 3000) - Next.js frontend
- **Jaeger** (port 16686) - Distributed tracing UI
- **Prometheus** (port 9090) - Metrics collection
- **Grafana** (port 3001) - Metrics visualization

### Environment Configuration

Copy the template:
```bash
cp .env.production.template .env
```

Required environment variables:
```env
# Database
OFFGRIDFLOW_DB_DSN=postgresql://offgridflow:changeme@postgres:5432/offgridflow?sslmode=disable

# Redis
OFFGRIDFLOW_REDIS_URL=redis://redis:6379/0

# API Keys (optional for local dev)
OFFGRIDFLOW_OPENAI_API_KEY=sk-...
OFFGRIDFLOW_STRIPE_SECRET_KEY=sk_test_...
```

### Accessing Services

- **Web UI**: http://localhost:3000
- **API**: http://localhost:8080
- **API Docs**: http://localhost:8080/swagger
- **Jaeger**: http://localhost:16686
- **Grafana**: http://localhost:3001 (admin/admin)
- **Prometheus**: http://localhost:9090

### Common Commands

```bash
# View logs
docker-compose logs -f api
docker-compose logs -f worker

# Restart a service
docker-compose restart api

# Run tests
make test

# Stop everything
docker-compose down

# Stop and clean volumes
docker-compose down -v

# Rebuild and restart
docker-compose up --build
```

---

## Docker Images

### Build Images Locally

```bash
# Build all images
make docker-build

# Build individual images
docker build -t offgridflow-api:latest .
docker build --target worker -t offgridflow-worker:latest .
docker build -t offgridflow-web:latest ./web
```

### Image Details

#### API Image (`offgridflow-api`)
- **Base**: `golang:1.24-alpine` (builder), `alpine:3.18` (runtime)
- **Size**: ~30MB (optimized multi-stage build)
- **Entry**: `/app/offgridflow-api`
- **Port**: 8080
- **Features**:
  - Static binary (no CGO)
  - Non-root user
  - Health checks
  - Automatic migrations on startup

#### Worker Image (`offgridflow-worker`)
- **Base**: Same as API
- **Size**: ~25MB
- **Entry**: `/app/offgridflow-worker`
- **Features**:
  - Polls queues for jobs
  - Processes emissions calculations
  - Handles connector syncs

#### Web Image (`offgridflow-web`)
- **Base**: `node:20-alpine`
- **Size**: ~150MB (with Next.js standalone)
- **Entry**: `node server.js`
- **Port**: 3000
- **Features**:
  - Standalone Next.js build
  - Optimized for production
  - Static asset caching

### Push to Registry

```bash
# Configure registry
export REGISTRY=ghcr.io/yourusername

# Push images
make docker-push

# Or manually
docker tag offgridflow-api:latest $REGISTRY/offgridflow-api:latest
docker push $REGISTRY/offgridflow-api:latest
```

---

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Secrets configured (see below)

### Quick Deploy

```bash
# Deploy everything
make k8s-deploy

# Or step by step
kubectl apply -f infra/k8s/namespace.yaml
kubectl apply -f infra/k8s/configmap.yaml
kubectl apply -f infra/k8s/secrets.yaml  # Create from secrets.yaml.example
kubectl apply -f infra/k8s/services.yaml
kubectl apply -f infra/k8s/api-deployment.yaml
kubectl apply -f infra/k8s/worker-deployment.yaml
kubectl apply -f infra/k8s/web-deployment.yaml
kubectl apply -f infra/k8s/hpa.yaml
kubectl apply -f infra/k8s/ingress.yaml
```

### Configure Secrets

1. **Copy the example**:
   ```bash
   cp infra/k8s/secrets.yaml.example infra/k8s/secrets.yaml
   ```

2. **Update values** in `secrets.yaml`:
   ```yaml
   stringData:
     database-url: "postgresql://user:pass@host:5432/offgridflow?sslmode=require"
     redis-url: "redis://redis-service:6379/0"
     jwt-secret: "$(openssl rand -base64 32)"
     stripe-secret: "sk_live_..."
     openai-api-key: "sk-..."
   ```

3. **Apply secrets**:
   ```bash
   kubectl apply -f infra/k8s/secrets.yaml
   ```

### Using Terraform Outputs

If you've deployed infrastructure with Terraform:

```bash
# Navigate to Terraform directory
cd infra/terraform

# Get database URL
terraform output -raw database_url

# Create secrets from outputs
kubectl create secret generic offgridflow-secrets \
  --from-literal=database-url="$(terraform output -raw database_url)" \
  --from-literal=redis-url="$(terraform output -raw redis_url)" \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  --namespace=offgridflow \
  --dry-run=client -o yaml | kubectl apply -f -
```

### Verify Deployment

```bash
# Check pods
kubectl get pods -n offgridflow

# Check services
kubectl get svc -n offgridflow

# Check ingress
kubectl get ingress -n offgridflow

# View logs
kubectl logs -f deployment/offgridflow-api -n offgridflow
kubectl logs -f deployment/offgridflow-worker -n offgridflow
```

### Auto-Scaling

Horizontal Pod Autoscalers are configured:

- **API**: 2-10 pods (CPU 70%, Memory 80%)
- **Web**: 2-8 pods (CPU 70%, Memory 80%)
- **Worker**: 1-5 pods (CPU 75%, Memory 85%)

Monitor with:
```bash
kubectl get hpa -n offgridflow
```

### Migration Handling

Migrations run automatically via init container in the API deployment:
- Init container runs migrations before API starts
- Safe for multiple replicas (uses database locks)
- Failures are logged but don't prevent deployment

---

## Terraform Infrastructure

### Prerequisites

- Terraform 1.6+
- AWS CLI configured
- S3 bucket for state (or configure different backend)

### Infrastructure Components

The Terraform setup creates:

1. **VPC & Networking**
   - VPC with public/private subnets across 3 AZs
   - NAT Gateways for private subnet internet access
   - Security groups

2. **Database (RDS PostgreSQL)**
   - PostgreSQL 15.4
   - Multi-AZ for production
   - Automated backups
   - Encryption at rest

3. **Cache (ElastiCache Redis)**
   - Redis 7
   - Cluster mode for HA

4. **Storage (S3)**
   - Encrypted buckets
   - Lifecycle policies
   - Versioning enabled

5. **Queue (SQS)**
   - Multiple queues for different job types
   - Dead letter queues

6. **Compute (ECS Fargate)**
   - API and Worker services
   - Auto-scaling
   - Load balancers

### Deploy Infrastructure

1. **Initialize Terraform**:
   ```bash
   cd infra/terraform
   terraform init
   ```

2. **Configure variables**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your values
   ```

3. **Plan changes**:
   ```bash
   terraform plan
   ```

4. **Apply infrastructure**:
   ```bash
   terraform apply
   ```

5. **Get outputs**:
   ```bash
   terraform output
   terraform output -raw database_url
   ```

### Terraform Modules

- **vpc**: VPC, subnets, NAT gateways, routing
- **db**: RDS PostgreSQL with backups
- **cache**: ElastiCache Redis cluster
- **storage**: S3 buckets with policies
- **queue**: SQS queues and DLQs
- **api**: ECS Fargate services for API/Worker

### State Management

State is stored in S3 with DynamoDB locking:

```hcl
backend "s3" {
  bucket         = "offgridflow-terraform-state"
  key            = "production/terraform.tfstate"
  region         = "us-west-2"
  encrypt        = true
  dynamodb_table = "offgridflow-terraform-locks"
}
```

Create state bucket:
```bash
aws s3 mb s3://offgridflow-terraform-state --region us-west-2
aws dynamodb create-table \
  --table-name offgridflow-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-west-2
```

### Destroy Infrastructure

```bash
cd infra/terraform
terraform destroy
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

Located in `.github/workflows/ci.yml`, the pipeline:

1. **Backend Tests** (Go 1.24)
   - `go mod verify`
   - `go fmt`, `go vet`
   - `go test -race -coverprofile=coverage.out`
   - Upload coverage to Codecov

2. **Frontend Tests** (Node 20)
   - `npm ci`
   - `npm run lint`
   - `npm test`

3. **Docker Build** (on push to main/master)
   - Build API, Worker, Web images
   - Push to GitHub Container Registry
   - Tag with branch, SHA, semver

### Triggering Workflow

Automatically runs on:
- Push to `main` or `master`
- Pull requests to `main` or `master`

### Secrets Required

Configure in GitHub repository settings:

- `GITHUB_TOKEN` - Automatically provided
- (Optional) `CODECOV_TOKEN` - For coverage reports

### Manual Workflow Dispatch

Enable manual runs in `.github/workflows/ci.yml`:

```yaml
on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy'
        required: true
        default: 'staging'
```

### Image Versioning

Images are tagged with:
- Branch name (e.g., `main`)
- Git SHA (e.g., `main-abc123def`)
- Semantic version (if tagged, e.g., `v1.2.3`)

### Deployment from CI

Add deployment step to workflow:

```yaml
- name: Deploy to staging
  if: github.ref == 'refs/heads/main'
  run: |
    kubectl set image deployment/offgridflow-api \
      api=ghcr.io/${{ github.repository }}/offgridflow-api:${{ github.sha }} \
      -n offgridflow
```

---

## Troubleshooting

### Docker Compose Issues

**Problem**: Services won't start
```bash
# Check logs
docker-compose logs

# Rebuild
docker-compose down -v
docker-compose up --build
```

**Problem**: Port already in use
```bash
# Find process using port
lsof -i :8080  # Linux/macOS
netstat -ano | findstr :8080  # Windows

# Change port in docker-compose.yml or stop conflicting service
```

**Problem**: Database migration fails
```bash
# Check API logs
docker-compose logs api

# Manually run migrations
docker-compose exec api /app/offgridflow-api migrate up

# Reset database (CAUTION: destroys data)
docker-compose down -v
docker-compose up -d postgres
docker-compose up -d api
```

### Kubernetes Issues

**Problem**: Pods in CrashLoopBackOff
```bash
# Check pod logs
kubectl logs pod-name -n offgridflow

# Describe pod for events
kubectl describe pod pod-name -n offgridflow

# Check init container logs
kubectl logs pod-name -c migrate -n offgridflow
```

**Problem**: Secrets not found
```bash
# Verify secrets exist
kubectl get secrets -n offgridflow

# Check secret content
kubectl get secret offgridflow-secrets -o yaml -n offgridflow

# Recreate secrets
kubectl delete secret offgridflow-secrets -n offgridflow
kubectl apply -f infra/k8s/secrets.yaml
```

**Problem**: Ingress not working
```bash
# Check ingress
kubectl describe ingress offgridflow-ingress -n offgridflow

# Verify ingress controller is running
kubectl get pods -n ingress-nginx

# Check ingress class
kubectl get ingressclass
```

### Terraform Issues

**Problem**: State locked
```bash
# Force unlock (use with caution)
terraform force-unlock LOCK_ID
```

**Problem**: Resource already exists
```bash
# Import existing resource
terraform import aws_s3_bucket.example bucket-name

# Or remove from state
terraform state rm aws_s3_bucket.example
```

**Problem**: Plan shows unexpected changes
```bash
# Refresh state
terraform refresh

# Compare with remote
terraform plan -refresh-only
```

### General Debugging

**Enable debug mode**:
```bash
# API
export OFFGRIDFLOW_LOG_LEVEL=debug

# Terraform
export TF_LOG=DEBUG

# Kubernetes
kubectl logs --tail=100 -f deployment/offgridflow-api -n offgridflow
```

**Health checks**:
```bash
# API
curl http://localhost:8080/health

# Database
docker-compose exec postgres psql -U offgridflow -c "SELECT version();"

# Redis
docker-compose exec redis redis-cli ping
```

---

## Additional Resources

- [API Documentation](http://localhost:8080/swagger)
- [Architecture Guide](../docs/ARCHITECTURE.md)
- [Development Guide](../README.md)
- [Security Guide](../docs/SECURITY.md)

## Support

For issues or questions:
- GitHub Issues: https://github.com/example/offgridflow/issues
- Email: support@offgridflow.example.com
