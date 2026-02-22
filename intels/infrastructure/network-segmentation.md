---
id: infra-network-segmentation
title: Missing Network Segmentation and Firewall Rules
severity: high
tags: [infrastructure, network-segmentation, firewall, vpc, zero-trust]
taxonomy: security/infrastructure/network
references:
  - https://owasp.org/www-project-cloud-native-application-security-top-10/
  - https://cheatsheetseries.owasp.org/cheatsheets/Infrastructure_Security_Cheat_Sheet.html
---

# Missing Network Segmentation and Firewall Rules

## Description

Flat network architectures where all services can communicate with each other dramatically increase the blast radius of a breach. A compromised frontend should not be able to directly reach internal databases or management interfaces. Defense-in-depth requires network-level segmentation as a layer below application-level access controls.

## Vulnerable Pattern

```terraform
# BAD — security group allows all traffic from internet to database
resource "aws_security_group_rule" "db_ingress" {
  type        = "ingress"
  from_port   = 5432
  to_port     = 5432
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]  # database open to internet!
  security_group_id = aws_security_group.db.id
}

# BAD — overly permissive egress (allows data exfiltration)
resource "aws_security_group_rule" "all_egress" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]  # application can reach any internet destination
}
```

```yaml
# BAD — Kubernetes: no NetworkPolicy (default: all pods can reach all pods)
# Without NetworkPolicy, a compromised pod in namespace A can reach DB in namespace B
```

## Secure Pattern

```terraform
# GOOD — database only accessible from application tier security group
resource "aws_security_group_rule" "db_from_app" {
  type                     = "ingress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.app.id  # only from app SG
  security_group_id        = aws_security_group.db.id
}

# GOOD — restrictive egress for application tier
resource "aws_security_group_rule" "app_egress_https" {
  type        = "egress"
  from_port   = 443
  to_port     = 443
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]  # HTTPS only
  security_group_id = aws_security_group.app.id
}
```

```yaml
# GOOD — Kubernetes NetworkPolicy: default deny, explicit allow
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny
  namespace: production
spec:
  podSelector: {}  # applies to all pods
  policyTypes: ["Ingress", "Egress"]
  # No rules = deny all ingress and egress
---
kind: NetworkPolicy
metadata:
  name: allow-db-from-api
spec:
  podSelector:
    matchLabels: { app: database }
  ingress:
  - from:
    - podSelector:
        matchLabels: { app: api }
    ports:
    - port: 5432
```

## Checks to Generate

- Grep Terraform for `cidr_blocks = ["0.0.0.0/0"]` on database/cache security group ingress rules.
- Flag absence of Kubernetes `NetworkPolicy` — default is allow-all between pods.
- Check VPC subnet configuration — databases should be in private subnets with no internet route.
- Grep for security groups allowing `0.0.0.0/0` on ports 22 (SSH), 3389 (RDP), 5432 (PG), 3306 (MySQL).
- Flag missing VPC Flow Logs — required for network-level incident investigation.
- Check for WAF (Web Application Firewall) absence on internet-facing load balancers.
