# Knowledge Base — Requirements

> **Parent**: [ARCHITECTURE.md](../ARCHITECTURE.md)  
> **Status**: Draft · February 2026

---

## 1. Purpose

The Knowledge Base (KB) is the **central repository of security intelligence** used by ReviewBot's AI agents. It stores structured documents — called **Intels** — organised in a hierarchical taxonomy of vulnerability classes. Agents query it to discover which attack vectors are relevant to a given codebase and to retrieve detailed guidance for each.

---

## 2. Core Concepts

### 2.1 Taxonomy

A **taxonomy** is an arbitrary-depth tree of **topics**. Each node in the tree represents a category of security concern.

Example tree (illustrative):

```
Security
├── Public Services
│   ├── Web Server
│   │   ├── Authentication
│   │   │   ├── Missing Auth
│   │   │   └── Broken Session Management
│   │   ├── Injection
│   │   │   ├── SQL Injection
│   │   │   └── Command Injection
│   │   └── CSRF
│   └── TCP Service
│       ├── Unauthenticated Access
│       └── Plaintext Credentials
├── CI/CD
│   ├── GitHub Actions
│   │   ├── Secret Leakage
│   │   ├── Unpinned Actions
│   │   └── Script Injection
│   └── Dockerfile
│       ├── Running as Root
│       └── Exposed Secrets in Layers
├── Dependencies
│   ├── Outdated Packages
│   └── Known CVEs
└── Secrets Management
    ├── Hardcoded Credentials
    └── Insecure Storage
```

Requirements:
- Nodes have a **name**, **description**, and optional **aliases**.
- A node can have zero or more children (unlimited depth).
- A node can have zero or more Intels attached.
- The full **path** from root to a node (e.g. `security/public-services/web-server/injection/sql-injection`) serves as a stable identifier.

### 2.2 Intel

An **Intel** is a security knowledge document attached to one or more taxonomy nodes. It describes a specific class of vulnerability, how to detect it, and what evidence to look for in code.

Fields:

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Globally unique identifier |
| `title` | string | Short human-readable name |
| `body` | Markdown | Full description: what is it, why dangerous, how to detect, examples |
| `severity` | enum | `critical / high / medium / low / info` |
| `tags` | []string | Free-form labels for cross-cutting retrieval (e.g. `github-actions`, `web-server`, `golang`) |
| `taxonomy_nodes` | []path | One or more taxonomy node paths this Intel belongs to |
| `references` | []URL | CVEs, OWASP links, blog posts |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |
| `embedding` | []float32 | Vector representation for semantic search (computed on ingest) |

### 2.3 Tags

Tags provide a **cross-cutting dimension** over the taxonomy tree. While the taxonomy is hierarchical and reflects *what kind* of vulnerability it is, tags reflect *where* it applies (technology, framework, language, deploy target).

Examples: `golang`, `python`, `web-server`, `grpc`, `github-actions`, `docker`, `kubernetes`, `aws`, `jwt`, `oauth2`.

One Intel can carry multiple tags and belong to multiple taxonomy nodes simultaneously.

---

## 3. Functional Requirements

### FR-1: Taxonomy Management

| ID | Requirement |
|----|-------------|
| FR-1.1 | Create, read, update, and delete taxonomy nodes |
| FR-1.2 | Retrieve all **direct children** of a node by path |
| FR-1.3 | Retrieve the **full subtree** (all descendants) of a node by path |
| FR-1.4 | Return the **list of Intels** attached to a given node (non-recursive) |
| FR-1.5 | Return **all Intels** within a subtree (recursive, de-duplicated) |

### FR-2: Intel Management

| ID | Requirement |
|----|-------------|
| FR-2.1 | Create, read, update, and delete Intels |
| FR-2.2 | Attach / detach an Intel to/from taxonomy nodes |
| FR-2.3 | Add / remove tags on an Intel |
| FR-2.4 | On create/update, automatically compute and store the vector embedding of the Intel body |

### FR-3: Retrieval

| ID | Requirement |
|----|-------------|
| FR-3.1 | Retrieve Intels by **tag** (exact match, AND / OR logic) |
| FR-3.2 | Retrieve Intels by **taxonomy path** (exact node or full subtree) |
| FR-3.3 | **Semantic search**: given a natural-language query, return top-K most similar Intels (cosine similarity over embeddings) |
| FR-3.4 | Combine filters: semantic search restricted to a taxonomy subtree or tag set |
| FR-3.5 | All list endpoints support **pagination** |

### FR-4: Bulk Operations

| ID | Requirement |
|----|-------------|
| FR-4.1 | Import Intels from a YAML/JSON file (seed the knowledge base) |
| FR-4.2 | Export all Intels and taxonomy to a portable format |

---

## 4. Non-Functional Requirements

| ID | Requirement |
|----|-------------|
| NFR-1 | **Language**: Go (to stay consistent with the rest of the project) |
| NFR-2 | **Deployment**: Single Docker container (process + embedded or sidecar DB) |
| NFR-3 | **API**: gRPC (primary) + HTTP/JSON REST gateway for browser / n8n access |
| NFR-4 | **Latency**: p99 < 200 ms for retrieval queries on a corpus of ≤ 50 000 Intels |
| NFR-5 | **Persistence**: Data survives container restarts (volume-backed storage) |
| NFR-6 | **Embedding model**: Default to a local model (e.g. `nomic-embed-text` via Ollama) with a configurable provider interface for OpenAI-compatible APIs |
| NFR-7 | **Auth**: API key authentication for all mutating operations; read operations may be unauthenticated within the internal network |
| NFR-8 | **Observability**: Structured JSON logs, `/metrics` (Prometheus), `/healthz` |
| NFR-9 | **Licence**: All dependencies must be permissively licensed (MIT / Apache 2) |

---

## 5. Out of Scope (v1)

- User-facing web UI for browsing the taxonomy (CLI tool only)
- Multi-tenancy / per-team knowledge bases
- Automatic Intel generation from CVE feeds (future)
- Access control beyond a single API key

---

## 6. Open Questions

1. Should the embedding model run **inside the KB container** (via CGo/Ollama sidecar) or be a **separate microservice**? Sidecar keeps deployment simple; separate service is more flexible.
2. Do Intels need **versioning** (audit history of changes to a document)?
3. Should the taxonomy be **editable at runtime** or managed exclusively via files in the repository (GitOps style)?

---

## 7. Next Steps

- [ ] Evaluate storage backends → see [DESIGN.md](DESIGN.md)
- [ ] Define gRPC proto schema
- [ ] Build seed knowledge base (initial set of Intels for common Go/web vulnerabilities)
- [ ] Integrate KB client into the pipeline engine
