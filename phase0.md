# Phase 0 — Detailed Implementation Plan

> **Goal:** Get documents in, find them again.
> **Success criteria:** A single user can create a home, add documents (by upload or by URL), and find them through keyword search.

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

### In

- Single user auth — email/password, server-side sessions
- One home per user — create, name, done
- Document upload — PDF and image files (JPG, PNG); text extracted directly from digital PDFs, OCR for scanned PDFs and images
- Document by URL — user pastes a URL pointing to a manual PDF, the system downloads and indexes it
- Keyword search — full-text plus fuzzy matching over the extracted text, returns matching documents with source references
- Local dev environment (minikube + Helm)

### Out

- Frontend / UI — Phase 0 is API-only, validated with curl and integration tests
- Make/model lookup agent — deferred to a later phase; the URL-fetch path covers the "get a manual in" need for now
- Multiple homes
- Rooms, appliances as structured entities
- Roles, access control, sharing
- Schedules, notifications
- LLM calls of any kind — no extraction, no RAG, no vector search, no AI at query time
- Production deployment
- Mobile

---

## 2. Components

### 2.1 Auth Service

Email/password registration and login. Server-side sessions stored in Postgres. No OAuth, no magic links, no password reset flow — those come later.

### 2.2 Home

A named container. One per user. No address, no rooms, no metadata beyond a name. The user creates it once and it persists.

### 2.3 Document Ingestion

Two entry points, same output: a stored file and a full-text index entry.

**Upload path:**
- User uploads a file (PDF, JPG, PNG)
- File stored in object storage (S3-compatible)
- Text extracted:
  - Digital PDF → text extracted directly (pdftotext or equivalent)
  - Scanned PDF / image → OCR
- Extracted text written to the full-text index, linked to the stored file

**URL path:**
- User provides a URL pointing to a PDF (e.g. a manufacturer's manual)
- System downloads the file with timeout, size, and content-type checks
- If valid: stored, extracted, and indexed through the same flow as an upload
- If not: clear error (unreachable, too large, not a PDF), offer manual upload as fallback

### 2.4 Full-Text Index

Postgres is sufficient for Phase 0: `tsvector`/`tsquery` for full-text ranking, `pg_trgm` for fuzzy/typo-tolerant matching. No Elasticsearch, no Meilisearch — avoid operational complexity until search quality becomes a bottleneck.

Each indexed document stores:
- The extracted text as a `tsvector`
- A reference to the source file in object storage
- The document name and source (uploaded vs fetched from URL)
- The home it belongs to

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
| Local orchestration | minikube + Helm | Matches the deployment shape from day one; chart already holds Postgres |
| Migrations | goose, run via `make migrate-up` (port-forward to in-cluster Postgres) | Already wired in the Makefile |
| PDF text extraction | TBD — pdftotext (poppler), unipdf, or equivalent Go library | Needs spike |
| OCR | Tesseract (local) | Keeps Phase 0 free of external API dependencies; revisit accuracy in Phase 1 |
| Document ingestion timing | Synchronous | Simple; acceptable latency for Phase 0. API returns explicit processing status |
| Document by URL | Simple HTTP fetcher — download PDF from user-supplied URL | Replaces the make/model agent for Phase 0; no source discovery, no LLM |
| Sessions | Server-side, stored in Postgres | Simple, auditable, no JWT complexity |
| Frontend | Deferred — Phase 0 is API-only | Revisit at Phase 1 |

---

## 4. Data Model

Minimal. Only what Phase 0 needs.

```
user
  id, email, password_hash, created_at

session
  id, user_id, expires_at, created_at

home
  id, user_id, name, created_at

document
  id, home_id
  name              -- display name
  source            -- "upload" | "url"
  source_url        -- populated for URL-fetched documents
  file_key          -- S3 object key
  file_type         -- "pdf" | "image"
  processing_status -- "pending" | "ready" | "failed"
  extracted_text    -- raw text after OCR/extraction
  search_vector     -- tsvector, generated from extracted_text
  indexed_at, created_at
```

No appliances table. No rooms table. No schedules. The document *is* the entity for now.

---

## 5. API

Minimal REST API. All routes require auth except `/auth/*`.

```
POST  /auth/register
POST  /auth/login
POST  /auth/logout

GET   /home              — get the user's home
POST  /home              — create home (if none exists)

POST  /home/documents          — upload a document
POST  /home/documents/from-url — add a document by URL { url, name? }
GET   /home/documents          — list documents
GET   /home/documents/:id      — get document details
GET   /home/documents/:id/file — download original file
DELETE /home/documents/:id     — delete document

GET   /home/search?q=...       — keyword search, returns matching documents
```

---

## 6. Frontend

**Deferred.** Phase 0 ships no UI — the API is exercised with curl and integration tests. The framework decision (htmx + Go templates vs a SPA) is revisited when Phase 1 introduces roles, sharing, and the guest portal.

---

## 7. Local Development

Everything runs on minikube, deployed via the Helm chart at `deployments/helm/nyumspace`:

```
postgres    — StatefulSet (already in the chart); full-text search + pg_trgm, no pgvector needed yet
localstack  — Deployment + Service; S3-compatible object storage, with bucket initialization
app         — Go API, built locally and deployed via the chart
```

Dev loop:

```
make mk-build     # build the app image into minikube's docker daemon
make helm-up      # install/upgrade the release (starts minikube tunnel)
make migrate-up   # apply goose migrations via port-forward to in-cluster postgres
```

No Temporal in Phase 0. Document ingestion runs synchronously in the request. Temporal is introduced in Phase 1 when async processing becomes a reliability concern.

---

## 8. Task Sequence

The Phase 0 work is broken into discrete tickets in [`plans/epochs/phase0/`](plans/epochs/phase0/README.md) — one file per ticket, each self-contained enough to hand to a person or an agent.

Ordering in brief: the project skeleton (001) and local environment (002) can proceed in parallel; auth (003) and home (004) build on them sequentially; document storage (005) follows; the PDF extraction spike (006) can run in parallel with any of 003–005; the extraction pipeline (007) needs 005 + 006; search (008) and document-by-URL (009) build on the extraction pipeline.

---

## 9. Done Definition

Phase 0 is done when:

1. A user can register, log in, and log out
2. A user can create a home
3. A user can upload a PDF or image and it becomes searchable
4. A user can add a manual by URL and get it indexed (or a clear failure message)
5. A user can search with a keyword and get back matching documents with excerpts
6. All of the above runs locally on minikube via the make targets
7. The code is tested well enough that a refactor won't silently break the above
