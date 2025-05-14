data "aws_secretsmanager_secret_version" "lonely_posts_cache_address" {
  secret_id = "bluesky/lonely-posts/cache-address"
}
