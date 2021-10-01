terraform {
  required_providers {
    aws = {
      source  = "hashicorp/awscc"
      version = "~> 0.1"
    }
  }
}

provider "awscc" {
  region = var.aws_region
}