# Phase 0 — Detailed Implementation Plan

> **Goal:** Get documents in, find them again.
> **Success criteria:** A single user can create a home, add documents (by upload or by make/model lookup), and find them through keyword search.

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
- Document upload — PDF files accepted; text extracted directly from digital PDFs, OCR fallback for scanned PDFs
- Make/model lookup — user enters make and model, agent searches for and downloads the manual, indexes it
- Keyword search — queries the full-text index, returns matching documents with source references
- Local dev environment

### Out

- Multiple homes
- Rooms, appliances as structured entities
- Roles, access control, sharing
- Schedules, notifications
- LLM calls of any kind — no extraction, no RAG, no AI at query time
- Production deployment
- Image file uploads (JPG, PNG) — deferred; OCR infrastructure added when needed
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
- File stored in object storage (S3/R2)
- Text extracted:
  - Digital PDF → text extracted directly (pdftotext or equivalent)
  - Scanned PDF / image → OCR
- Extracted text written to the full-text index, linked to the stored file

**Make/model lookup path:**
- User enters make and model (and optionally a model number or product ID)
- Agent searches public manual databases and manufacturer sites for a matching PDF
- If found: download, store in object storage, extract text, index
- If not found: inform the user, offer manual upload as fallback

### 2.4 Full-Text Index

Postgres `tsvector`/`tsquery` is sufficient for Phase 0. No Elasticsearch, no Meilisearch — avoid operational complexity until search quality becomes a bottleneck.

Each indexed document stores:
- The extracted text as a `tsvector`
- A reference to the source file in object storage
- The document name and type (uploaded vs fetched)
- The home it belongs to

### 2.5 Search

Single search bar. Input is a keyword query. Output is a ranked list of matching documents with:
- Document name
- A short excerpt showing where the match occurred
- Link to view or download the original file

---

## 3. Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Consistent with project direction |
| HTTP router | chi | Already chosen |
| Database | PostgreSQL | Already chosen; `tsvector` handles Phase 0 search |
| Full-text search | Postgres `tsvector` | Zero additional infra; revisit if search quality is a blocker |
| Object storage | S3-compatible (LocalStack locally, R2 or S3 in prod) | Already chosen |
| PDF text extraction | TBD — pdftotext (poppler), unipdf, or equivalent Go library | Needs spike |
| OCR | Tesseract (local) | Keeps Phase 0 free of external API dependencies; revisit accuracy in Phase 1 |
| Document ingestion timing | Synchronous | Simple; acceptable latency for Phase 0. UI must show loading state and explicit success/failure feedback |
| Manual-fetching agent | Internal Go service | Searches ManualLib, ManualsOnline, and direct manufacturer URLs; no LLM involved |
| Sessions | Server-side, stored in Postgres | Simple, auditable, no JWT complexity |
| Frontend | TBD | See Section 6 |

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
  source            -- "upload" | "fetched"
  make, model       -- populated for fetched manuals
  file_key          -- S3 object key
  file_type         -- "pdf" | "image"
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
POST  /home/documents/fetch    — fetch by make/model
GET   /home/documents          — list documents
GET   /home/documents/:id      — get document details
GET   /home/documents/:id/file — download original file
DELETE /home/documents/:id     — delete document

GET   /home/search?q=...       — keyword search, returns matching documents
```

---

## 6. Frontend

### Options

| Option | Pros | Cons |
|--------|------|------|
| **Simple HTML + htmx** | No build tooling, Go templates, fast to build | Limited interactivity, harder to scale UI |
| **React (Vite)** | Component model, good ecosystem | More setup, separate build pipeline |
| **SvelteKit** | Simple, fast, good DX | Smaller ecosystem |

### Recommendation

For Phase 0, **htmx + Go templates** is the fastest path to a working UI without introducing a separate frontend build pipeline. The UI is simple — a few forms, a search bar, a list of results. No complex state management needed.

Revisit when Phase 1 introduces roles, sharing, and the guest portal, which will demand a more capable frontend.

### Pages

- `/` — home page: document list + search bar
- `/login`, `/register` — auth
- `/upload` — upload form
- `/fetch` — make/model entry form
- `/search` — search results
- `/documents/:id` — document detail + download

---

## 7. Local Development

Docker Compose services:

```
postgres   — database (with full-text search, no pgvector needed yet)
minio      — S3-compatible object storage
app        — Go API + frontend (hot reload)
```

No Temporal in Phase 0. Document ingestion runs synchronously in the request or as a simple goroutine. Temporal is introduced in Phase 1 when async processing becomes a reliability concern.

---

## 8. Task Sequence

Tasks are ordered by dependency. Each group can be worked in parallel within the group.

### Group 1 — Skeleton
- [ ] Clean repository structure — remove scratch code, establish `cmd/`, `internal/`, `migrations/` layout
- [ ] Docker Compose — Postgres, MinIO, app service with hot reload
- [ ] MinIO bucket initialization script in Compose
- [ ] Goose migrations setup — connection config, `make migrate` command
- [ ] chi HTTP server with graceful shutdown
- [ ] `GET /health` endpoint
- [ ] Environment config — load from `.env` for local, document required vars

### Group 2 — Auth
- [ ] Migration: `user` table — `id`, `email` (unique), `password_hash`, `created_at`
- [ ] Migration: `session` table — `id`, `user_id` (FK), `expires_at`, `created_at`
- [ ] Password hashing with bcrypt
- [ ] `POST /auth/register` — validate email format, check uniqueness, hash password, create user, start session, set cookie
- [ ] `POST /auth/login` — look up user by email, verify password, create session, set cookie
- [ ] `POST /auth/logout` — delete session from DB, clear cookie
- [ ] Session middleware — read session cookie, validate against DB, attach user to request context, reject expired sessions
- [ ] Basic rate limiting on `POST /auth/login` — max 10 attempts per IP per minute
- [ ] Integration tests: full register → login → authenticated request → logout cycle

### Group 3 — Home
- [ ] Migration: `home` table — `id`, `user_id` (FK, unique), `name`, `created_at`
- [ ] `POST /home` — create home; return 409 if user already has one
- [ ] `GET /home` — return home; return 404 if none exists
- [ ] After-login redirect: if user has no home, redirect to home creation page
- [ ] Integration tests: create home, get home, enforce one-per-user

### Group 4 — Document Storage
- [ ] Migration: `document` table — `id`, `home_id` (FK), `name`, `source` (`upload`/`fetched`), `make`, `model`, `file_key`, `file_type`, `processing_status` (`pending`/`ready`/`failed`), `extracted_text`, `search_vector`, `indexed_at`, `created_at`
- [ ] S3/MinIO client — upload, download, delete, pre-signed URL generation
- [ ] File validation — PDF only, max 50MB, reject everything else with a clear error
- [ ] `POST /home/documents` — accept multipart upload, validate, store in S3, create document record with `status=pending`, kick off extraction synchronously, update status on completion
- [ ] `GET /home/documents` — list documents for home, ordered by `created_at` desc
- [ ] `GET /home/documents/:id` — document detail including processing status
- [ ] `GET /home/documents/:id/file` — redirect to pre-signed S3 URL (short TTL)
- [ ] `DELETE /home/documents/:id` — delete from S3 and DB
- [ ] Integration tests: upload → list → get → delete

### Group 5 — Text Extraction
- [ ] Spike: evaluate PDF text extraction — test `pdfcpu`, `unipdf`, and `pdftotext` (via exec) against a sample set of real manuals; pick based on accuracy and license
- [ ] Digital PDF extraction — extract embedded text; treat as scanned if result is empty or under 100 characters
- [ ] Scanned PDF OCR — run Tesseract on rendered PDF pages when digital extraction yields insufficient text
- [ ] Tesseract setup in Docker Compose app container
- [ ] Extraction wired into upload flow — runs synchronously after S3 store; writes to `document.extracted_text`, sets `status=ready`
- [ ] Extraction failure handling — on error, set `status=failed`, document still listed and searchable by name; user sees clear status indicator
- [ ] UI loading state during upload + extraction — spinner or progress indicator; explicit success/failure feedback on completion

### Group 6 — Full-Text Index
- [ ] `search_vector tsvector` column on `document`
- [ ] Postgres trigger to regenerate `search_vector` from `extracted_text` on insert and update
- [ ] GIN index on `search_vector`
- [ ] `GET /home/search?q=...` — parse query into `tsquery`, run against index filtered by `home_id`, rank by `ts_rank_cd`
- [ ] Excerpt generation with `ts_headline` — highlight matching terms in result snippet
- [ ] Response shape: `{ id, name, source, make, model, excerpt, status }`
- [ ] Empty results state — return empty array with a message, not a 404
- [ ] Integration tests: upload document → search for term in its content → verify result appears

### Group 7 — Manual Fetching
- [ ] Spike (timeboxed 2 days): identify reliable manual sources — test ManualLib, ManualsOnline, and direct manufacturer URL patterns for coverage and reliability; document findings
- [ ] If spike succeeds: HTTP fetcher — given make + model, attempt PDF download from identified sources in priority order
- [ ] Fetcher resilience — timeout per source (10s), try next source on failure, return not-found only after all sources exhausted
- [ ] `POST /home/documents/fetch` — accepts `{ make, model }`, runs fetcher synchronously, on success routes through same store → extract → index flow as upload
- [ ] Graceful failure response — explicit message when manual not found, prompt to upload manually
- [ ] UI loading state during fetch — same pattern as upload; success/failure feedback
- [ ] If spike fails or coverage is too low: stub the endpoint, return "manual fetch not available yet", ship without it — do not block Phase 0 on this

### Group 8 — Frontend
- [ ] Base layout — header, nav, main content area; htmx + Go templates
- [ ] Static asset serving (CSS, minimal JS)
- [ ] `GET /login` — login form; `GET /register` — register form; wire to auth endpoints
- [ ] Redirect to `/` after login; redirect to `/login` on unauthenticated access
- [ ] `GET /` — home page: document list (name, source, status) + search bar; empty state for new users
- [ ] `GET /upload` — upload form with file picker; loading state on submit; success/error feedback
- [ ] `GET /fetch` — make/model entry form; loading state on submit; success/error/not-found feedback
- [ ] Search bar on home page — submits to `GET /home/search`, renders results inline (htmx) or on separate page
- [ ] Search results — list of excerpts with document name and link; empty state message
- [ ] `GET /documents/:id` — document name, make/model if fetched, processing status, download link

---

## 9. Done Definition

Phase 0 is done when:

1. A user can register, log in, and log out
2. A user can create a home
3. A user can upload a PDF or image and it becomes searchable
4. A user can enter a make and model and get the manual indexed (or a clear failure message)
5. A user can search with a keyword and get back matching documents with excerpts
6. All of the above runs locally via `docker compose up`
7. The code is tested well enough that a refactor won't silently break the above
