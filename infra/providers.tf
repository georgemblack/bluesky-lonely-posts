terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.77.0"
    }
  }

  backend "s3" {
    bucket = "bluesky-terraform"
    key    = "state"
    region = "us-east-2"
  }
}

provider "aws" {
  region = "us-east-2"
}
