resource "aws_ecr_repository" "lonely_posts" {
  name                 = "lonely-posts"
  image_tag_mutability = "MUTABLE"
}

resource "aws_cloudwatch_log_group" "bluesky" {
  name              = "bluesky"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_stream" "bluesky" {
  name           = "bluesky"
  log_group_name = aws_cloudwatch_log_group.bluesky.name
}

resource "aws_ecs_cluster" "bluesky" {
  name = "bluesky"
}

resource "aws_ecs_cluster_capacity_providers" "bluesky" {
  cluster_name = aws_ecs_cluster.bluesky.name
  capacity_providers = [
    "FARGATE",
    "FARGATE_SPOT",
  ]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
  }
}
