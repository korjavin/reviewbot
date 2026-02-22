---
id: infra-cloud-iam-misconfiguration
title: Cloud IAM Misconfiguration (AWS/GCP/Azure)
severity: critical
tags: [infrastructure, cloud, iam, aws, gcp, azure, privilege-escalation]
taxonomy: security/infrastructure/iam
references:
  - https://owasp.org/www-project-cloud-native-application-security-top-10/
  - https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html
---

# Cloud IAM Misconfiguration (AWS/GCP/Azure)

## Description

Overly permissive IAM policies are the most common cloud security issue. `*` resource and action wildcards, missing condition keys, and unnecessary admin roles create privilege escalation paths and allow lateral movement across cloud services. The principle of least privilege must be applied to all identities.

## Vulnerable Pattern

```json
// BAD — AWS IAM policy with wildcard resource and actions (admin equivalent)
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }]
}

// BAD — overly broad S3 access
{
  "Action": ["s3:*"],
  "Resource": ["arn:aws:s3:::*"]  // access to ALL buckets!
}
```

```yaml
# BAD — GCP service account with Editor role (too broad)
resource "google_project_iam_binding" "sa_binding" {
  role = "roles/editor"  # full project edit access
  members = ["serviceAccount:app@project.iam.gserviceaccount.com"]
}
```

## Secure Pattern

```json
// GOOD — least-privilege: specific actions on specific resources
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "s3:GetObject",
      "s3:PutObject"
    ],
    "Resource": "arn:aws:s3:::my-app-bucket/uploads/*",
    "Condition": {
      "StringEquals": {
        "aws:RequestedRegion": "us-east-1"
      }
    }
  }]
}
```

```yaml
# GOOD — GCP: minimal custom role
resource "google_project_iam_custom_role" "app_role" {
  role_id = "appRole"
  title   = "App Service Role"
  permissions = [
    "storage.objects.get",
    "storage.objects.create",
    "pubsub.topics.publish"
  ]
}
```

## Checks to Generate

- Grep IaC (Terraform, CloudFormation, Pulumi) for `Action: "*"` or `"*"` in `Action` array — flag as overly permissive.
- Flag `Resource: "*"` combined with `Effect: Allow` on any statement.
- Grep for `roles/editor`, `roles/owner`, `AdministratorAccess` assigned to application service accounts.
- Check for missing `Condition` keys on sensitive S3/GCS bucket policies.
- Flag IAM roles with `sts:AssumeRole` on `*` resource (privilege escalation path).
- Check for cross-account assume role policies without external ID condition.
