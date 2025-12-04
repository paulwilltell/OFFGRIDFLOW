variable "environment" {
  description = "Environment name"
  type        = string
}

variable "bucket_prefix" {
  description = "S3 bucket name prefix"
  type        = string
}

variable "enable_versioning" {
  description = "Enable S3 versioning"
  type        = bool
  default     = true
}

variable "lifecycle_rules" {
  description = "S3 lifecycle rules"
  type        = any
  default     = []
}
