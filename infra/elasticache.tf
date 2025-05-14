# Previously, we used a serverless Valkey cache.
# To save money, we are now using a t4g.micro Valkey cache (with reserved nodes).
# Terraform does not yet support a standard ElastiCache cluster with Valkey, so this is done manually via the console.
# TODO: Manage ElastiCache with Terraform, once the use case is supported.

resource "aws_elasticache_subnet_group" "bluesky" {
  name       = "bluesky-subnets"
  subnet_ids = [aws_subnet.bluesky_subnet_2a.id, aws_subnet.bluesky_subnet_2b.id, aws_subnet.bluesky_subnet_2c.id]
}
