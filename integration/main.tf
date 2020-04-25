terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "kvrhdn"

    workspaces {
      name = "tfe-run_integration"
    }
  }
}

provider "aws" {
  version = "~> 2.57"
  region  = "eu-central-1"
}

locals {
  bucket_name = "tfe-run_integration"
  tags = {
    project   = "tfe-run"
    terraform = true
  }
}

variable "run_number" {
  description = "Run number that will be published on the bucket"
  type        = number
}

resource "aws_s3_bucket" "main" {
  bucket = local.bucket_name
  acl    = "public-read"

  policy = templatefile("public_access_s3_policy.json", {
    bucket_name = local.bucket_name
  })

  force_destroy = true

  website {
    index_document = "index.txt"
    error_document = "error.html"
  }

  tags = local.tags
}

resource "aws_s3_bucket_object" "index" {
  bucket = aws_s3_bucket.main.id
  key    = "index.txt"
  content = templatefile("index.txt", {
      run_number = var.run_number
  })

  tags = local.tags
}

resource "aws_s3_bucket_object" "error" {
  bucket = aws_s3_bucket.main.id
  key    = "error.html"
  content = file("error.html")

  tags = local.tags
}

output "endpoint" {
  value = aws_s3_bucket.main.website_endpoint
}
