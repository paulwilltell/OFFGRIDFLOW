terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "offgridflow-terraform-state"
    key            = "production/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "offgridflow-terraform-locks"
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project     = "OffGridFlow"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# VPC and Networking
module "vpc" {
  source = "./modules/vpc"

  environment         = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  private_subnets    = var.private_subnets
  public_subnets     = var.public_subnets
}

# Database (RDS PostgreSQL)
module "db" {
  source = "./modules/db"

  environment          = var.environment
  vpc_id              = module.vpc.vpc_id
  private_subnet_ids  = module.vpc.private_subnet_ids
  db_instance_class   = var.db_instance_class
  db_allocated_storage = var.db_allocated_storage
  db_name             = var.db_name
  db_username         = var.db_username
  db_password         = var.db_password
  backup_retention_days = var.backup_retention_days
  multi_az            = var.multi_az
}

# Storage (S3)
module "storage" {
  source = "./modules/storage"

  environment    = var.environment
  bucket_prefix  = var.bucket_prefix
  enable_versioning = true
  lifecycle_rules = var.storage_lifecycle_rules
}

# Message Queue (SQS/SNS)
module "queue" {
  source = "./modules/queue"

  environment = var.environment
  queue_names = var.queue_names
}

# API Infrastructure (ECS Fargate)
module "api" {
  source = "./modules/api"

  environment        = var.environment
  vpc_id            = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  public_subnet_ids  = module.vpc.public_subnet_ids
  
  container_image   = var.api_container_image
  container_port    = var.api_container_port
  cpu              = var.api_cpu
  memory           = var.api_memory
  desired_count    = var.api_desired_count
  
  db_host          = module.db.endpoint
  db_name          = var.db_name
  db_username      = var.db_username
  db_password      = var.db_password
  
  storage_bucket   = module.storage.bucket_name
  queue_url        = module.queue.default_queue_url
}

# Redis Cache (ElastiCache)
module "cache" {
  source = "./modules/cache"

  environment        = var.environment
  vpc_id            = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  node_type         = var.redis_node_type
  num_cache_nodes   = var.redis_num_nodes
}

# Outputs
output "api_endpoint" {
  description = "API load balancer endpoint"
  value       = module.api.load_balancer_dns
}

output "database_endpoint" {
  description = "Database endpoint"
  value       = module.db.endpoint
  sensitive   = true
}

output "storage_bucket" {
  description = "S3 storage bucket name"
  value       = module.storage.bucket_name
}

output "redis_endpoint" {
  description = "Redis cache endpoint"
  value       = module.cache.endpoint
  sensitive   = true
}
