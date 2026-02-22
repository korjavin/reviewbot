---
id: infra-docker-security
title: Docker Container Security Misconfigurations
severity: high
tags: [infrastructure, docker, container, privilege, root, security]
taxonomy: security/infrastructure/docker
references:
  - https://owasp.org/www-project-docker-top-10/
  - https://docs.docker.com/engine/security/
  - https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html
---

# Docker Container Security Misconfigurations

## Description

Insecure Docker configurations allow container escape, privilege escalation, and lateral movement. Running containers as root, mounting the Docker socket, using `--privileged` mode, and disabling seccomp/AppArmor profiles all reduce container isolation.

## Vulnerable Pattern

```dockerfile
# BAD — running as root (default)
FROM node:18
WORKDIR /app
COPY . .
RUN npm install
CMD ["node", "server.js"]  # no USER → root process

# BAD — installing unnecessary tools (expanded attack surface)
RUN apt-get install -y curl wget netcat nmap vim
```

```yaml
# BAD — docker-compose: mounting Docker socket (container escape)
services:
  app:
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock  # full Docker control from container!
    privileged: true  # full host kernel access
    cap_add:
      - SYS_ADMIN  # dangerous capability
```

```bash
# BAD — docker run with dangerous flags
docker run --privileged \
           --pid=host \       # see host processes
           --network=host \   # bypass network isolation
           -v /:/host \       # mount entire host filesystem
           myapp
```

## Secure Pattern

```dockerfile
# GOOD — minimal image, non-root user, read-only filesystem
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:18-alpine
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY --chown=appuser:appgroup . .
USER appuser
EXPOSE 3000
CMD ["node", "server.js"]
```

```yaml
# GOOD — docker-compose with security constraints
services:
  app:
    read_only: true
    tmpfs:
      - /tmp
    security_opt:
      - no-new-privileges:true
      - seccomp:seccomp-profile.json
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE  # only if needed for port <1024
```

## Checks to Generate

- Flag Dockerfiles without `USER` instruction — container runs as root.
- Flag `--privileged` in `docker run` commands or `privileged: true` in compose files.
- Grep for `/var/run/docker.sock` volume mount — grants full Docker control.
- Flag `cap_add: SYS_ADMIN`, `SYS_PTRACE`, `NET_ADMIN` — dangerous capabilities.
- Flag `--network=host`, `--pid=host`, `--ipc=host` options.
- Check for missing `no-new-privileges:true` security option.
- Grep for `DEBUG` or `DEVELOPMENT` environment variables set in production Dockerfiles.
