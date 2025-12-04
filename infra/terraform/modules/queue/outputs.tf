output "queue_urls" {
  description = "Map of queue names to URLs"
  value = {
    for k, v in aws_sqs_queue.main : k => v.url
  }
}

output "queue_arns" {
  description = "Map of queue names to ARNs"
  value = {
    for k, v in aws_sqs_queue.main : k => v.arn
  }
}

output "dlq_urls" {
  description = "Map of DLQ names to URLs"
  value = {
    for k, v in aws_sqs_queue.dlq : k => v.url
  }
}

output "dlq_arns" {
  description = "Map of DLQ names to ARNs"
  value = {
    for k, v in aws_sqs_queue.dlq : k => v.arn
  }
}

output "default_queue_url" {
  description = "Default queue URL"
  value       = aws_sqs_queue.main["default"].url
}

output "sns_topic_arn" {
  description = "SNS topic ARN for notifications"
  value       = aws_sns_topic.notifications.arn
}
