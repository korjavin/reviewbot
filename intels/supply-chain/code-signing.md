---
id: supply-chain-code-signing
title: Missing Code Signing and Artifact Integrity Verification
severity: high
tags: [supply-chain, code-signing, artifacts, integrity, ci-cd]
taxonomy: security/supply-chain/code-signing
references:
  - https://owasp.org/www-project-top-ten/2021/A08_2021-Software_and_Data_Integrity_Failures/
  - https://slsa.dev/
---

# Missing Code Signing and Artifact Integrity Verification

## Description

Build artifacts (binaries, container images, release archives) without cryptographic signatures can be tampered with between build and deployment. Without verification, a compromised artifact store or delivery channel can push malicious code to production. SLSA (Supply-chain Levels for Software Artifacts) framework defines provenance requirements.

## Vulnerable Pattern

```bash
# BAD — downloading and executing scripts without verification
curl https://install.example.com/install.sh | bash
# Attacker intercepts or compromises CDN → malicious script executes

# BAD — container image pulled without digest verification
FROM myregistry.io/myapp:latest
# Container orchestrator pulls "latest" — could be different image than tested
```

```yaml
# BAD — CI/CD: artifact uploaded without signing
- name: Build release
  run: |
    go build -o myapp-v1.0.0-linux-amd64 .
    aws s3 cp myapp-v1.0.0-linux-amd64 s3://releases/
    # No signature generated — users can't verify authenticity
```

## Secure Pattern

```bash
# GOOD — verify script signature before executing
# Method 1: GPG-signed releases
curl https://releases.example.com/install.sh -o install.sh
curl https://releases.example.com/install.sh.sig -o install.sh.sig
gpg --verify install.sh.sig install.sh && bash install.sh

# Method 2: checksum verification
curl https://releases.example.com/myapp-v1.0.0-linux-amd64 -o myapp
curl https://releases.example.com/SHA256SUMS -o SHA256SUMS
sha256sum --check --ignore-missing SHA256SUMS
```

```yaml
# GOOD — sign container images with cosign (Sigstore)
- name: Sign container image
  run: |
    cosign sign --yes \
      --key env://COSIGN_PRIVATE_KEY \
      ${{ env.REGISTRY }}/${{ env.IMAGE }}@${{ steps.build.outputs.digest }}

# Verify at deploy time
- name: Verify image signature
  run: |
    cosign verify \
      --certificate-identity=${{ env.SIGNER_IDENTITY }} \
      --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
      ${{ env.REGISTRY }}/${{ env.IMAGE }}@${{ env.DIGEST }}
```

## Checks to Generate

- Grep CI/CD for `curl ... | bash` or `wget ... | sh` — pipe to shell without verification.
- Flag container image references without digest pinning (`@sha256:...`).
- Check for absence of cosign/notation image signing in container build pipeline.
- Grep for release upload steps without corresponding SHA256SUMS or GPG signature generation.
- Flag download-and-execute patterns in installation scripts without integrity check.
- Check for SLSA provenance attestation generation in release pipeline.
