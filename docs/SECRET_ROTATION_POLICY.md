# OffGridFlow Secret Rotation Policy
**Version**: 1.0.0  
**Last Updated**: December 4, 2025  
**Owner**: Security Team

## Overview
This policy defines procedures for rotating all secrets used by OffGridFlow
to minimize risk from compromised credentials.

## Rotation Schedule

### CRITICAL Secrets (Rotate every 90 days)
- **JWT Signing Secret** (`OFFGRIDFLOW_JWT_SECRET`)
- **Database Master Password** (`DB_PASSWORD`)
- **Stripe Secret Key** (`STRIPE_SECRET_KEY`)
- **AWS Access Keys** (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)

### HIGH Priority (Rotate every 180 days)
- **API Keys** (per-tenant keys)
- **Azure Client Secrets**
- **GCP Service Account Keys**
- **Email Service Keys**

### MEDIUM Priority (Rotate every 365 days)
- **Monitoring/Logging API Keys**
- **Third-party Integration Keys**

### Event-Driven Rotation (Immediate)
- **Employee Termination**: Rotate all secrets accessed by terminated employee
- **Security Incident**: Rotate all potentially compromised secrets
- **Detected Exposure**: Rotate immediately (e.g., committed to git)

## Rotation Procedures

### JWT Secret Rotation

**Preparation**:
1. Generate new secret: `openssl rand -base64 48`
2. Store in secrets manager (AWS Secrets Manager / HashiCorp Vault)
3. Update Kubernetes secrets in staging

**Execution**:
```bash
# Update Kubernetes secret
kubectl create secret generic offgridflow-jwt-new \
  --from-literal=jwt-secret=NEW_SECRET \
  --namespace=offgridflow

# Rolling update with both secrets active
kubectl set env deployment/offgridflow-api \
  JWT_SECRET_NEW=NEW_SECRET

# Monitor for 24 hours
# If no issues, make new secret primary
kubectl set env deployment/offgridflow-api \
  JWT_SECRET=NEW_SECRET

# Remove old secret after 7 days (grace period)
```

**Verification**:
- Monitor login success rates
- Check error logs for auth failures
- Verify no session invalidation

### Database Password Rotation

**Using AWS RDS**:
```bash
# Create new password
NEW_PASSWORD=$(openssl rand -base64 32)

# Modify DB credentials
aws rds modify-db-instance \
  --db-instance-identifier offgridflow-prod \
  --master-user-password "$NEW_PASSWORD" \
  --apply-immediately

# Update application secrets
kubectl set env deployment/offgridflow-api \
  DB_PASSWORD="$NEW_PASSWORD"

# Verify connectivity
kubectl exec -it offgridflow-api-xxx -- \
  psql -h $DB_HOST -U offgridflow -c "SELECT 1"
```

### Stripe Secret Key Rotation

**Process**:
1. Generate new restricted key in Stripe Dashboard
2. Update Kubernetes secrets
3. Deploy with new key
4. Verify webhooks still work
5. Revoke old key after 48 hours

### API Key Rotation (Per-Tenant)

**Automated via API**:
```bash
# Generate new key for tenant
NEW_KEY=$(curl -X POST https://api.offgridflow.com/api/v1/keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"label": "Rotated Key", "expires_in_days": 90}' \
  | jq -r '.key')

# Notify customer
# Revoke old key after grace period
```

## Automation

### Automated Rotation Tool

```go
// cmd/rotate-secrets/main.go

package main

import (
    "context"
    "fmt"
    "time"
)

type SecretRotator struct {
    secretsManager SecretsManager
    k8sClient      KubernetesClient
}

func (r *SecretRotator) RotateJWTSecret(ctx context.Context) error {
    // 1. Generate new secret
    newSecret := generateRandomSecret(48)
    
    // 2. Store in secrets manager
    if err := r.secretsManager.Store(ctx, "jwt-secret-new", newSecret); err != nil {
        return fmt.Errorf("store secret: %w", err)
    }
    
    // 3. Update Kubernetes deployment
    if err := r.k8sClient.UpdateEnv(ctx, "offgridflow-api", map[string]string{
        "JWT_SECRET_NEW": newSecret,
    }); err != nil {
        return fmt.Errorf("update k8s: %w", err)
    }
    
    // 4. Wait for rollout
    if err := r.k8sClient.WaitForRollout(ctx, "offgridflow-api", 5*time.Minute); err != nil {
        return fmt.Errorf("rollout: %w", err)
    }
    
    // 5. Promote new secret to primary
    time.Sleep(24 * time.Hour) // Grace period
    if err := r.k8sClient.UpdateEnv(ctx, "offgridflow-api", map[string]string{
        "JWT_SECRET": newSecret,
    }); err != nil {
        return fmt.Errorf("promote secret: %w", err)
    }
    
    return nil
}
```

### Cron Schedule

```yaml
# k8s/cronjobs/secret-rotation.yaml

apiVersion: batch/v1
kind: CronJob
metadata:
  name: rotate-jwt-secret
  namespace: offgridflow
spec:
  schedule: "0 2 1 */3 *"  # 2 AM on 1st day every 3 months
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: rotator
            image: offgridflow-secret-rotator:latest
            command: ["./rotate-secrets", "--type=jwt"]
            env:
            - name: SECRET_MANAGER
              value: "aws-secrets-manager"
```

## Incident Response

### Suspected Secret Compromise

1. **Immediate (< 1 hour)**:
   - Revoke compromised secret
   - Generate and deploy new secret
   - Force logout all sessions (if JWT compromised)
   - Block compromised API keys

2. **Investigation (< 24 hours)**:
   - Review access logs for unauthorized usage
   - Identify scope of compromise
   - Assess data exposure

3. **Remediation (< 7 days)**:
   - Rotate all potentially related secrets
   - Implement additional monitoring
   - Update security procedures

## Monitoring & Alerts

### Secret Expiry Alerts
```yaml
# prometheus-alerts.yaml

- alert: SecretExpiringIn30Days
  expr: (secret_rotation_due_days < 30)
  annotations:
    summary: "Secret {{ $labels.secret_name }} expires in {{ $value }} days"
```

### Rotation Failure Alerts
```yaml
- alert: SecretRotationFailed
  expr: (secret_rotation_status{status="failed"} == 1)
  annotations:
    summary: "Secret rotation failed for {{ $labels.secret_name }}"
```

## Documentation Requirements

Every secret must have:
- **Owner**: Team responsible
- **Last Rotated**: Date of last rotation
- **Next Rotation**: Scheduled rotation date
- **Rotation Procedure**: Link to runbook
- **Emergency Contact**: On-call engineer

## Compliance

This policy supports:
- **SOC 2 Type II**: Access control requirements
- **ISO 27001**: Key management requirements
- **PCI DSS**: Cryptographic key management (if handling cards)

## Review Schedule

This policy is reviewed:
- Annually on January 1st
- After any security incident
- When adding new critical secrets
```
