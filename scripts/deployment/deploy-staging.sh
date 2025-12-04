#!/bin/bash
# OffGridFlow Staging Deployment Script

set -e

echo "================================================"
echo "OffGridFlow Staging Deployment"
echo "================================================"
echo ""

# Configuration
CLUSTER_NAME=${CLUSTER_NAME:-offgridflow-staging}
NAMESPACE=${NAMESPACE:-offgridflow}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-ghcr.io/your-org}
VERSION=${VERSION:-$(git rev-parse --short HEAD)}

echo "Deployment Configuration:"
echo "  Cluster: $CLUSTER_NAME"
echo "  Namespace: $NAMESPACE"
echo "  Registry: $DOCKER_REGISTRY"
echo "  Version: $VERSION"
echo ""

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found. Please install kubectl."
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "Error: docker not found. Please install Docker."
    exit 1
fi

echo "✓ Prerequisites check passed"
echo ""

# Build Docker images
echo "Building Docker images..."

echo "Building API image..."
docker build -t $DOCKER_REGISTRY/offgridflow-api:$VERSION -t $DOCKER_REGISTRY/offgridflow-api:staging .

echo "Building Web image..."
docker build -t $DOCKER_REGISTRY/offgridflow-web:$VERSION -t $DOCKER_REGISTRY/offgridflow-web:staging ./web

echo "✓ Docker images built"
echo ""

# Push images to registry
echo "Pushing images to registry..."
docker push $DOCKER_REGISTRY/offgridflow-api:$VERSION
docker push $DOCKER_REGISTRY/offgridflow-api:staging
docker push $DOCKER_REGISTRY/offgridflow-web:$VERSION
docker push $DOCKER_REGISTRY/offgridflow-web:staging

echo "✓ Images pushed to registry"
echo ""

# Create namespace if it doesn't exist
echo "Setting up Kubernetes namespace..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

echo "✓ Namespace ready"
echo ""

# Apply Kubernetes configurations
echo "Deploying to Kubernetes..."

# Deploy secrets (ensure these are created separately for security)
echo "Note: Ensure secrets are created separately using:"
echo "  kubectl create secret generic offgridflow-secrets --from-env-file=.env.staging -n $NAMESPACE"
echo ""

# Deploy OpenTelemetry Collector
echo "Deploying OpenTelemetry Collector..."
kubectl apply -f infra/k8s/otel-collector.yaml -n $NAMESPACE

# Deploy API
echo "Deploying API..."
kubectl apply -f infra/k8s/api-deployment.yaml -n $NAMESPACE

# Deploy Worker
echo "Deploying Worker..."
kubectl apply -f infra/k8s/worker-deployment.yaml -n $NAMESPACE

# Deploy Web
echo "Deploying Web..."
kubectl apply -f infra/k8s/web-deployment.yaml -n $NAMESPACE

# Deploy Ingress
echo "Deploying Ingress..."
kubectl apply -f infra/k8s/ingress.yaml -n $NAMESPACE

echo "✓ Kubernetes resources deployed"
echo ""

# Wait for deployments to be ready
echo "Waiting for deployments to be ready..."

kubectl rollout status deployment/offgridflow-api -n $NAMESPACE --timeout=5m
kubectl rollout status deployment/offgridflow-worker -n $NAMESPACE --timeout=5m
kubectl rollout status deployment/offgridflow-web -n $NAMESPACE --timeout=5m
kubectl rollout status deployment/otel-collector -n $NAMESPACE --timeout=5m

echo "✓ All deployments ready"
echo ""

# Display deployment status
echo "Deployment Status:"
kubectl get deployments -n $NAMESPACE
echo ""

echo "Services:"
kubectl get services -n $NAMESPACE
echo ""

echo "Pods:"
kubectl get pods -n $NAMESPACE
echo ""

# Get ingress info
echo "Ingress:"
kubectl get ingress -n $NAMESPACE
echo ""

echo "================================================"
echo "Deployment completed successfully!"
echo "================================================"
echo ""
echo "Next steps:"
echo "1. Verify health endpoints:"
echo "   kubectl port-forward svc/offgridflow-api 8080:8080 -n $NAMESPACE"
echo "   curl http://localhost:8080/health"
echo ""
echo "2. Check logs:"
echo "   kubectl logs -f deployment/offgridflow-api -n $NAMESPACE"
echo ""
echo "3. View Grafana dashboards"
echo "4. Run smoke tests"
echo ""
