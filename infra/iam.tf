# IAM execution role and policy for Blue Report ECS services.
# The only permissions required are access to CloudWatch Logs and ECR, to deploy services.
resource "aws_iam_role" "lonely_posts_execution" {
  name = "lonely-posts-service-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect    = "Allow",
        Principal = { Service = "ecs-tasks.amazonaws.com" },
        Action    = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_policy" "lonely_posts_execution" {
  name        = "lonely-posts-service-execution"
  description = "Policy for Lonely Posts ECS services"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Resource = "*"
      },
      {
        Effect = "Allow",
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage"
        ],
        Resource = [
          "*"
        ]
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "execution" {
  role       = aws_iam_role.lonely_posts_execution.name
  policy_arn = aws_iam_policy.lonely_posts_execution.arn
}
