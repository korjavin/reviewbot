---
id: dns-subdomain-takeover
title: Subdomain Takeover via Dangling DNS Records
severity: high
tags: [dns, subdomain-takeover, cname, cloud, azure, github-pages, heroku]
taxonomy: security/dns/subdomain-takeover
references:
  - https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/02-Configuration_and_Deployment_Management_Testing/10-Test_for_Subdomain_Takeover
  - https://github.com/EdOverflow/can-i-take-over-xyz
---

# Subdomain Takeover via Dangling DNS Records

## Description

Subdomain takeover occurs when a DNS CNAME record points to a cloud service endpoint (GitHub Pages, Azure App Service, Heroku, S3, Fastly, etc.) that no longer exists. An attacker can claim the abandoned cloud resource and host arbitrary content under the victim's trusted subdomain — enabling phishing, cookie theft (if `Domain=.example.com` cookies), and XSS via `*.example.com` CSP trust.

High-risk providers: GitHub Pages, Azure (azurewebsites.net), Heroku, Fastly, AWS Elastic Beanstalk, Shopify, Ghost.

## Vulnerable Pattern

```
# Company previously used GitHub Pages for staging:
staging.example.com. CNAME examplecorp.github.io.

# The GitHub Pages site was deleted/renamed — but DNS record remains
# examplecorp.github.io returns 404 → CNAME is dangling

# Attacker creates GitHub account "examplecorp", creates repo with GitHub Pages
# → staging.example.com now serves attacker's content!

# Same pattern with other services:
api-old.example.com.   CNAME  old-app.azurewebsites.net    (Azure deleted)
app-staging.example.com. CNAME old-app.herokuapp.com       (Heroku deleted)
assets.example.com.    CNAME  my-bucket.s3.amazonaws.com   (S3 bucket deleted)
```

```yaml
# BAD — DNS zone not audited when services are decommissioned
# Infrastructure teardown checklist lacks: "Remove all DNS records pointing to this resource"
```

## Secure Pattern

```yaml
# GOOD — DNS audit process on decommission
# When removing any cloud resource (GitHub Pages, Heroku app, Azure app, S3 bucket, CDN):
# 1. Remove DNS CNAME/A record FIRST
# 2. Wait for TTL to expire (e.g., 1 hour if TTL=3600)
# 3. Then delete the cloud resource
# Order matters — deleting resource first creates the takeover window!
```

```bash
# GOOD — automated subdomain takeover scanning in CI
# Install: pip install dnstwist subjack  OR  use nuclei templates

# nuclei scan for dangling CNAMEs
nuclei -t dns/subdomain-takeover/ -l subdomains.txt

# subjack for common cloud providers
subjack -w subdomains.txt -t 20 -timeout 30 -o results.txt -ssl

# Manual check: resolve CNAME target
dig staging.example.com CNAME
# If target returns NXDOMAIN or provider-specific "not found" page → vulnerable
```

```python
# GOOD — automated check integrated into infrastructure pipeline
import dns.resolver

TAKEOVER_SIGNATURES = {
    "github.io": "There isn't a GitHub Pages site here",
    "azurewebsites.net": "404 Web Site not found",
    "herokuapp.com": "no such app",
    "s3.amazonaws.com": "NoSuchBucket",
    "fastly.net": "Fastly error: unknown domain",
}

def check_dangling_cname(subdomain: str) -> bool:
    try:
        answers = dns.resolver.resolve(subdomain, "CNAME")
        cname_target = str(answers[0].target)
        for provider, signature in TAKEOVER_SIGNATURES.items():
            if provider in cname_target:
                response = requests.get(f"https://{subdomain}", timeout=5)
                if signature in response.text:
                    return True  # dangling — takeover possible
    except dns.resolver.NXDOMAIN:
        pass
    return False
```

## Checks to Generate

- Scan all DNS CNAME records pointing to known cloud provider domains (github.io, azurewebsites.net, herokuapp.com, s3.amazonaws.com, cloudfront.net, fastly.net).
- Verify each CNAME target resolves and returns valid content — flag 404s or provider-specific error pages.
- Check infrastructure teardown runbooks/checklists for DNS cleanup step.
- Grep Terraform/CloudFormation for `aws_route53_record` resources without lifecycle management.
- Flag CI/CD pipelines that delete cloud resources without corresponding DNS record removal.
- Run `subjack` or `nuclei` subdomain takeover templates as part of security CI.
