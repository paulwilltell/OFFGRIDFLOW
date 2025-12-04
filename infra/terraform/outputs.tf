# Terraform Outputs
# These outputs can be used to configure Kubernetes secrets

output "database_url" {
  description = "PostgreSQL connection string for applications"
  value       = "postgresql://${var.db_username}:${var.db_password}@${module.db.endpoint}/${var.db_name}?sslmode=require"
  sensitive   = true
}

output "database_host" {
  description = "Database host endpoint"
  value       = module.db.endpoint
}

output "database_name" {
  description = "Database name"
  value       = var.db_name
}

output "redis_url" {
  description = "Redis connection string"
  value       = "redis://${module.cache.endpoint}:6379/0"
  sensitive   = true
}

output "redis_host" {
  description = "Redis cache endpoint"
  value       = module.cache.endpoint
}

output "storage_bucket_name" {
  description = "S3 storage bucket name"
  value       = module.storage.bucket_name
}

output "storage_bucket_arn" {
  description = "S3 storage bucket ARN"
  value       = module.storage.bucket_arn
}

output "api_load_balancer_dns" {
  description = "API load balancer DNS name"
  value       = module.api.load_balancer_dns
}

output "api_load_balancer_zone_id" {
  description = "API load balancer zone ID for Route53"
  value       = module.api.load_balancer_zone_id
}

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnet_ids
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnet_ids
}

output "queue_urls" {
  description = "SQS queue URLs"
  value       = module.queue.queue_urls
  sensitive   = false
}

output "queue_arns" {
  description = "SQS queue ARNs"
  value       = module.queue.queue_arns
  sensitive   = false
}

# Script to create Kubernetes secrets from Terraform outputs
output "kubectl_secret_command" {
  description = "Command to create Kubernetes secret from Terraform outputs"
  value       = <<-EOT
    # Run this command to create Kubernetes secrets from Terraform outputs:
    kubectl create secret generic offgridflow-secrets \
      --from-literal=database-url="${self.database_url.value}" \
      --from-literal=redis-url="${self.redis_url.value}" \
      --from-literal=jwt-secret="$(openssl rand -base64 32)" \
      --dry-run=client -o yaml | kubectl apply -f -
  EOT
  sensitive   = true
}
