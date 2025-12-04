variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs"
  type        = list(string)
}

variable "public_subnet_ids" {
  description = "Public subnet IDs"
  type        = list(string)
}

variable "container_image" {
  description = "Docker image for API"
  type        = string
}

variable "container_port" {
  description = "Port exposed by container"
  type        = number
  default     = 8080
}

variable "cpu" {
  description = "CPU units for task"
  type        = number
  default     = 512
}

variable "memory" {
  description = "Memory in MB for task"
  type        = number
  default     = 1024
}

variable "desired_count" {
  description = "Desired number of tasks"
  type        = number
  default     = 2
}

variable "db_host" {
  description = "Database host"
  type        = string
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_username" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "storage_bucket" {
  description = "S3 storage bucket name"
  type        = string
}

variable "queue_url" {
  description = "SQS queue URL"
  type        = string
}
