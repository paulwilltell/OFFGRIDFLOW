variable "environment" {
  description = "Environment name"
  type        = string
}

variable "queue_names" {
  description = "Names of queues to create"
  type        = list(string)
  default     = ["default", "emissions-processing", "connectors", "reports"]
}
