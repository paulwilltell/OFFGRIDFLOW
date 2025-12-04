variable "region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name (dev, staging, production)"
  type        = string
  default     = "production"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones for deployment"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b", "us-west-2c"]
}

variable "private_subnets" {
  description = "Private subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnets" {
  description = "Public subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

# Database variables
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "db_allocated_storage" {
  description = "Allocated storage for RDS in GB"
  type        = number
  default     = 100
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "offgridflow"
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "offgridflow_admin"
  sensitive   = true
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
}

variable "multi_az" {
  description = "Enable multi-AZ deployment for RDS"
  type        = bool
  default     = true
}

# Storage variables
variable "bucket_prefix" {
  description = "Prefix for S3 bucket names"
  type        = string
  default     = "offgridflow"
}

variable "storage_lifecycle_rules" {
  description = "S3 lifecycle rules"
  type = list(object({
    id                            = string
    enabled                       = bool
    transition_days               = number
    transition_storage_class      = string
    expiration_days              = number
  }))
  default = [
    {
      id                       = "archive-old-data"
      enabled                  = true
      transition_days          = 90
      transition_storage_class = "GLACIER"
      expiration_days         = 365
    }
  ]
}

# Queue variables
variable "queue_names" {
  description = "Names of SQS queues to create"
  type        = list(string)
  default     = ["default", "emissions-processing", "connectors", "reports"]
}

# API variables
variable "api_container_image" {
  description = "Docker image for API"
  type        = string
  default     = "ghcr.io/example/offgridflow-api:latest"
}

variable "api_container_port" {
  description = "Port exposed by API container"
  type        = number
  default     = 8080
}

variable "api_cpu" {
  description = "CPU units for API task"
  type        = number
  default     = 512
}

variable "api_memory" {
  description = "Memory in MB for API task"
  type        = number
  default     = 1024
}

variable "api_desired_count" {
  description = "Desired number of API tasks"
  type        = number
  default     = 2
}

# Redis variables
variable "redis_node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "redis_num_nodes" {
  description = "Number of cache nodes"
  type        = number
  default     = 1
}
