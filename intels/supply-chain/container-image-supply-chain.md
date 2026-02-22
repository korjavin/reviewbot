---
id: supply-chain-container-image
title: Container Image Supply Chain Security
severity: high
tags: [supply-chain, docker, container, image-scanning, sbom]
taxonomy: security/supply-chain/container
references:
  - https://owasp.org/www-project-devsecops-guideline/
  - https://docs.docker.com/build/attestations/sbom/
---

# Container Image Supply Chain Security

## Description

Container images inherit vulnerabilities from base images, installed OS packages, and application dependencies. Without image scanning, teams ship known CVEs to production. Supply chain attacks target popular base images and package registries. Images from `latest` tags may change between builds.

## Vulnerable Pattern

```dockerfile
# BAD — unpinned base image, runs as root, no vulnerability scanning
FROM python:latest          # different image each build
RUN pip install flask       # no version pin

# BAD — secrets baked into image layer (visible in docker history)
RUN export DB_PASSWORD=secret123 && ./configure_db.sh
ENV API_KEY=sk_live_abc123   # stored in image layer permanently
```

```yaml
# BAD — CI builds without image scanning
- name: Build and push
  run: |
    docker build -t myapp:latest .
    docker push myapp:latest  # no vulnerability scan before push
```

## Secure Pattern

```dockerfile
# GOOD — pinned digest, non-root user, minimal base
FROM python:3.11.9-slim@sha256:abc123...  # exact digest

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .
RUN useradd --system --no-create-home appuser && chown -R appuser /app
USER appuser

EXPOSE 8080
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8080"]
```

```yaml
# GOOD — scan image before push, fail on HIGH/CRITICAL
- name: Build image
  run: docker build -t myapp:${{ github.sha }} .

- name: Scan with Trivy
  uses: aquasecurity/trivy-action@main
  with:
    image-ref: myapp:${{ github.sha }}
    severity: HIGH,CRITICAL
    exit-code: 1  # fail build if vulnerabilities found

- name: Generate SBOM
  run: |
    docker sbom myapp:${{ github.sha }} > sbom.json

- name: Push (only if scan passed)
  run: docker push myapp:${{ github.sha }}
```

## Checks to Generate

- Flag `FROM <image>:latest` — must pin to specific version or digest.
- Flag Dockerfile without `USER` instruction — runs as root.
- Flag `ENV` or `RUN export` with credential patterns in Dockerfile.
- Check CI/CD for absence of image vulnerability scanning step (Trivy, Snyk, Grype, Clair).
- Flag `docker push` without preceding security scan in CI pipeline.
- Check for SBOM (Software Bill of Materials) generation in build pipeline.
- Flag base images from unofficial/unverified sources (not `docker.io/library/`, `gcr.io/`, `public.ecr.aws/`).
