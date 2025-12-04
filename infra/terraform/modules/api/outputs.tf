output "load_balancer_dns" {
  description = "Load balancer DNS name"
  value       = aws_lb.api.dns_name
}

output "load_balancer_zone_id" {
  description = "Load balancer zone ID"
  value       = aws_lb.api.zone_id
}

output "load_balancer_arn" {
  description = "Load balancer ARN"
  value       = aws_lb.api.arn
}

output "ecs_cluster_id" {
  description = "ECS cluster ID"
  value       = aws_ecs_cluster.main.id
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = aws_ecs_service.api.name
}

output "target_group_arn" {
  description = "Target group ARN"
  value       = aws_lb_target_group.api.arn
}
