# Infrastructure & DevOps - 100% Complete ✅

## Overview

The OffGridFlow infrastructure is production-ready with:

- ✅ **One-command local setup**
- ✅ **Separate Docker images** for API, Worker, Web
- ✅ **Complete Kubernetes manifests** with namespace, autoscaling, and migrations
- ✅ **Full Terraform infrastructure** for AWS (VPC, RDS, Redis, S3, SQS, ECS)
- ✅ **CI/CD pipeline** with tests, builds, and image publishing

## Quick Start

### Local Development (One Command)

**Linux/macOS:**
```bash
make start
# or
./scripts/dev-start.sh
```

**Windows:**
```powershell
.\scripts\dev-start.ps1
```

**Using Docker Compose:**
```bash
docker-compose up
```

This starts:
- PostgreSQL (with auto-migrations)
- Redis
- API server
- Worker process
- Web frontend
- Full observability stack (Jaeger, Prometheus, Grafana)

### Access Services

- **Web UI**: http://localhost:3000
- **API**: http://localhost:8080
- **API Docs**: http://localhost:8080/swagger
- **Grafana**: http://localhost:3001
- **Jaeger**: http://localhost:16686

## Docker Images

### Building Images

```bash
make docker-build
```

This creates three optimized images:

1. **offgridflow-api** (~30MB)
   - Go 1.24, Alpine-based
   - Runs API server
   - Auto-runs migrations on startup

2. **offgridflow-worker** (~25MB)
   - Same base as API
   - Background job processor
   - Handles emissions calculations, connector syncs

3. **offgridflow-web** (~150MB)
   - Node 20, Next.js standalone
   - Optimized production build

### Pushing to Registry

```bash
make docker-push
```

Images are tagged with:
- Branch name
- Git SHA
- Semantic version (if tagged)

## Kubernetes Deployment

### Prerequisites

- Kubernetes 1.24+
- kubectl configured
- Namespace and secrets created

### Deploy

```bash
# Complete deployment
make k8s-deploy

# Or step by step
kubectl apply -f infra/k8s/namespace.yaml
kubectl apply -f infra/k8s/configmap.yaml
kubectl apply -f infra/k8s/secrets.yaml
kubectl apply -f infra/k8s/services.yaml
kubectl apply -f infra/k8s/api-deployment.yaml
kubectl apply -f infra/k8s/worker-deployment.yaml
kubectl apply -f infra/k8s/web-deployment.yaml
kubectl apply -f infra/k8s/hpa.yaml
kubectl apply -f infra/k8s/ingress.yaml
```

### Features

- **Namespace isolation**: Everything in `offgridflow` namespace
- **Auto-migrations**: Init container runs migrations before API starts
- **Auto-scaling**: HPA for API (2-10), Web (2-8), Worker (1-5)
- **Health checks**: Liveness, readiness, startup probes
- **Resource limits**: CPU and memory quotas
- **TLS/SSL**: Ingress with cert-manager support

### Configure Secrets

```bash
# Copy example
cp infra/k8s/secrets.yaml.example infra/k8s/secrets.yaml

# Edit with real values
vim infra/k8s/secrets.yaml

# Apply
kubectl apply -f infra/k8s/secrets.yaml
```

Or use Terraform outputs:

```bash
cd infra/terraform
kubectl create secret generic offgridflow-secrets \
  --from-literal=database-url="$(terraform output -raw database_url)" \
  --from-literal=redis-url="$(terraform output -raw redis_url)" \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  --namespace=offgridflow
```

## Terraform Infrastructure

### Components

The Terraform configuration creates:

- **VPC & Networking**: Multi-AZ VPC with public/private subnets
- **Database**: RDS PostgreSQL 15 with automated backups
- **Cache**: ElastiCache Redis cluster
- **Storage**: S3 buckets with encryption and lifecycle policies
- **Queue**: SQS queues with DLQs for different job types
- **Compute**: ECS Fargate services with load balancing

### Deploy Infrastructure

```bash
cd infra/terraform

# Initialize
terraform init

# Configure
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars

# Plan
terraform plan

# Apply
terraform apply

# Get outputs
terraform output
terraform output -raw database_url
```

### State Management

State is stored in S3 with DynamoDB locking for team collaboration.

Create state bucket:
```bash
aws s3 mb s3://offgridflow-terraform-state
aws dynamodb create-table \
  --table-name offgridflow-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST
```

## CI/CD Pipeline

### GitHub Actions Workflow

Located in `.github/workflows/ci.yml`:

1. **Backend Tests** (Go 1.24)
   - Linting: `go fmt`, `go vet`
   - Tests: `go test -race -coverprofile`
   - Coverage upload to Codecov

2. **Frontend Tests** (Node 20)
   - Install: `npm ci`
   - Lint: `npm run lint`
   - Test: `npm test`

3. **Docker Build** (on main/master push)
   - Build all 3 images
   - Push to ghcr.io
   - Tag with branch, SHA, semver

### Triggers

- Push to `main` or `master`
- Pull requests
- Manual dispatch (optional)

### Required Secrets

- `GITHUB_TOKEN` - Auto-provided
- `CODECOV_TOKEN` - Optional for coverage

## Architecture Diagram

```
┌─────────────────────────────────────────────┐
│         Load Balancer / Ingress             │
└─────────────┬───────────────────────────────┘
              │
    ┌─────────┴─────────┐
    │                   │
┌───▼────┐        ┌─────▼──────┐
│  API   │        │    Web     │
│(2-10)  │        │   (2-8)    │
└───┬────┘        └────────────┘
    │
    ├──────────┬──────────┐
    │          │          │
┌───▼──┐   ┌──▼───┐  ┌──▼─────┐
│ PG   │   │Redis │  │ Worker │
│ SQL  │   │      │  │ (1-5)  │
└──────┘   └──────┘  └────────┘
```

## Monitoring & Observability

### Metrics

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001

### Tracing

- **Jaeger**: http://localhost:16686
- OpenTelemetry collector configured

### Logs

```bash
# Docker Compose
docker-compose logs -f api
docker-compose logs -f worker

# Kubernetes
kubectl logs -f deployment/offgridflow-api -n offgridflow
kubectl logs -f deployment/offgridflow-worker -n offgridflow
```

## Testing

### Run All Tests

```bash
make test
```

This runs:
- Go backend tests
- Frontend tests
- Integration tests

### Individual Test Suites

```bash
# Backend only
go test -v ./...

# Frontend only
cd web && npm test

# With coverage
go test -coverprofile=coverage.out ./...
```

## Troubleshooting

### Local Development

**Services won't start:**
```bash
docker-compose down -v
docker-compose up --build
```

**Port conflicts:**
```bash
# Check what's using the port
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows
```

**Database issues:**
```bash
# Reset database
docker-compose down -v
docker-compose up -d postgres
docker-compose up -d api  # Migrations run automatically
```

### Kubernetes

**Pods crash:**
```bash
kubectl logs pod-name -n offgridflow
kubectl describe pod pod-name -n offgridflow
kubectl logs pod-name -c migrate -n offgridflow  # Init container
```

**Secrets missing:**
```bash
kubectl get secrets -n offgridflow
kubectl delete secret offgridflow-secrets -n offgridflow
kubectl apply -f infra/k8s/secrets.yaml
```

### Terraform

**State locked:**
```bash
terraform force-unlock LOCK_ID
```

**Resource conflicts:**
```bash
terraform import aws_s3_bucket.example bucket-name
```

## Performance Tuning

### API Auto-Scaling

HPA scales based on:
- CPU: 70% average
- Memory: 80% average

Adjust in `infra/k8s/hpa.yaml`:
```yaml
metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Database

Adjust instance class in `terraform.tfvars`:
```hcl
db_instance_class = "db.t3.medium"  # or larger
```

### Cache

Adjust Redis node type:
```hcl
redis_node_type = "cache.t3.micro"  # or larger
```

## Security

### Secrets Management

- Never commit secrets to Git
- Use Kubernetes secrets or AWS Secrets Manager
- Rotate credentials regularly

### Network Security

- Private subnets for databases
- Security groups restrict access
- TLS/SSL for all external traffic

### Access Control

- IAM roles for ECS tasks
- RBAC for Kubernetes
- Least privilege principle

## Maintenance

### Backups

**Database:**
- Automated daily backups (RDS)
- 7-day retention (configurable)

**Application data:**
- S3 versioning enabled
- Lifecycle policies for archival

### Updates

**Dependencies:**
```bash
go get -u ./...
cd web && npm update
```

**Infrastructure:**
```bash
cd infra/terraform
terraform plan
terraform apply
```

## Production Checklist

- [ ] Configure production secrets
- [ ] Set up SSL certificates
- [ ] Configure DNS records
- [ ] Set up monitoring alerts
- [ ] Configure log retention
- [ ] Set up backup policies
- [ ] Review security groups
- [ ] Enable multi-AZ for RDS
- [ ] Configure auto-scaling limits
- [ ] Set up disaster recovery plan

## Additional Resources

- [Full Infrastructure Guide](../docs/INFRASTRUCTURE.md)
- [API Documentation](http://localhost:8080/swagger)
- [Architecture Overview](../docs/ARCHITECTURE.md)

## Support

For issues:
- GitHub Issues
- Email: devops@offgridflow.example.com
