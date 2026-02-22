---
id: infra-s3-bucket-exposure
title: S3 Bucket Public Exposure and Misconfiguration
severity: critical
tags: [infrastructure, aws, s3, cloud-storage, public-access, data-exposure]
taxonomy: security/infrastructure/s3
references:
  - https://owasp.org/www-project-cloud-native-application-security-top-10/
  - https://docs.aws.amazon.com/AmazonS3/latest/userguide/security-best-practices.html
---

# S3 Bucket Public Exposure and Misconfiguration

## Description

Misconfigured S3 buckets have caused some of the largest data breaches in history (Capital One, Twitch, GoDaddy). Public access, overly permissive bucket policies, missing encryption, and disabled access logging are common mistakes. Similar risks apply to GCS, Azure Blob Storage, and other cloud object storage.

## Vulnerable Pattern

```json
// BAD — S3 bucket policy allowing public read
{
  "Statement": [{
    "Principal": "*",
    "Action": "s3:GetObject",
    "Resource": "arn:aws:s3:::my-company-data/*"
    // No condition — anyone on internet can read all objects
  }]
}
```

```terraform
# BAD — Terraform: public ACL on bucket
resource "aws_s3_bucket" "data" {
  bucket = "my-company-sensitive-data"
}
resource "aws_s3_bucket_acl" "data" {
  bucket = aws_s3_bucket.data.id
  acl    = "public-read"  # entire bucket publicly readable!
}

# BAD — public access block disabled
resource "aws_s3_bucket_public_access_block" "data" {
  bucket                  = aws_s3_bucket.data.id
  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}
```

## Secure Pattern

```terraform
# GOOD — all public access blocked, encryption enforced, versioning enabled
resource "aws_s3_bucket" "data" {
  bucket = "my-company-data"
}
resource "aws_s3_bucket_public_access_block" "data" {
  bucket                  = aws_s3_bucket.data.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
resource "aws_s3_bucket_server_side_encryption_configuration" "data" {
  bucket = aws_s3_bucket.data.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"  # KMS encryption
    }
  }
}
resource "aws_s3_bucket_versioning" "data" {
  bucket = aws_s3_bucket.data.id
  versioning_configuration { status = "Enabled" }
}
resource "aws_s3_bucket_logging" "data" {
  bucket        = aws_s3_bucket.data.id
  target_bucket = aws_s3_bucket.logs.id
  target_prefix = "s3-access-logs/"
}
```

## Checks to Generate

- Grep Terraform/CloudFormation for `acl = "public-read"` or `acl = "public-read-write"`.
- Flag `block_public_acls = false` or `block_public_policy = false` in S3 public access block.
- Grep for `"Principal": "*"` in S3 bucket policies without restrictive `Condition`.
- Check for missing `aws_s3_bucket_server_side_encryption_configuration` resource.
- Flag S3 buckets containing sensitive prefixes in name (`backup`, `prod`, `data`, `logs`) without encryption.
- Check for missing `aws_s3_bucket_logging` — access logs needed for incident response.
