resource "aws_vpc" "bluesky" {
  cidr_block                       = "10.0.0.0/16"
  enable_dns_support               = true
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = true

  tags = {
    "Name" = "bluesky-vpc"
  }
}

resource "aws_subnet" "bluesky_subnet_2a" {
  vpc_id            = aws_vpc.bluesky.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-2a"

  tags = {
    Name = "bluesky-subnet-2a"
  }
}

resource "aws_subnet" "bluesky_subnet_2b" {
  vpc_id            = aws_vpc.bluesky.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-east-2b"

  tags = {
    Name = "bluesky-subnet-2b"
  }
}

resource "aws_subnet" "bluesky_subnet_2c" {
  vpc_id            = aws_vpc.bluesky.id
  cidr_block        = "10.0.3.0/24"
  availability_zone = "us-east-2c"

  tags = {
    Name = "bluesky-subnet-2c"
  }
}

resource "aws_internet_gateway" "bluesky" {
  vpc_id = aws_vpc.bluesky.id

  tags = {
    Name = "bluesky-igw"
  }
}

resource "aws_route_table" "bluesky" {
  vpc_id = aws_vpc.bluesky.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.bluesky.id
  }

  tags = {
    Name = "bluesky-rt"
  }
}

resource "aws_main_route_table_association" "bluesky" {
  vpc_id         = aws_vpc.bluesky.id
  route_table_id = aws_route_table.bluesky.id
}

resource "aws_security_group" "lonely_posts_intake" {
  vpc_id = aws_vpc.bluesky.id

  # Egress rule to allow all outbound traffic
  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "lonely_posts_server" {
  vpc_id = aws_vpc.bluesky.id

  # Egress rule to allow all outbound traffic
  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Ingress rule to allow all traffic from Cloudflare IPv4 IPs
  ingress {
    description = "Allow all inbound traffic from Cloudflare IPv4 IPs"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [
      "173.245.48.0/20",
      "103.21.244.0/22",
      "103.22.200.0/22",
      "103.31.4.0/22",
      "141.101.64.0/18",
      "108.162.192.0/18",
      "190.93.240.0/20",
      "188.114.96.0/20",
      "197.234.240.0/22",
      "198.41.128.0/17",
      "162.158.0.0/15",
      "104.16.0.0/13",
      "104.24.0.0/14",
      "172.64.0.0/13",
      "131.0.72.0/22"
    ]
  }

  # Ingress rule to allow all traffic from Cloudflare IPv6 IPs
  ingress {
    description = "Allow all inbound traffic from Cloudflare IPv6 IPs"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    ipv6_cidr_blocks = [
      "2400:cb00::/32",
      "2606:4700::/32",
      "2803:f800::/32",
      "2405:b500::/32",
      "2405:8100::/32",
      "2a06:98c0::/29",
      "2c0f:f248::/32"
    ]
  }

  # Ingress rule to allow all traffic from anywhere
  # TODO: Remove after testing
  ingress {
    description      = "Allow all inbound traffic from anywhere"
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_security_group" "lonely_posts_cache" {
  vpc_id = aws_vpc.bluesky.id

  # Ingress rule to allow all traffic from lonely_posts_server
  ingress {
    description     = "Allow all inbound traffic from lonely_posts_server"
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    security_groups = [aws_security_group.lonely_posts_server.id]
  }
}
