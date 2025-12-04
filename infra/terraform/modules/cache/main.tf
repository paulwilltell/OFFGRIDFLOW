resource "aws_elasticache_subnet_group" "main" {
  name       = "offgridflow-${var.environment}-cache-subnet-group"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name        = "offgridflow-${var.environment}-cache-subnet-group"
    Environment = var.environment
  }
}

resource "aws_security_group" "cache" {
  name        = "offgridflow-${var.environment}-cache-sg"
  description = "Security group for Redis ElastiCache"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
    description = "Redis from VPC"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = {
    Name        = "offgridflow-${var.environment}-cache-sg"
    Environment = var.environment
  }
}

resource "aws_elasticache_cluster" "main" {
  cluster_id           = "offgridflow-${var.environment}-redis"
  engine               = "redis"
  engine_version       = "7.0"
  node_type            = var.node_type
  num_cache_nodes      = var.num_cache_nodes
  parameter_group_name = "default.redis7"
  port                 = 6379

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = [aws_security_group.cache.id]

  snapshot_retention_limit = var.environment == "production" ? 5 : 0
  snapshot_window          = "03:00-05:00"
  maintenance_window       = "mon:05:00-mon:07:00"

  tags = {
    Name        = "offgridflow-${var.environment}-redis"
    Environment = var.environment
  }
}
