# Phase 0 — Detailed Implementation Plan

> **Goal:** Get documents in, find them again.
> **Structure:** Phase 0 splits into **0a** — API-only technical validation, exiting on a measurable corpus benchmark — and **0b** — hosted deployment with an MCP server as the tester interface, exiting on a written go/no-go verdict memo.
> **Success criteria:** 0a — a ~30-document real-home corpus (mixed digital PDFs, scans, phone photos) ingests ≥90% without error and is ≥85% findable via obvious keyword in top-3 results. 0b — ≥2 external testers each ingest ≥10 real documents and run ≥5 searches; a one-page verdict memo (continue / pivot / stop) is committed to the repo.
> **Decision record:** [`plans/phase0-interview.md`](plans/phase0-interview.md)

---

## Table of Contents

1. [Scope](#1-scope)
2. [Components](#2-components)
3. [Technical Decisions](#3-technical-decisions)
4. [Data Model](#4-data-model)
5. [API](#5-api)
6. [Frontend](#6-frontend)
7. [Local Development](#7-local-development)
8. [Task Sequence](#8-task-sequence)
9. [Done Definition](#9-done-definition)

---

## 1. Scope

### In — 0a

- Single user auth — email/password, server-side sessions
- One **space** per user — create, name, done. A space is a physical volume a person has ownership/charge over (home, office, warehouse); "home" is UI wording, not schema
- Document upload — PDF and image files (JPG, PNG); text extracted directly from digital PDFs, OCR for scanned PDFs and images
- Async ingestion via Temporal — upload returns immediately with `processing_status`; the client polls until `ready`/`failed`
- Keyword search — full-text plus fuzzy matching over the extracted text, returns matching documents with source references
- Corpus benchmark harness — `make benchmark` runs the ~30-document real corpus and prints ingest% / findable%; the 0a exit meter
- Local dev environment (minikube + Helm, including Temporal server + worker)

### In — 0b

- Document by URL — user pastes a URL pointing to a manual PDF, the system downloads and indexes it through the same flow
- MCP server — upload / list / search exposed as tools; the tester interface (no web UI in Phase 0)
- Hosted deployment with the privacy floor: TLS everywhere, private bucket + short-TTL pre-signed URLs, sessions ≤ 30 days, no document text in logs, delete-my-account hard-purges
- 2–3 external testers using real documents after plain-language consent

### Out

- Frontend / web UI — MCP is the Phase 0 interface; the web framework decision moves to Phase 1
- Make/model lookup agent — deferred to a later phase; the URL-fetch path covers the "get a manual in" need for now
- Multiple spaces
- Rooms, appliances as structured entities
- Roles, access control, sharing
- Schedules, notifications
- LLM calls of any kind — no extraction, no RAG, no vector search, no AI at query time
- Deduplication — probable duplicates evaluated by metadata later, if ever
- Mobile

---

## 2. Components

### 2.1 Auth Service

Email/password registration and login. Server-side sessions stored in Postgres. No OAuth, no magic links, no password reset flow — those come later.

### 2.2 Space

A named container. One per user. No address, no rooms, no metadata beyond a name. The user creates it once and it persists. Internally the entity is a **space** — a physical volume a person has ownership/charge over; how it's surfaced to end-users (home, place, …) is a UI decision deferred to 0b.

### 2.3 Document Ingestion

Two entry points, same output: a stored file and a full-text index entry. Ingestion is **asynchronous via a Temporal workflow**: the API stores the original, creates the document row with `processing_status=pending`, starts the workflow, and returns immediately; the client polls the document until `ready` or `failed`. Every ingest terminally reaches one of those two states — a document stuck `pending` longer than 30 minutes is a monitored invariant violation.

**Upload path (0a):**
- User uploads a file (PDF, JPG, PNG), capped at 50 MB
- File stored in object storage (S3-compatible); document row created as `pending`
- Temporal workflow extracts text:
  - Digital PDF → text extracted directly (pdftotext or equivalent)
  - Scanned PDF / image → OCR
- Extracted text written to the full-text index, linked to the stored file; status flips to `ready` (or `failed`, with the original retained)

**URL path (0b):**
- User provides a URL pointing to a PDF (e.g. a manufacturer's manual)
- System downloads the file with timeout, size, and content-type checks
- If valid: stored, extracted, and indexed through the same flow as an upload
- If not: clear error (unreachable, too large, not a PDF), offer manual upload as fallback

### 2.4 Full-Text Index

Postgres is sufficient for Phase 0: `tsvector`/`tsquery` for full-text ranking, `pg_trgm` for fuzzy/typo-tolerant matching. No Elasticsearch, no Meilisearch — avoid operational complexity until search quality becomes a bottleneck.

Each indexed document stores:
- The extracted text as a `tsvector`
- A reference to the source file in object storage
- The document name, source (uploaded vs fetched from URL), and kind (`common` vs `space_specific`)
- The space it belongs to

### 2.5 Search

Single search endpoint. Input is a keyword query. Output is a ranked list of matching documents with:
- Document name
- A short excerpt showing where the match occurred
- Link to view or download the original file

---

## 3. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Consistent with project direction |
| HTTP router | chi | Already chosen |
| Database | PostgreSQL | Already chosen; handles Phase 0 search |
| Full-text search | Postgres `tsvector` + `pg_trgm` | Zero additional infra; fuzzy matching via trigrams; revisit if search quality is a blocker |
| Object storage | S3-compatible (LocalStack in-cluster locally, R2 or S3 in prod) | Keeps the S3 API without extra services locally |
| Local orchestration | minikube + Helm | Matches the deployment shape from day one; chart holds Postgres and (per the Temporal decision) Temporal server + worker |
| Migrations | goose, run via `make migrate-up` (port-forward to in-cluster Postgres) | Already wired in the Makefile |
| PDF text extraction | TBD — pdftotext (poppler), unipdf, or equivalent Go library | Needs spike |
| OCR | Tesseract (local) | Keeps Phase 0 free of external API dependencies; the NYUM-006 spike probes it against 10 real phone photos before the pipeline is built — pre-approved fallback: cloud OCR API or photos-out-of-0a |
| Document ingestion timing | **Async via Temporal workflow** — client polls `processing_status` | Strategic bet ([interview dec. 12](plans/phase0-interview.md)): Temporal is the committed Phase 1+ platform; no interim plumbing; de-risking Temporal-on-minikube is itself validation. Reverses the earlier synchronous decision |
| Document by URL | Simple HTTP fetcher — download PDF from user-supplied URL | **Moved to 0b** (interview dec. 11): its value claim is onboarding friction, a user-behavior concern; 0a validates technical claims only |
| Delete semantics | Soft delete (tombstone) by default; hard delete only for space-specific items and account purge | "Knowledge compounds" — documents for common appliances are rarely-if-ever deleted (interview dec. 8) |
| Document classes | `document.kind` — `common` \| `space_specific` (default) | One column now preserves the future global-manual-library option and the split delete semantics (interview dec. 9) |
| Sessions | Server-side, stored in Postgres | Simple, auditable, no JWT complexity |
| Interface | 0a: curl + integration tests. 0b: **MCP server** (upload/list/search as tools) — no web UI in Phase 0 | Cheapest tester interface, rehearses the Tier 2 conversational experience; web framework decision moves to Phase 1 |

---

## 4. Data Model

Minimal. Only what Phase 0 needs.

```
user
  id, email, password_hash, created_at

session
  id, user_id, expires_at, created_at

space
  id, user_id, name, created_at

document
  id, space_id
  name              -- display name
  kind              -- "common" | "space_specific" (default) — preserves the future
                    --   global-manual-library option and split delete semantics
  source            -- "upload" | "url"
  source_url        -- populated for URL-fetched documents
  file_key          -- S3 object key
  file_type         -- "pdf" | "image"
  processing_status -- "pending" | "ready" | "failed"
  extracted_text    -- raw text after OCR/extraction
  search_vector     -- tsvector, generated from extracted_text
  deleted_at        -- soft delete; every query filters this
  indexed_at, created_at
```

No appliances table. No rooms table. No schedules. The document *is* the entity for now. DELETE tombstones (`deleted_at`); the only hard-delete path is account purge (0b privacy floor).

---

## 5. API

Minimal REST API. All routes require auth except `/auth/*`.

```
POST  /auth/register
POST  /auth/login
POST  /auth/logout

GET   /space             — get the user's space
POST  /space             — create space (if none exists)

POST  /space/documents          — upload a document (async: returns pending, client polls)
POST  /space/documents/from-url — add a document by URL { url, name? }   [0b]
GET   /space/documents          — list documents
GET   /space/documents/:id      — get document details (incl. processing_status)
GET   /space/documents/:id/file — download original file
DELETE /space/documents/:id     — soft-delete document (tombstone)

GET   /space/search?q=...       — keyword search, returns matching documents
```

---

## 6. Interface

**No web UI in Phase 0.** In 0a the API is exercised with curl and integration tests. In 0b an **MCP server** exposes upload / add-by-URL / list / search as tools — testers (chosen from Claude users) manage their space conversationally, which doubles as an early rehearsal of the Tier 2 assistant experience. The web framework decision (htmx + Go templates vs a SPA) is revisited when Phase 1 introduces roles, sharing, and the guest portal.

---

## 7. Local Development

Everything runs on minikube, deployed via the Helm chart at `deployments/helm/nyumspace`:

```
postgres    — StatefulSet (already in the chart); full-text search + pg_trgm, no pgvector needed yet
localstack  — Deployment + Service; S3-compatible object storage, with bucket initialization
temporal    — Temporal server (+ UI); backs the ingestion workflow
app         — Go API, built locally and deployed via the chart
worker      — Temporal worker running the ingestion workflow; second deployable
```

Dev loop:

```
make mk-build     # build the app image into minikube's docker daemon
make helm-up      # install/upgrade the release (starts minikube tunnel)
make migrate-up   # apply goose migrations via port-forward to in-cluster postgres
```

Temporal is in from 0a — a deliberate reversal of the earlier "no Temporal in Phase 0" decision ([interview dec. 12](plans/phase0-interview.md)): it's the committed Phase 1+ platform, so no interim queue plumbing gets built, and running it on minikube early is itself technical validation. Accepted cost: heavier chart, a second deployable (worker). This risk is explicitly carried unmitigated — the rejected fallback (river) is recorded in the interview for the record.

---

## 8. Task Sequence

The Phase 0 work is broken into discrete tickets in [`plans/epochs/phase0/`](plans/epochs/phase0/README.md) — one file per ticket, each self-contained enough to hand to a person or an agent.

Ordering in brief (0a): the project skeleton (001) and local environment (002, now including Temporal) can proceed in parallel; auth (003) and space (004) build on them sequentially; document storage (005) follows; the PDF extraction spike (006, including the 10-photo OCR probe) can run in parallel with any of 003–005; the ingestion workflow (007) needs 002 + 005 + 006; search (008) and the benchmark harness (010) close 0a. 0b tickets (document-by-URL 009, hosted deploy + privacy floor 011, MCP server 012, tester onboarding 013) live in [`plans/epochs/phase0b/`](plans/epochs/phase0b/README.md).

---

## 9. Done Definition

**0a is done when:**

1. A user can register, log in, and log out
2. A user can create a space
3. A user can upload a PDF or image; it becomes searchable asynchronously (poll `processing_status`), p95 upload→`ready`/`failed` ≤ 5 min, nothing stuck `pending` > 30 min
4. A user can search with a keyword and get back matching documents with excerpts; search p95 ≤ 300 ms at a 1,000-document synthetic corpus
5. `make benchmark` on the ~30-document real corpus prints **≥90% ingest** and **≥85% findable** (obvious keyword, top-3)
6. All four structural invariants have explicit green integration tests: no cross-space leakage, no orphaned index entries, no silent ingestion failure, no credential exposure
7. All of the above runs locally on minikube via the make targets

**0b (and Phase 0) is done when:**

8. The system is deployed on real infrastructure with the full privacy floor (TLS, private bucket + short-TTL pre-signed URLs, sessions ≤ 30 days, no document text in logs, delete-my-account hard-purges)
9. A user can add a manual by URL and get it indexed (or a clear failure message)
10. The MCP server exposes upload / add-by-URL / list / search, and ≥2 external testers have each ingested ≥10 real documents and run ≥5 searches through it
11. A one-page verdict memo (continue / pivot / stop + what Phase 1 must fix) is committed to the repo — **the memo, not the code, is Phase 0's final deliverable**
