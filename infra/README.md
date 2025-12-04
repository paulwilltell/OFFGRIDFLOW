# OffGridFlow Infrastructure

Complete infrastructure setup for OffGridFlow with Docker, Kubernetes, and Terraform.

## Quick Start - Local Development

### Prerequisites
- Docker and Docker Compose
- Go 1.24
- Node.js 20
- PostgreSQL 15 (or use Docker)

### One Command Setup

```bash
docker-compose up
```

This will:
- Start PostgreSQL, Redis, Jaeger, Prometheus, Grafana
- Run database migrations automatically
- Start the API server on http://localhost:8080
- Start the worker for background jobs
- Start the web frontend on http://localhost:3000

### Access Services

- **API**: http://localhost:8080
- **Web UI**: http://localhost:3000
- **Jaeger (Tracing)**: http://localhost:16686
- **Prometheus (Metrics)**: http://localhost:9090
- **Grafana (Dashboards)**: http://localhost:3001 (admin/admin)
- **PostgreSQL**: localhost:5432 (offgridflow/changeme)
- **Redis**: localhost:6379

### Environment Configuration

Copy `.env.production.template` to `.env` and configure:

```bash
cp .env.production.template .env
# Edit .env with your credentials
```

Required variables:
- `OFFGRIDFLOW_DB_DSN` - PostgreSQL connection string
- `OFFGRIDFLOW_JWT_SECRET` - Secret for JWT tokens
- `OFFGRIDFLOW_OPENAI_API_KEY` - OpenAI API key (optional)
- `OFFGRIDFLOW_STRIPE_SECRET_KEY` - Stripe secret key (optional)

## Docker Images

### Building Images

```bash
# Build API image
docker build -t offgridflow-api:latest .

# Build worker image
docker build --target worker -t offgridflow-worker:latest .

# Build web image
docker build -t offgridflow-web:latest ./web
```

### Available Images

- **offgridflow-api**: Main API server
- **offgridflow-worker**: Background job processor
- **offgridflow-web**: Next.js frontend

## Kubernetes Deployment

### Prerequisites
- Kubernetes cluster (EKS, GKE, AKS, or local with minikube/kind)
- kubectl configured
- Helm (optional, for cert-manager and ingress)

### Setup

1. **Create namespace**:
```bash
kubectl create namespace offgridflow
```

2. **Create secrets**:
```bash
# From Terraform outputs
kubectl create secret generic offgridflow-secrets \
  --from-literal=database-url="$(terraform -chdir=infra/terraform output -raw database_url)" \
  --from-literal=redis-url="$(terraform -chdir=infra/terraform output -raw redis_url)" \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  --namespace offgridflow

# Add additional secrets
kubectl create secret generic offgridflow-secrets \
  --from-literal=stripe-secret="${STRIPE_SECRET_KEY}" \
  --from-literal=openai-api-key="${OPENAI_API_KEY}" \
  --namespace offgridflow \
  --dry-run=client -o yaml | kubectl apply -f -
```

Or use the example template:
```bash
cp infra/k8s/secrets.yaml.example infra/k8s/secrets.yaml
# Edit secrets.yaml with real values
kubectl apply -f infra/k8s/secrets.yaml -n offgridflow
```

3. **Deploy ConfigMap**:
```bash
kubectl apply -f infra/k8s/configmap.yaml -n offgridflow
```

4. **Deploy Services**:
```bash
kubectl apply -f infra/k8s/services.yaml -n offgridflow
```

5. **Deploy Applications**:
```bash
kubectl apply -f infra/k8s/api-deployment.yaml -n offgridflow
kubectl apply -f infra/k8s/worker-deployment.yaml -n offgridflow
kubectl apply -f infra/k8s/web-deployment.yaml -n offgridflow
```

6. **Deploy HPA (Auto-scaling)**:
```bash
kubectl apply -f infra/k8s/hpa.yaml -n offgridflow
```

7. **Deploy Ingress**:
```bash
kubectl apply -f infra/k8s/ingress.yaml -n offgridflow
```

### Verify Deployment

```bash
# Check pods
kubectl get pods -n offgridflow

# Check services
kubectl get svc -n offgridflow

# Check HPA status
kubectl get hpa -n offgridflow

# View logs
kubectl logs -f deployment/offgridflow-api -n offgridflow
kubectl logs -f deployment/offgridflow-worker -n offgridflow
```

### Update Deployments

```bash
# Rolling update with new image
kubectl set image deployment/offgridflow-api \
  api=ghcr.io/example/offgridflow-api:v1.2.3 \
  -n offgridflow

# Rollback if needed
kubectl rollout undo deployment/offgridflow-api -n offgridflow
```

## Terraform Infrastructure

### Prerequisites
- Terraform >= 1.6.0
- AWS CLI configured
- S3 bucket for Terraform state (or update backend in main.tf)

### Initialize

```bash
cd infra/terraform

# Initialize Terraform
terraform init

# Create terraform.tfvars from example
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
```

### Plan and Apply

```bash
# Review changes
terraform plan

# Apply infrastructure
terraform apply

# Get outputs
terraform output
```

### Outputs

After applying, Terraform provides:
- `api_endpoint` - Load balancer DNS for API
- `database_endpoint` - RDS PostgreSQL endpoint
- `redis_endpoint` - ElastiCache Redis endpoint
- `storage_bucket` - S3 bucket name

Use these outputs to configure Kubernetes secrets:

```bash
# Database URL
terraform output -raw database_endpoint
# Returns: offgridflow-production-postgres.xyz.us-west-2.rds.amazonaws.com:5432

# Redis URL  
terraform output -raw redis_endpoint
# Returns: offgridflow-production-redis.xyz.cache.amazonaws.com

# Create K8s secret
kubectl create secret generic offgridflow-secrets \
  --from-literal=database-url="postgresql://$(terraform output -raw db_username):$(terraform output -raw db_password)@$(terraform output -raw database_endpoint)/offgridflow?sslmode=require" \
  --from-literal=redis-url="redis://$(terraform output -raw redis_endpoint):6379/0" \
  --namespace offgridflow
```

### Modules

Infrastructure is modular:
- **vpc**: VPC, subnets, NAT gateways, route tables
- **db**: RDS PostgreSQL with backups and encryption
- **cache**: ElastiCache Redis
- **storage**: S3 bucket with versioning and lifecycle
- **queue**: SQS/SNS for async messaging
- **api**: ECS Fargate or EKS for API deployment

### Destroy

```bash
# Destroy all infrastructure (careful!)
terraform destroy
```

## CI/CD Pipeline

### GitHub Actions

The `.github/workflows/ci.yml` pipeline:

1. **Backend Tests**
   - Go 1.24
   - Run tests with race detection
   - Upload coverage

2. **Frontend Tests**
   - Node.js 20
   - Lint with ESLint
   - Run Jest tests

3. **Docker Build** (on main/master push)
   - Build API, Worker, Web images
   - Tag with commit SHA and branch
   - Push to GitHub Container Registry

### Manual Deployment

```bash
# Pull latest images
docker pull ghcr.io/example/offgridflow-api:main
docker pull ghcr.io/example/offgridflow-worker:main
docker pull ghcr.io/example/offgridflow-web:main

# Update K8s deployments
kubectl set image deployment/offgridflow-api api=ghcr.io/example/offgridflow-api:main -n offgridflow
kubectl set image deployment/offgridflow-worker worker=ghcr.io/example/offgridflow-worker:main -n offgridflow
kubectl set image deployment/offgridflow-web web=ghcr.io/example/offgridflow-web:main -n offgridflow
```

## Monitoring & Observability

### Metrics (Prometheus)

Metrics exposed at `/metrics`:
- HTTP request duration
- Database query performance
- Background job processing
- Business metrics (emissions calculated, activities ingested)

### Tracing (Jaeger)

Distributed tracing with OpenTelemetry:
- HTTP requests end-to-end
- Database queries
- External API calls
- Background jobs

### Logging

Structured JSON logs to stdout:
- Request/response logging
- Error tracking
- Audit logs

## Scaling

### Horizontal Pod Autoscaler (HPA)

Auto-scaling based on CPU/memory:
- **API**: 2-10 replicas (70% CPU, 80% memory)
- **Web**: 2-8 replicas (70% CPU, 80% memory)
- **Worker**: 1-5 replicas (75% CPU, 85% memory)

### Database Scaling

- RDS: Vertical scaling (change instance class)
- Read replicas for read-heavy workloads
- Connection pooling in application

### Cache Scaling

- ElastiCache: Add nodes or upgrade instance type
- Redis Cluster mode for horizontal scaling

## Security

### Network Security

- Private subnets for databases and caches
- Security groups restrict access
- NAT gateways for outbound internet

### Secrets Management

- Kubernetes Secrets for sensitive data
- AWS Secrets Manager integration (optional)
- No secrets in code or Docker images

### TLS/SSL

- RDS encryption at rest
- S3 server-side encryption
- HTTPS with cert-manager (K8s)

## Troubleshooting

### Migrations not running

```bash
# Check API logs
kubectl logs deployment/offgridflow-api -n offgridflow | grep migration

# Manually run migrations
kubectl exec -it deployment/offgridflow-api -n offgridflow -- /app/offgridflow-api migrate
```

### Database connection issues

```bash
# Test from pod
kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \
  psql postgresql://user:pass@host:5432/dbname

# Check security groups allow traffic
```

### Worker not processing jobs

```bash
# Check worker logs
kubectl logs deployment/offgridflow-worker -n offgridflow

# Check Redis connection
kubectl exec -it deployment/offgridflow-worker -n offgridflow -- \
  redis-cli -h redis-service ping
```

## Cost Optimization

### Development Environment

- Use smaller instance types (t3.micro)
- Single-AZ deployments
- Shorter backup retention
- Spot instances for worker nodes

### Production Environment

- Multi-AZ for high availability
- Reserved instances for steady workloads
- Auto-scaling for variable load
- S3 lifecycle policies for old data

## Support

For issues:
1. Check logs: `kubectl logs -f deployment/offgridflow-api -n offgridflow`
2. Check pod status: `kubectl describe pod <pod-name> -n offgridflow`
3. Review metrics in Grafana
4. Check traces in Jaeger
