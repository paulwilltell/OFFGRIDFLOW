# OffGridFlow Staging Deployment Script (PowerShell)

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "OffGridFlow Staging Deployment" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$CLUSTER_NAME = if ($env:CLUSTER_NAME) { $env:CLUSTER_NAME } else { "offgridflow-staging" }
$NAMESPACE = if ($env:NAMESPACE) { $env:NAMESPACE } else { "offgridflow" }
$DOCKER_REGISTRY = if ($env:DOCKER_REGISTRY) { $env:DOCKER_REGISTRY } else { "ghcr.io/your-org" }

# Get git commit hash for version
try {
    $VERSION = git rev-parse --short HEAD
} catch {
    $VERSION = "latest"
}

if ($env:VERSION) {
    $VERSION = $env:VERSION
}

Write-Host "Deployment Configuration:"
Write-Host "  Cluster: $CLUSTER_NAME"
Write-Host "  Namespace: $NAMESPACE"
Write-Host "  Registry: $DOCKER_REGISTRY"
Write-Host "  Version: $VERSION"
Write-Host ""

# Check prerequisites
Write-Host "Checking prerequisites..."

try {
    $null = Get-Command kubectl -ErrorAction Stop
    Write-Host "  ✓ kubectl found" -ForegroundColor Green
} catch {
    Write-Host "  ✗ kubectl not found. Please install kubectl." -ForegroundColor Red
    exit 1
}

try {
    $null = Get-Command docker -ErrorAction Stop
    Write-Host "  ✓ docker found" -ForegroundColor Green
} catch {
    Write-Host "  ✗ docker not found. Please install Docker." -ForegroundColor Red
    exit 1
}

Write-Host ""

# Build Docker images
Write-Host "Building Docker images..." -ForegroundColor Yellow

Write-Host "  Building API image..."
docker build -t "${DOCKER_REGISTRY}/offgridflow-api:$VERSION" -t "${DOCKER_REGISTRY}/offgridflow-api:staging" .
if ($LASTEXITCODE -ne 0) {
    Write-Host "  ✗ Failed to build API image" -ForegroundColor Red
    exit 1
}

Write-Host "  Building Web image..."
docker build -t "${DOCKER_REGISTRY}/offgridflow-web:$VERSION" -t "${DOCKER_REGISTRY}/offgridflow-web:staging" ./web
if ($LASTEXITCODE -ne 0) {
    Write-Host "  ✗ Failed to build Web image" -ForegroundColor Red
    exit 1
}

Write-Host "  ✓ Docker images built" -ForegroundColor Green
Write-Host ""

# Push images to registry
Write-Host "Pushing images to registry..." -ForegroundColor Yellow

docker push "${DOCKER_REGISTRY}/offgridflow-api:$VERSION"
docker push "${DOCKER_REGISTRY}/offgridflow-api:staging"
docker push "${DOCKER_REGISTRY}/offgridflow-web:$VERSION"
docker push "${DOCKER_REGISTRY}/offgridflow-web:staging"

Write-Host "  ✓ Images pushed to registry" -ForegroundColor Green
Write-Host ""

# Create namespace if it doesn't exist
Write-Host "Setting up Kubernetes namespace..." -ForegroundColor Yellow
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

Write-Host "  ✓ Namespace ready" -ForegroundColor Green
Write-Host ""

# Apply Kubernetes configurations
Write-Host "Deploying to Kubernetes..." -ForegroundColor Yellow

# Deploy secrets reminder
Write-Host "Note: Ensure secrets are created separately using:" -ForegroundColor Yellow
Write-Host "  kubectl create secret generic offgridflow-secrets --from-env-file=.env.staging -n $NAMESPACE" -ForegroundColor Cyan
Write-Host ""

# Deploy OpenTelemetry Collector
Write-Host "  Deploying OpenTelemetry Collector..."
kubectl apply -f infra/k8s/otel-collector.yaml -n $NAMESPACE

# Deploy API
Write-Host "  Deploying API..."
kubectl apply -f infra/k8s/api-deployment.yaml -n $NAMESPACE

# Deploy Worker
Write-Host "  Deploying Worker..."
kubectl apply -f infra/k8s/worker-deployment.yaml -n $NAMESPACE

# Deploy Web
Write-Host "  Deploying Web..."
kubectl apply -f infra/k8s/web-deployment.yaml -n $NAMESPACE

# Deploy Ingress
Write-Host "  Deploying Ingress..."
kubectl apply -f infra/k8s/ingress.yaml -n $NAMESPACE

Write-Host "  ✓ Kubernetes resources deployed" -ForegroundColor Green
Write-Host ""

# Wait for deployments to be ready
Write-Host "Waiting for deployments to be ready..." -ForegroundColor Yellow

kubectl rollout status deployment/offgridflow-api -n $NAMESPACE --timeout=5m
kubectl rollout status deployment/offgridflow-worker -n $NAMESPACE --timeout=5m
kubectl rollout status deployment/offgridflow-web -n $NAMESPACE --timeout=5m
kubectl rollout status deployment/otel-collector -n $NAMESPACE --timeout=5m

Write-Host "  ✓ All deployments ready" -ForegroundColor Green
Write-Host ""

# Display deployment status
Write-Host "Deployment Status:" -ForegroundColor Cyan
kubectl get deployments -n $NAMESPACE
Write-Host ""

Write-Host "Services:" -ForegroundColor Cyan
kubectl get services -n $NAMESPACE
Write-Host ""

Write-Host "Pods:" -ForegroundColor Cyan
kubectl get pods -n $NAMESPACE
Write-Host ""

# Get ingress info
Write-Host "Ingress:" -ForegroundColor Cyan
kubectl get ingress -n $NAMESPACE
Write-Host ""

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Deployment completed successfully!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:"
Write-Host "1. Verify health endpoints:"
Write-Host "   kubectl port-forward svc/offgridflow-api 8080:8080 -n $NAMESPACE"
Write-Host "   curl http://localhost:8080/health"
Write-Host ""
Write-Host "2. Check logs:"
Write-Host "   kubectl logs -f deployment/offgridflow-api -n $NAMESPACE"
Write-Host ""
Write-Host "3. View Grafana dashboards"
Write-Host "4. Run smoke tests"
Write-Host ""
