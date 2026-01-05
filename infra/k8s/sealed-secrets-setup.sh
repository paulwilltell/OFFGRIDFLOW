#!/bin/bash
# OffGridFlow Sealed Secrets Setup
# Installs and configures Sealed Secrets for secure secret management
# Requires: kubectl, kubeseal, helm

set -euo pipefail

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi
    
    if ! command -v helm &> /dev/null; then
        log_error "helm is not installed"
        exit 1
    fi
    
    if ! command -v kubeseal &> /dev/null; then
        log_warn "kubeseal is not installed. Installing..."
        install_kubeseal
    fi
    
    log_info "✅ All prerequisites met"
}

# Install kubeseal CLI
install_kubeseal() {
    KUBESEAL_VERSION="0.24.0"
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    if [ "$ARCH" = "x86_64" ]; then
        ARCH="amd64"
    elif [ "$ARCH" = "aarch64" ]; then
        ARCH="arm64"
    fi
    
    log_info "Downloading kubeseal v${KUBESEAL_VERSION} for ${OS}/${ARCH}..."
    
    curl -L "https://github.com/bitnami-labs/sealed-secrets/releases/download/v${KUBESEAL_VERSION}/kubeseal-${KUBESEAL_VERSION}-${OS}-${ARCH}.tar.gz" \
        -o /tmp/kubeseal.tar.gz
    
    tar xzf /tmp/kubeseal.tar.gz -C /tmp
    sudo mv /tmp/kubeseal /usr/local/bin/
    sudo chmod +x /usr/local/bin/kubeseal
    rm /tmp/kubeseal.tar.gz
    
    log_info "✅ kubeseal installed successfully"
}

# Install Sealed Secrets controller
install_controller() {
    log_info "Installing Sealed Secrets controller..."
    
    # Add Sealed Secrets Helm repository
    helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
    helm repo update
    
    # Install Sealed Secrets controller
    helm upgrade --install sealed-secrets sealed-secrets/sealed-secrets \
        --namespace kube-system \
        --set commandArgs[0]="--update-status" \
        --set commandArgs[1]="--key-renew-period=720h" \
        --create-namespace \
        --wait
    
    log_info "✅ Sealed Secrets controller installed"
    
    # Wait for controller to be ready
    log_info "Waiting for controller to be ready..."
    kubectl wait --for=condition=available --timeout=300s \
        deployment/sealed-secrets -n kube-system
    
    log_info "✅ Controller is ready"
}

# Get public sealing key
get_public_key() {
    log_info "Fetching public sealing key..."
    
    kubeseal --fetch-cert \
        --controller-name=sealed-secrets \
        --controller-namespace=kube-system \
        > infra/k8s/sealed-secrets-public-cert.pem
    
    log_info "✅ Public key saved to: infra/k8s/sealed-secrets-public-cert.pem"
    log_warn "⚠️  Keep this file in version control (it's safe to commit)"
}

# Create sealed secrets from template
create_sealed_secrets() {
    log_info "Creating sealed secrets for OffGridFlow..."
    
    # Check if secrets.yaml.example exists
    if [ ! -f "infra/k8s/secrets.yaml.example" ]; then
        log_error "secrets.yaml.example not found"
        exit 1
    fi
    
    # Prompt user to fill in secrets
    log_warn "You need to provide actual secret values"
    log_info "Creating temporary secrets file..."
    
    # Generate random secrets
    JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')
    POSTGRES_PASSWORD=$(openssl rand -base64 32 | tr -d '\n')
    REDIS_PASSWORD=$(openssl rand -base64 32 | tr -d '\n')
    
    log_info "Generated secure random secrets"
    
    # Create temporary unsealed secrets file
    cat > /tmp/offgridflow-secrets.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: offgridflow-secrets
  namespace: offgridflow
type: Opaque
stringData:
  # Database
  postgres-password: "${POSTGRES_PASSWORD}"
  database-url: "postgresql://offgridflow:${POSTGRES_PASSWORD}@postgres-service:5432/offgridflow?sslmode=disable"
  
  # Redis
  redis-password: "${REDIS_PASSWORD}"
  redis-url: "redis://:${REDIS_PASSWORD}@redis-service:6379/0"
  
  # JWT
  jwt-secret: "${JWT_SECRET}"
  
  # Stripe (add your keys)
  stripe-secret: "sk_live_REPLACE_WITH_YOUR_KEY"
  stripe-webhook-secret: "whsec_REPLACE_WITH_YOUR_SECRET"
  
  # OpenAI (optional)
  openai-api-key: "sk-REPLACE_OR_LEAVE_EMPTY"
  
  # AWS Credentials (for cloud ingestion)
  aws-access-key-id: "REPLACE_WITH_YOUR_KEY"
  aws-secret-access-key: "REPLACE_WITH_YOUR_SECRET"
  
  # Azure Credentials
  azure-tenant-id: "REPLACE_WITH_YOUR_TENANT_ID"
  azure-client-id: "REPLACE_WITH_YOUR_CLIENT_ID"
  azure-client-secret: "REPLACE_WITH_YOUR_CLIENT_SECRET"
  
  # GCP Service Account
  gcp-service-account-key: |
    {
      "type": "service_account",
      "project_id": "your-project-id"
    }
EOF
    
    log_warn "⚠️  Edit /tmp/offgridflow-secrets.yaml to add your actual credentials"
    log_info "Opening editor..."
    ${EDITOR:-vi} /tmp/offgridflow-secrets.yaml
    
    # Seal the secrets
    log_info "Sealing secrets..."
    kubeseal --format yaml \
        --cert infra/k8s/sealed-secrets-public-cert.pem \
        < /tmp/offgridflow-secrets.yaml \
        > infra/k8s/sealed-secrets.yaml
    
    # Cleanup temporary file
    rm -f /tmp/offgridflow-secrets.yaml
    
    log_info "✅ Sealed secrets created: infra/k8s/sealed-secrets.yaml"
    log_info "✅ This file is safe to commit to git"
}

# Apply sealed secrets
apply_sealed_secrets() {
    log_info "Applying sealed secrets to cluster..."
    
    # Create namespace if it doesn't exist
    kubectl create namespace offgridflow --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply sealed secrets
    kubectl apply -f infra/k8s/sealed-secrets.yaml
    
    log_info "✅ Sealed secrets applied"
    
    # Verify secret was created
    log_info "Verifying secret creation..."
    kubectl get secret offgridflow-secrets -n offgridflow &> /dev/null
    
    if [ $? -eq 0 ]; then
        log_info "✅ Secret 'offgridflow-secrets' created successfully"
    else
        log_error "Failed to create secret"
        exit 1
    fi
}

# Create rotation script
create_rotation_script() {
    log_info "Creating secret rotation script..."
    
    cat > scripts/rotate-secrets.sh <<'EOF'
#!/bin/bash
# Rotate secrets in Sealed Secrets

set -euo pipefail

SECRET_NAME=${1:-offgridflow-secrets}
NAMESPACE=${2:-offgridflow}

echo "Rotating secrets for ${SECRET_NAME} in namespace ${NAMESPACE}"

# Extract current secret
kubectl get secret ${SECRET_NAME} -n ${NAMESPACE} -o json > /tmp/current-secret.json

# Generate new values (example for JWT secret)
NEW_JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')

# Update secret
kubectl patch secret ${SECRET_NAME} -n ${NAMESPACE} \
  --type='json' \
  -p='[{"op": "replace", "path": "/data/jwt-secret", "value": "'$(echo -n $NEW_JWT_SECRET | base64)'"}]'

echo "✅ Secret rotated successfully"
echo "⚠️  Remember to restart pods to use new secret"
EOF
    
    chmod +x scripts/rotate-secrets.sh
    
    log_info "✅ Secret rotation script created: scripts/rotate-secrets.sh"
}

# Print usage instructions
print_usage() {
    cat <<EOF

${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}
${GREEN}Sealed Secrets Setup Complete!${NC}
${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}

${YELLOW}Important Files Created:${NC}
  📄 infra/k8s/sealed-secrets-public-cert.pem (COMMIT THIS)
  📄 infra/k8s/sealed-secrets.yaml (COMMIT THIS)
  📄 scripts/rotate-secrets.sh (Secret rotation tool)

${YELLOW}How to Update Secrets:${NC}
  1. Edit your plain secrets: ${GREEN}kubectl create secret generic offgridflow-secrets --dry-run=client -o yaml > /tmp/secrets.yaml${NC}
  2. Seal them: ${GREEN}kubeseal --format yaml --cert infra/k8s/sealed-secrets-public-cert.pem < /tmp/secrets.yaml > infra/k8s/sealed-secrets.yaml${NC}
  3. Apply: ${GREEN}kubectl apply -f infra/k8s/sealed-secrets.yaml${NC}
  4. Commit the sealed version to git

${YELLOW}Secret Rotation:${NC}
  ${GREEN}./scripts/rotate-secrets.sh${NC}
  Then restart pods: ${GREEN}kubectl rollout restart deployment/offgridflow-api -n offgridflow${NC}

${YELLOW}Backup Your Sealing Key:${NC}
  ${GREEN}kubectl get secret -n kube-system sealed-secrets-key -o yaml > sealed-secrets-master-key-backup.yaml${NC}
  ${RED}Store this backup in a SECURE location (NOT in git!)${NC}

${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}

EOF
}

# Main execution
main() {
    log_info "OffGridFlow Sealed Secrets Setup"
    log_info "================================="
    
    check_prerequisites
    install_controller
    get_public_key
    create_sealed_secrets
    apply_sealed_secrets
    create_rotation_script
    print_usage
}

main "$@"
