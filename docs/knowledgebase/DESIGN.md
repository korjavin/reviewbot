# Knowledge Base — Design & Storage Evaluation

> **Parent**: [REQUIREMENTS.md](REQUIREMENTS.md)  
> **Status**: Draft · February 2026

---

## 1. Data Model

```
Taxonomy Node
┌────────────────────────────┐
│ path        string (PK)    │  e.g. "security/ci-cd/github-actions"
│ name        string         │  e.g. "GitHub Actions"
│ description string         │
│ aliases     []string       │
│ parent_path string (FK)    │
│ created_at  timestamp      │
│ updated_at  timestamp      │
└────────────────────────────┘
        │
        │ (many-to-many)
        │
Intel
┌────────────────────────────────────────────────────┐
│ id              UUID (PK)                          │
│ title           string                             │
│ body            Markdown text                      │
│ severity        enum (critical/high/medium/low/info)│
│ tags            []string                           │
│ taxonomy_nodes  []string  (FK → node.path)         │
│ references      []URL                              │
│ embedding       []float32  (computed on ingest)    │
│ created_at      timestamp                          │
│ updated_at      timestamp                          │
└────────────────────────────────────────────────────┘
```

The `tags` and `taxonomy_nodes` fields on Intel create two orthogonal retrieval dimensions (see [REQUIREMENTS §2](REQUIREMENTS.md)).

---

## 2. Storage Architecture Decision

### 2.1 Option A — libSQL (Turso) ✅ Recommended

libSQL is an open-contribution fork of SQLite by Turso that adds **native vector search** directly into the database engine. Both structured data (taxonomy, Intel metadata) and vector embeddings live in a **single file** — no sidecar service required.

| Attribute | Value |
|-----------|-------|
| Language | Rust (server / embedded) + official Go client `tursodatabase/libsql-client-go` |
| Licence | MIT |
| Vector type | Native `F32_BLOB(dim)` column — no extension needed (GA since Oct 2024) |
| ANN algorithm | DiskANN — cosine similarity via `vector_top_k()` |
| Metadata filtering | Standard SQL `WHERE` — join vector results with structured tables |
| Taxonomy | Recursive CTEs on `nodes` table; adjacency list — identical to SQLite |
| Deployment | **Single binary / single `.db` file** — zero extra infrastructure |
| Maturity | Medium-high — production use at Turso, active development |
| Go support | `tursodatabase/libsql-client-go` + `libsql-vector-go` helper for embedding types |

**Why libSQL:** Eliminates the two-store complexity entirely. Taxonomy tree, Intel metadata, tags, and embedding vectors all live in one `.db` file. DiskANN provides production-quality approximate nearest-neighbour search. The Go driver is a drop-in for `database/sql`. Backup is a single file copy.

#### Schema sketch (libSQL)

```sql
-- Taxonomy
CREATE TABLE nodes (
  path        TEXT PRIMARY KEY,   -- e.g. 'security/ci-cd/github-actions'
  name        TEXT NOT NULL,
  description TEXT,
  parent_path TEXT REFERENCES nodes(path)
);

-- Intels
CREATE TABLE intels (
  id             TEXT PRIMARY KEY,   -- UUID
  title          TEXT NOT NULL,
  body           TEXT NOT NULL,      -- Markdown
  severity       TEXT NOT NULL,
  references     TEXT,               -- JSON array of URLs
  created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
  embedding      F32_BLOB(768)        -- native libSQL vector column
);

-- Many-to-many: intels ↔ taxonomy nodes
CREATE TABLE intel_nodes (
  intel_id TEXT REFERENCES intels(id),
  node_path TEXT REFERENCES nodes(path),
  PRIMARY KEY (intel_id, node_path)
);

-- Tags
CREATE TABLE intel_tags (
  intel_id TEXT REFERENCES intels(id),
  tag      TEXT NOT NULL,
  PRIMARY KEY (intel_id, tag)
);

-- ANN index on the embedding column
CREATE INDEX intels_embedding_idx ON intels (libsql_vector_idx(embedding));
```

#### Semantic search query (libSQL)

```sql
-- Top-10 closest Intels to a query vector, filtered by tag
SELECT i.id, i.title, i.severity,
       vector_distance_cos(i.embedding, vector(?)) AS score
FROM vector_top_k('intels_embedding_idx', vector(?), 20) AS v
JOIN intels i ON i.id = v.id
JOIN intel_tags t ON t.intel_id = i.id AND t.tag = ?
ORDER BY score
LIMIT 10;
```

#### docker-compose (Option A — libSQL only)

```yaml
services:
  kb:
    build: ./services/kb
    environment:
      DB_PATH: /data/kb.db
    volumes:
      - kb-data:/data

volumes:
  kb-data:
```

---

### 2.2 Option B — SQLite + Qdrant (Scale-Out)

For workloads exceeding ~50 k Intels or requiring concurrent high-throughput vector queries, a two-store architecture becomes worthwhile:

| Layer | Engine | Role |
|-------|--------|------|
| Structured | SQLite (`modernc.org/sqlite` — pure Go) | Taxonomy + Intel metadata |
| Vector | **Qdrant** (`qdrant/qdrant` Docker image) | HNSW embeddings + payload filter |

Qdrant advantages over libSQL's DiskANN at scale: payload-filtered ANN in one round-trip (no SQL join), scalar/product quantisation, horizontal scaling.

```yaml
# docker-compose (Option B)
services:
  kb:
    build: ./services/kb
    environment:
      QDRANT_URL: http://qdrant:6334
      SQLITE_PATH: /data/kb.db
    volumes: [kb-data:/data]
    depends_on: [qdrant]
  qdrant:
    image: qdrant/qdrant:latest
    volumes: [qdrant-storage:/qdrant/storage]
volumes:
  kb-data:
  qdrant-storage:
```

---

### 2.3 Decision

> **Start with libSQL (Option A).** A single `.db` file is the lowest-friction path to a working KB. The storage layer is hidden behind a `Store` interface, so migrating to Option B later requires changing only the adapter, not the service API.

---

### 2.4 Candidates Considered (Not Chosen For v1)

| Option | Reason skipped |
|--------|----------------|
| PostgreSQL + pgvector | Extra infra; not needed at this scale |
| Weaviate | Heavier footprint; no native tree traversal |
| chromem-go | In-memory only; no concurrent writes |
| KektorDB | Too early-stage |
| SQLite + sqlite-vec | Extension maturity lower than libSQL native vectors |

---

## 3. Embedding Pipeline

```
New / Updated Intel
        │
        ▼
EmbeddingProvider.Embed(text)
        │
   ┌────┴──────────────────────┐
   │  Provider interface       │
   │  ─────────────────        │
   │  OllamaProvider           │  (default; local model, e.g. nomic-embed-text)
   │  OpenAIProvider           │  (configurable via env)
   │  MockProvider             │  (tests)
   └───────────────────────────┘
        │
        ▼
Store embedding in DB alongside Intel
```

Embedding dimension: 768 (nomic-embed-text) or 1536 (OpenAI text-embedding-3-small). Dimension is stored in DB metadata.

---

## 4. API Design

### gRPC (primary — internal service-to-service)

```protobuf
service KnowledgeBase {
  // Taxonomy
  rpc GetNode(GetNodeRequest)          returns (Node);
  rpc ListChildren(ListChildrenRequest) returns (NodeList);
  rpc ListSubtree(ListSubtreeRequest)  returns (NodeList);
  rpc UpsertNode(UpsertNodeRequest)    returns (Node);
  rpc DeleteNode(DeleteNodeRequest)    returns (google.protobuf.Empty);

  // Intels
  rpc GetIntel(GetIntelRequest)        returns (Intel);
  rpc CreateIntel(CreateIntelRequest)  returns (Intel);
  rpc UpdateIntel(UpdateIntelRequest)  returns (Intel);
  rpc DeleteIntel(DeleteIntelRequest)  returns (google.protobuf.Empty);

  // Retrieval
  rpc SearchByTag(SearchByTagRequest)           returns (IntelList);
  rpc SearchByTaxonomy(SearchByTaxonomyRequest) returns (IntelList);
  rpc SearchSemantic(SearchSemanticRequest)     returns (ScoredIntelList);
  rpc SearchHybrid(SearchHybridRequest)         returns (ScoredIntelList);

  // Bulk
  rpc Import(ImportRequest)            returns (ImportResponse);
  rpc Export(ExportRequest)            returns (ExportResponse);
}
```

### REST/JSON gateway (for n8n nodes and browser tooling)

Thin HTTP wrapper generated via `grpc-gateway`. All gRPC methods are exposed under `/api/v1/...`.

Notable endpoints:

```
GET  /api/v1/taxonomy/{path}/children
GET  /api/v1/taxonomy/{path}/subtree
GET  /api/v1/taxonomy/{path}/intels

POST /api/v1/intels/search/tags
POST /api/v1/intels/search/semantic
POST /api/v1/intels/search/hybrid

POST /api/v1/admin/import
GET  /api/v1/admin/export
```

---

## 5. Package Structure

```
services/kb/
├── cmd/
│   └── main.go              ← gRPC server + REST gateway
├── internal/
│   ├── store/
│   │   ├── store.go         ← Store interface (structured data)
│   │   └── sqlite.go        ← SQLite implementation (taxonomy + Intel metadata)
│   ├── vector/
│   │   ├── client.go        ← Qdrant client wrapper
│   │   └── mock.go          ← In-memory mock for tests
│   ├── embedding/
│   │   ├── provider.go      ← EmbeddingProvider interface
│   │   ├── ollama.go
│   │   ├── openai.go
│   │   └── mock.go
│   └── taxonomy/
│       └── tree.go          ← In-memory tree helpers (path validation, subtree walks)
├── proto/
│   └── kb.proto
├── Dockerfile
└── README.md
```

The `pkg/knowledgebase/` directory at the repo root holds only the **client library** (generated gRPC stubs + convenience wrappers) used by the pipeline engine and future integrations.

---

## 6. Seed Knowledge Base

Initial Intels will be authored in YAML and committed to the repository under `services/kb/seed/`. The service automatically imports them on first start.

Categories to cover in the seed corpus:

- Web server vulnerabilities (OWASP Top 10 mapped to code patterns)
- gRPC / protobuf service vulnerabilities
- GitHub Actions security (secret leakage, unpinned actions, script injection)
- Dockerfile hardening
- Go-specific patterns (SSRF, path traversal, unsafe deserialization)
- Common dependency vulnerabilities checklist
