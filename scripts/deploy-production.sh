#!/bin/bash
# OffGridFlow Production Deployment Script
# Orchestrates complete production deployment with safety checks
# Usage: ./scripts/deploy-production.sh [version]

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
VERSION=${1:-$(git describe --tags --always)}
NAMESPACE="offgridflow"
KUBECTL="kubectl"
DEPLOYMENT_TIMEOUT="600s"
ROLLBACK_ON_FAILURE="true"

# Validation flags
PRE_DEPLOYMENT_CHECKS_PASSED=false
DEPLOYMENT_SUCCESSFUL=false

log_info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✅${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ❌${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠️${NC} $1"
}

# Banner
clear
cat <<EOF
${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}
${BLUE}                 OffGridFlow Production Deployment${NC}
${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}
  Version:      ${VERSION}
  Namespace:    ${NAMESPACE}
  Timestamp:    $(date)
  User:         $(whoami)
${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}
EOF

# ============================================================================
# Pre-Deployment Checks
# ============================================================================
log_info "Running pre-deployment checks..."

# Check if kubectl is installed and configured
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl not found. Please install kubectl."
    exit 1
fi

# Check cluster connectivity
if ! kubectl cluster-info &> /dev/null; then
    log_error "Cannot connect to Kubernetes cluster"
    exit 1
fi
log_success "Kubernetes cluster connection verified"

# Verify correct cluster context
CURRENT_CONTEXT=$(kubectl config current-context)
log_info "Current context: ${CURRENT_CONTEXT}"

read -p "$(echo -e ${YELLOW}Is this the correct production cluster? [yes/no]: ${NC})" -r
if [[ ! $REPLY =~ ^[Yy](es)?$ ]]; then
    log_error "Deployment cancelled by user"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
    log_error "Namespace ${NAMESPACE} does not exist"
    exit 1
fi
log_success "Namespace ${NAMESPACE} exists"

# Verify Docker images exist
log_info "Verifying Docker images..."
IMAGES=(
    "ghcr.io/example/offgridflow-api:${VERSION}"
    "ghcr.io/example/offgridflow-worker:${VERSION}"
    "ghcr.io/example/offgridflow-web:${VERSION}"
)

for IMAGE in "${IMAGES[@]}"; do
    if docker manifest inspect "${IMAGE}" &> /dev/null; then
        log_success "Image found: ${IMAGE}"
    else
        log_error "Image not found: ${IMAGE}"
        log_error "Please build and push images first"
        exit 1
    fi
done

# Check database connectivity
log_info "Checking database connectivity..."
DB_POD=$(kubectl get pods -n ${NAMESPACE} -l app=postgres -o jsonpath='{.items[0].metadata.name}')
if kubectl exec -n ${NAMESPACE} ${DB_POD} -- pg_isready &> /dev/null; then
    log_success "Database is accessible"
else
    log_error "Cannot connect to database"
    exit 1
fi

# Verify secrets exist
log_info "Verifying secrets..."
if kubectl get secret offgridflow-secrets -n ${NAMESPACE} &> /dev/null; then
    log_success "Secrets configured"
else
    log_error "Secrets not found. Please configure secrets first."
    exit 1
fi

# Check disk space
log_info "Checking database disk space..."
DISK_USAGE=$(kubectl exec -n ${NAMESPACE} ${DB_POD} -- df -h /var/lib/postgresql/data | tail -1 | awk '{print $5}' | sed 's/%//')
if [ ${DISK_USAGE} -lt 85 ]; then
    log_success "Disk space: ${DISK_USAGE}% used"
else
    log_warning "Disk space: ${DISK_USAGE}% used (high)"
fi

# Run pre-deployment smoke tests on staging
log_info "Running smoke tests on staging..."
if ./scripts/smoke-tests.sh staging; then
    log_success "Staging smoke tests passed"
else
    log_error "Staging smoke tests failed"
    exit 1
fi

PRE_DEPLOYMENT_CHECKS_PASSED=true
log_success "All pre-deployment checks passed"

# ============================================================================
# Backup Current State
# ============================================================================
log_info "Creating backup of current state..."

# Backup database
log_info "Triggering database backup..."
kubectl create job --from=cronjob/postgres-backup backup-pre-deploy-${VERSION} -n ${NAMESPACE}
kubectl wait --for=condition=complete job/backup-pre-deploy-${VERSION} -n ${NAMESPACE} --timeout=300s
log_success "Database backup completed"

# Save current deployment state
mkdir -p backups
kubectl get all -n ${NAMESPACE} -o yaml > backups/pre-deploy-${VERSION}-$(date +%Y%m%d-%H%M%S).yaml
log_success "Current state backed up"

# ============================================================================
# Database Migrations
# ============================================================================
log_info "Running database migrations..."

# Create migration job
cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: migrate-${VERSION}
  namespace: ${NAMESPACE}
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: migrate
        image: ghcr.io/example/offgridflow-api:${VERSION}
        command: ["/app/migrate"]
        args: ["-command", "up"]
        env:
        - name: OFFGRIDFLOW_DB_DSN
          valueFrom:
            secretKeyRef:
              name: offgridflow-secrets
              key: database-url
EOF

# Wait for migration to complete
if kubectl wait --for=condition=complete job/migrate-${VERSION} -n ${NAMESPACE} --timeout=300s; then
    log_success "Database migrations completed"
else
    log_error "Database migrations failed"
    kubectl logs job/migrate-${VERSION} -n ${NAMESPACE}
    exit 1
fi

# Cleanup migration job
kubectl delete job migrate-${VERSION} -n ${NAMESPACE}

# ============================================================================
# Deploy Backend (API)
# ============================================================================
log_info "Deploying API service..."

# Update image in deployment
kubectl set image deployment/offgridflow-api \
    api=ghcr.io/example/offgridflow-api:${VERSION} \
    -n ${NAMESPACE}

# Wait for rollout
if kubectl rollout status deployment/offgridflow-api -n ${NAMESPACE} --timeout=${DEPLOYMENT_TIMEOUT}; then
    log_success "API deployment successful"
else
    log_error "API deployment failed"
    if [ "${ROLLBACK_ON_FAILURE}" = "true" ]; then
        log_warning "Rolling back API deployment..."
        kubectl rollout undo deployment/offgridflow-api -n ${NAMESPACE}
        kubectl rollout status deployment/offgridflow-api -n ${NAMESPACE}
        log_error "API rolled back to previous version"
    fi
    exit 1
fi

# ============================================================================
# Deploy Worker
# ============================================================================
log_info "Deploying Worker service..."

kubectl set image deployment/offgridflow-worker \
    worker=ghcr.io/example/offgridflow-worker:${VERSION} \
    -n ${NAMESPACE}

if kubectl rollout status deployment/offgridflow-worker -n ${NAMESPACE} --timeout=${DEPLOYMENT_TIMEOUT}; then
    log_success "Worker deployment successful"
else
    log_error "Worker deployment failed"
    if [ "${ROLLBACK_ON_FAILURE}" = "true" ]; then
        log_warning "Rolling back Worker deployment..."
        kubectl rollout undo deployment/offgridflow-worker -n ${NAMESPACE}
        kubectl rollout status deployment/offgridflow-worker -n ${NAMESPACE}
        log_error "Worker rolled back to previous version"
    fi
    exit 1
fi

# ============================================================================
# Deploy Frontend (Web)
# ============================================================================
log_info "Deploying Web frontend..."

kubectl set image deployment/offgridflow-web \
    web=ghcr.io/example/offgridflow-web:${VERSION} \
    -n ${NAMESPACE}

if kubectl rollout status deployment/offgridflow-web -n ${NAMESPACE} --timeout=${DEPLOYMENT_TIMEOUT}; then
    log_success "Web deployment successful"
else
    log_error "Web deployment failed"
    if [ "${ROLLBACK_ON_FAILURE}" = "true" ]; then
        log_warning "Rolling back Web deployment..."
        kubectl rollout undo deployment/offgridflow-web -n ${NAMESPACE}
        kubectl rollout status deployment/offgridflow-web -n ${NAMESPACE}
        log_error "Web rolled back to previous version"
    fi
    exit 1
fi

DEPLOYMENT_SUCCESSFUL=true

# ============================================================================
# Post-Deployment Verification
# ============================================================================
log_info "Running post-deployment verification..."

# Wait for pods to be ready
sleep 10

# Check pod status
log_info "Checking pod status..."
kubectl get pods -n ${NAMESPACE} -l app=offgridflow-api
kubectl get pods -n ${NAMESPACE} -l app=offgridflow-worker
kubectl get pods -n ${NAMESPACE} -l app=offgridflow-web

# Health checks
log_info "Performing health checks..."
API_URL="https://api.offgridflow.com"

if curl -f -s ${API_URL}/health/live > /dev/null; then
    log_success "Liveness check passed"
else
    log_error "Liveness check failed"
    exit 1
fi

if curl -f -s ${API_URL}/health/ready > /dev/null; then
    log_success "Readiness check passed"
else
    log_error "Readiness check failed"
    exit 1
fi

# Run smoke tests
log_info "Running production smoke tests..."
if ./scripts/smoke-tests.sh production; then
    log_success "Production smoke tests passed"
else
    log_error "Production smoke tests failed"
    log_error "Deployment verification failed. Consider rollback."
    exit 1
fi

# Check for errors in logs
log_info "Checking recent logs for errors..."
ERROR_COUNT=$(kubectl logs -n ${NAMESPACE} -l app=offgridflow-api --tail=100 --since=5m | grep -i "ERROR\|FATAL" | wc -l)
if [ ${ERROR_COUNT} -eq 0 ]; then
    log_success "No errors found in recent logs"
else
    log_warning "Found ${ERROR_COUNT} errors in recent logs"
fi

# ============================================================================
# Monitoring Setup
# ============================================================================
log_info "Setting up monitoring..."

# Verify Prometheus is scraping
PROMETHEUS_URL="http://prometheus.offgridflow.com"
if curl -s "${PROMETHEUS_URL}/api/v1/targets" | jq -e '.data.activeTargets[] | select(.labels.job=="offgridflow-api") | .health == "up"' > /dev/null; then
    log_success "Prometheus scraping API metrics"
else
    log_warning "Prometheus may not be scraping metrics properly"
fi

# ============================================================================
# Final Summary
# ============================================================================
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}                 🎉 Deployment Successful! 🎉${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  Version Deployed:     ${GREEN}${VERSION}${NC}"
echo -e "  Deployment Time:      $(date)"
echo -e "  Duration:             ${SECONDS}s"
echo ""
echo -e "  ${BLUE}Service URLs:${NC}"
echo -e "    API:       https://api.offgridflow.com"
echo -e "    Web:       https://app.offgridflow.com"
echo -e "    Grafana:   https://grafana.offgridflow.com"
echo -e "    Prometheus: https://prometheus.offgridflow.com"
echo ""
echo -e "  ${BLUE}Next Steps:${NC}"
echo -e "    1. Monitor metrics in Grafana for 2 hours"
echo -e "    2. Check for alerts in Prometheus"
echo -e "    3. Monitor error rates and latency"
echo -e "    4. Communicate deployment to team"
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Save deployment metadata
cat > "deployments/history/deploy-${VERSION}-$(date +%Y%m%d-%H%M%S).json" <<EOF
{
  "version": "${VERSION}",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "deployer": "$(whoami)",
  "duration_seconds": ${SECONDS},
  "success": true,
  "images": {
    "api": "ghcr.io/example/offgridflow-api:${VERSION}",
    "worker": "ghcr.io/example/offgridflow-worker:${VERSION}",
    "web": "ghcr.io/example/offgridflow-web:${VERSION}"
  }
}
EOF

exit 0
