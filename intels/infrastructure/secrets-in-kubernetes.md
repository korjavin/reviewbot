---
id: infra-kubernetes-secrets
title: Kubernetes Secrets — Insecure Storage and Access
severity: high
tags: [infrastructure, kubernetes, secrets, k8s, vault, encryption]
taxonomy: security/infrastructure/kubernetes-secrets
references:
  - https://owasp.org/www-project-kubernetes-top-ten/
  - https://kubernetes.io/docs/concepts/configuration/secret/
  - https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/
---

# Kubernetes Secrets — Insecure Storage and Access

## Description

Kubernetes Secrets are base64-encoded by default — not encrypted. Without encryption at rest, secrets are readable by anyone with etcd access. Additionally, secrets are often overly shared (mounted to all pods in a namespace), logged by default in some configurations, and accessible via `kubectl get secret` to broad RBAC roles.

## Vulnerable Pattern

```yaml
# BAD — secret in base64 (not encryption — trivially reversible)
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
data:
  DB_PASSWORD: U3VwZXJTZWNyZXQxMjMh  # just base64("SuperSecret123!") — not secure

# BAD — secret values hardcoded in Deployment manifest (committed to git)
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - env:
        - name: DB_PASSWORD
          value: "SuperSecret123!"  # plaintext in YAML!
```

```bash
# BAD — listing secrets with broad RBAC
kubectl get secret db-credentials -o yaml  # reveals base64 content to anyone with get/list
```

## Secure Pattern

```yaml
# GOOD — use external secrets operator with Vault/AWS Secrets Manager
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: db-credentials
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: ClusterSecretStore
  target:
    name: db-credentials  # creates K8s Secret from Vault
  data:
  - secretKey: DB_PASSWORD
    remoteRef:
      key: production/db
      property: password

# GOOD — restrict secret access to specific service accounts
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["db-credentials"]  # only specific secret
  verbs: ["get"]
```

```bash
# GOOD — enable encryption at rest for etcd
# apiserver configuration:
# --encryption-provider-config=/etc/kubernetes/encryption-config.yaml
# encryption-config.yaml should use aescbc or secretbox provider
```

## Checks to Generate

- Flag Kubernetes Secrets with `data` values that are not backed by external secrets manager.
- Grep Deployment/StatefulSet manifests for `value: "secret"` patterns (plaintext secrets in YAML).
- Check for etcd encryption at rest configuration in cluster setup.
- Flag RBAC Roles with `get`/`list` on `secrets` resource for broad subjects.
- Grep for secrets referenced via `secretKeyRef` — verify RBAC restricts access to these secrets.
- Flag `kubectl get secret -o yaml` in CI scripts — secrets exposed in CI logs.
- Check for `automountServiceAccountToken: true` on pods that don't need API server access.
