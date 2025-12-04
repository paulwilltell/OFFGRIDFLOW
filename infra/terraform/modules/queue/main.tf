# SQS Queues for different job types
resource "aws_sqs_queue" "main" {
  for_each = toset(var.queue_names)

  name                       = "offgridflow-${var.environment}-${each.key}"
  delay_seconds              = 0
  max_message_size           = 262144
  message_retention_seconds  = 1209600 # 14 days
  receive_wait_time_seconds  = 10      # Long polling
  visibility_timeout_seconds = 300

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq[each.key].arn
    maxReceiveCount     = 3
  })

  tags = {
    Name        = "offgridflow-${var.environment}-${each.key}"
    Environment = var.environment
    QueueType   = each.key
  }
}

# Dead Letter Queues
resource "aws_sqs_queue" "dlq" {
  for_each = toset(var.queue_names)

  name                      = "offgridflow-${var.environment}-${each.key}-dlq"
  message_retention_seconds = 1209600 # 14 days

  tags = {
    Name        = "offgridflow-${var.environment}-${each.key}-dlq"
    Environment = var.environment
    QueueType   = "dlq"
  }
}

# SNS Topic for notifications (optional)
resource "aws_sns_topic" "notifications" {
  name = "offgridflow-${var.environment}-notifications"

  tags = {
    Name        = "offgridflow-${var.environment}-notifications"
    Environment = var.environment
  }
}

