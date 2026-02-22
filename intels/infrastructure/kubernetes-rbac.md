---
id: infra-kubernetes-rbac
title: Kubernetes RBAC Misconfigurations
severity: high
tags: [infrastructure, kubernetes, rbac, k8s, container, privilege-escalation]
taxonomy: security/infrastructure/kubernetes
references:
  - https://owasp.org/www-project-kubernetes-top-ten/
  - https://kubernetes.io/docs/reference/access-authn-authz/rbac/
---

# Kubernetes RBAC Misconfigurations

## Description

Kubernetes RBAC misconfigurations are a primary attack path for cluster compromise. Overly permissive ClusterRoles, binding `cluster-admin` to service accounts, and wildcard verbs/resources allow attackers who compromise a pod to escalate to cluster-wide control, steal secrets, or pivot to other workloads.

## Vulnerable Pattern

```yaml
# BAD — ClusterRoleBinding to cluster-admin for application service account
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: app-admin
subjects:
- kind: ServiceAccount
  name: app
  namespace: production
roleRef:
  kind: ClusterRole
  apiRef: cluster-admin  # full cluster control!

# BAD — ClusterRole with wildcard verbs and resources
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: too-permissive
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]  # equivalent to cluster-admin
```

```yaml
# BAD — Pod spec mounting service account token unnecessarily
spec:
  automountServiceAccountToken: true  # default — token accessible in every pod
```

## Secure Pattern

```yaml
# GOOD — minimal Role for application needs
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: app-role
  namespace: production
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list"]  # only what's needed
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["app-secret"]  # only specific secret
  verbs: ["get"]

# GOOD — disable automounting when not needed
spec:
  automountServiceAccountToken: false
  containers:
  - name: app
    image: myapp:1.0.0
    securityContext:
      runAsNonRoot: true
      runAsUser: 1000
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
```

## Checks to Generate

- Flag `cluster-admin` ClusterRoleBinding to service accounts or non-admin users.
- Grep for `verbs: ["*"]` or `resources: ["*"]` in Role/ClusterRole — flag as overly permissive.
- Flag `automountServiceAccountToken: true` (default) on pods that don't need API access.
- Check for missing `runAsNonRoot: true` in pod security context.
- Flag `allowPrivilegeEscalation: true` or absence of `allowPrivilegeEscalation: false`.
- Grep for `privileged: true` in container security context — extremely dangerous.
- Check for missing `capabilities: drop: ["ALL"]` in security context.
