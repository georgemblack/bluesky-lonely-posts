locals {
  intake_version = "0.2.2"
}

resource "aws_ecs_task_definition" "lonely-posts-intake" {
  family                   = "lonely-posts-intake"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  task_role_arn            = aws_iam_role.lonely_posts_service.arn
  execution_role_arn       = aws_iam_role.lonely_posts_execution.arn

  container_definitions = jsonencode([
    {
      name      = "intake"
      image     = "013249834781.dkr.ecr.us-east-2.amazonaws.com/lonely-posts:${local.intake_version}"
      essential = true
      command   = ["/intake"]
      environment = [
        {
          name  = "VALKEY_ADDRESS"
          value = data.aws_secretsmanager_secret_version.lonely_posts_cache_address.secret_string
        },
        {
          name  = "VALKEY_TLS_ENABLED"
          value = "true"
        },
      ]
      cpu    = 256
      memory = 512
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-region" = "us-east-2"
          "awslogs-group"  = aws_cloudwatch_log_stream.bluesky.name
          "awslogs-stream-prefix" : "lonely-posts-intake"
        }
      }
    },
  ])

  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }
}

resource "aws_ecs_service" "lonely-posts-intake" {
  name            = "intake"
  desired_count   = 1
  cluster         = aws_ecs_cluster.bluesky.id
  task_definition = aws_ecs_task_definition.lonely-posts-intake.arn

  network_configuration {
    subnets          = [aws_subnet.bluesky_subnet_2a.id, aws_subnet.bluesky_subnet_2b.id, aws_subnet.bluesky_subnet_2c.id]
    assign_public_ip = true
    security_groups  = [aws_security_group.lonely_posts_intake.id]
  }

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
}
