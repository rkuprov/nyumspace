# nyumspace — Design & Implementation Plan

> **Status:** Draft
> **Last updated:** 2026-04-15

---

## Table of Contents

1. [Vision](#1-vision)
2. [Data Domains](#2-data-domains)
3. [Value Tiers & Differentiators](#3-value-tiers--differentiators)
4. [Critical Analysis](#4-critical-analysis)
5. [Users & Journeys](#5-users--journeys)
6. [Architecture](#6-architecture)
7. [Data Model](#7-data-model)
8. [Document Ingestion Pipeline](#8-document-ingestion-pipeline-temporal-workflow)
9. [API Design](#9-api-design)
10. [Infrastructure & Deployment](#10-infrastructure--deployment)
11. [AI / LLM Integration](#11-ai--llm-integration)
12. [Security Considerations](#12-security-considerations)
13. [Roadmap](#13-roadmap)
14. [Open Questions](#14-open-questions)

---

## 1. Vision

Homeowners, renters, and guests need a single, reliable place to access everything about a home — layouts, diagrams, appliance manuals, maintenance schedules, receipts, access codes, points of interest, and more.

**nyumspace** is that place. At its simplest, it's a searchable repository of everything about your home — upload a document or enter an appliance model number, and the system indexes it so you can find it when you need it. At its most powerful, it's an AI-powered home assistant that reads your documents, extracts structured data, answers natural language questions, and proactively reminds you when maintenance is due.

The product is designed around a core principle: **the home's knowledge belongs to the home.** It persists through ownership changes, accumulates over time, and is accessible to anyone with the right relationship to the property — owners, renters, guests — each seeing what's relevant to them.

### Core Value Propositions

- **Everything in one place** — all home documents, indexed and searchable, always available
- **Near-zero onboarding** — enter a make and model, the system fetches the manual for you
- **Effortless by design** — the system does the organizing, the user does the living
- **Knowledge that compounds** — every document, annotation, and service record makes the home profile more valuable over time, across owners

---

## 2. Data Domains

All home knowledge falls into five distinct domains. Every feature, API endpoint, and AI pipeline maps to one or more of these. They define what nyumspace *knows* and how it *knows* it.

### 2.1 Asset Registry

**"What do I have, and what's its status?"**

The living inventory of everything in and around the home — appliances, systems, fixtures — with lifecycle history attached.

- Do I have a warranty on my dishwasher? Is it expired?
- Who serviced the furnace last? What was their number? How much did it cost?
- What are the trees that are growing in my yard?

**Data sources:** Manuals, receipts, invoices, service records.

### 2.2 Operational Knowledge

**"How do I use, fix, or maintain this thing?"**

Procedures, manuals, error codes, troubleshooting — the kind of knowledge you need *right now* when something breaks or needs attention.

- My washer is giving me an error code — what does it mean?
- How do I reset my router?
- My furnace needs maintenance — what kind? Can I do it myself? What equipment do I need, or should I hire someone?
- How aggressively should I trim these trees?

**Data sources:** Manuals, guides, spec sheets.

### 2.3 Spatial Knowledge

**"Where is the physical thing?"**

Location-specific knowledge that can't be extracted from any document — it requires someone who's been in the house to capture it.

- Where is the main water line shutoff?
- Where is the irrigation system valve?
- Where is the water filter?

**Data sources:** User-captured photos, floor plan annotations, written descriptions.

### 2.4 Schedules & Triggers

**"What should I be doing, and when?"**

Maintenance intervals, seasonal tasks, replacement timelines.

- When should I turn the irrigation on and off?
- When should I change the air filter?
- When is the next HVAC service due?

**Data sources:** Manuals (maintenance intervals), service records, user-set reminders.

### 2.5 Agreements & Policies

**"What are the rules?"**

Leases, HOA bylaws, warranty terms, insurance policies — documents that need to be queried as full text, not just extracted into fields.

- Does my rental contract allow long-term guests?
- What does my home warranty actually cover?
- What are the HOA restrictions on exterior modifications?

**Data sources:** Leases, warranties, HOA docs, insurance policies.

---

## 3. Value Tiers & Differentiators

### 3.1 Tier 1 — Indexed Repository

The baseline product. A searchable index of everything about your home.

**What it does:**
- User uploads documents (manuals, receipts, leases) or enters appliance make/model numbers
- Documents are OCR'd (if scanned/photographed) and full-text indexed
- User searches using keywords and a search bar — traditional search experience
- When a user enters a make/model and the manual isn't uploaded yet, an agent fetches and downloads it from manufacturer sources on first lookup

**What it doesn't do:**
- No LLM-powered extraction into structured fields
- No natural language Q&A
- No proactive intelligence or suggestions
- The burden of lookup and judgment lies with the user

**Why this works as a standalone product:**
- Dead simple to understand — "all your home stuff, searchable"
- Near-zero marginal cost per user — storage and indexing only, no LLM API calls
- The manual-fetching agent solves the cold start problem — onboarding is "enter your appliance model numbers" not "find and upload all your PDFs"
- Valuable on its own even if the user never upgrades

### 3.2 Tier 2 — Home Assistant

The same document repository, now with an AI layer on top.

**What it adds beyond Tier 1:**
- LLM-powered structured extraction — upload a document, the system pulls out appliance details, warranty dates, maintenance schedules, and organizes them automatically
- Natural language queries — "when was the HVAC filter last replaced?", "what does error code F21 mean on my washer?"
- Proactive maintenance schedules derived from extracted data
- Notifications and alerts for upcoming maintenance, warranty expirations, seasonal tasks
- Human-in-the-loop review — AI extracts, user confirms or corrects

**What it could become (future):**
- Integration with smart home platforms — receives signals from devices, cross-references with documentation
- Plugin architecture — a home assistant that can be extended
- Local or cloud hosting options

### 3.3 The Differentiator

The differentiator is not features — it's **ease and smoothness of experience.**

The real competition isn't other home management apps. It's *doing nothing* — the inertia of "I'll just Google it when I need it." nyumspace only wins if the friction of using it is lower than the friction of not having it.

This means:
- The system infers structure; the user corrects it — never the other way around
- Minimal concepts to learn: you have a home, you put stuff in it, you find it later
- No setup wizard, no required fields, no mandatory taxonomy
- Very few knobs — every configuration option is an admission that the system couldn't figure it out on its own
- Onboarding is incremental — the home profile builds up through use, not upfront investment

### 3.4 Tier Relationship

Tier 1 is not a crippled version of Tier 2. It's a complete, useful product. Tier 2 layers intelligence on top of the same data. A user on Tier 1 already has all their documents indexed — upgrading to Tier 2 means those documents immediately become *smarter*, not that they need to start over.

---

## 4. Critical Analysis

Honest objections to this project, and how the current design addresses (or doesn't address) them.

### 4.1 "This is a solution looking for a problem."

**The objection:** How often does someone actually need their dishwasher warranty info? Once every 3-5 years? The washer error code moment is real but it happens twice a decade. You're building an always-on system for episodic needs. Google + the model number sticker solves 80% of these cases in 30 seconds.

**How the design addresses it:** Tier 1 makes the cost of "having it" near zero — for both the user and the platform. There's no LLM cost at rest, no complex infrastructure to maintain per user. It just sits there until you need it. The value isn't in daily use — it's in being there for the moments that matter, and in compounding over time. The home transfer story also means the *next* owner gets value from day one without doing the work.

**Remaining risk:** The episodic usage pattern means engagement metrics will look terrible by consumer app standards. This is a utility, not a habit product. Monetization needs to account for that.

### 4.2 "Nobody will do the onboarding."

**The objection:** "Upload and forget" sounds great, but the upload part is the problem. Someone has to find the manuals, scan them or find the PDFs, and upload them. That's an afternoon of work for a house. The people who would do this are the same people who already have a labeled binder in the closet. The people who need it most won't do it.

**How the design addresses it:** The manual-fetching agent eliminates the biggest friction. Onboarding drops from "find and upload 20 PDFs" to "enter 5-10 model numbers." That's a 10-minute task, not an afternoon. And it can be *incremental* — you don't need to catalog your whole house on day one. The furnace breaks, you enter the model number, the agent pulls the manual, now that appliance is covered. The home profile builds up organically through use rather than demanding upfront investment.

**Remaining risk:** The manual-fetching agent is non-trivial. Manufacturer sites are inconsistent, manuals may be behind registrations, PDFs may not exist for older appliances. The agent needs to be resilient and honest when it can't find something.

### 4.3 "The home-outlives-the-user principle is aspirational fiction."

**The objection:** For home transfer to work, the seller has to have used nyumspace, the buyer has to want to use it, and the transfer has to happen during the chaos of closing. The odds of all three being true are near zero for years. You're designing for a network effect that doesn't exist yet. In practice, every user starts from scratch.

**How the design addresses it:** Partially. The design acknowledges this by making "starting from scratch" a first-class journey, not a fallback. The transfer story is a long-term vision, not a launch requirement. But it shapes the data model early — home knowledge belongs to the home, not the user — so that when transfers do happen, the architecture supports it without retrofitting.

**Remaining risk:** This is genuinely years away from mattering. Don't over-invest in transfer mechanics early.

### 4.4 "You're building two products and calling it tiers."

**The objection:** A searchable document index and a conversational AI assistant are fundamentally different products with different UX, different infrastructure costs, different reliability requirements, and different user expectations. If Tier 1 is the free/cheap tier, you're giving away the useful part (indexed search) and charging for the expensive part (LLM queries) that most users may never reach.

**How the design addresses it:** The tiered split is now cleaner. Tier 1 has *no* LLM costs — it's OCR + full-text indexing. The economics work. Tier 2 adds LLM calls and needs to charge enough to cover them. The two tiers share a data layer (documents in storage, full-text index) so Tier 2 genuinely builds on Tier 1 infrastructure rather than being a parallel system.

**Remaining risk:** The UX for Tier 1 (search bar) and Tier 2 (conversational) are different interaction models. The upgrade path needs to feel like gaining a capability, not switching to a different app.

### 4.5 "The data model is deceptively complex."

**The objection:** "Appliance" sounds simple until you realize a central HVAC system spans the whole house, has sub-components (compressor, air handler, thermostat), has been partially replaced twice, and has three different service providers in its history. A "document" could be a 2-page receipt or a 200-page home inspection report that mentions every system in the house.

**How the design addresses it:** Not yet. This is a real problem that will surface during implementation. The design should resist the urge to model every edge case upfront and instead keep the data model simple, using freeform fields (notes, JSONB) as escape hatches until patterns emerge from real usage.

### 4.6 "LLM extraction is unreliable for the thing you need it most."

**The objection:** Extracting a brand name from a clean PDF manual is easy. Extracting the warranty expiration date from a photographed receipt with a coffee stain is where the user actually needs help — and where the LLM will fail. A wrong warranty date is worse than no warranty date.

**How the design addresses it:** Tier 2's human-in-the-loop review is the answer — AI extracts, user confirms. But the UX for this needs careful design. The system must clearly signal what was AI-extracted vs user-confirmed, and make correction effortless.

**Remaining risk:** Users won't review everything. The system needs a confidence signal and should only auto-commit high-confidence extractions, flagging uncertain ones for review.

### 4.7 "The competitive moat is thin."

**The objection:** Any company with an LLM API key can build document extraction. Apple could add this to Home. Google could add it to Nest. A Notion template with an AI plugin gets you 60% of the way there today. The moat is accumulated home data — but that takes years per user to build.

**How the design addresses it:** The moat isn't technology — it's the accumulated home profile and the transfer story. A home with five years of indexed documents, service records, and spatial annotations is genuinely hard to replicate. But this only works if data portability is offered (export), which builds trust and paradoxically increases retention.

### 4.8 "Who pays for this?"

**The objection:** Every Tier 2 document upload costs money (LLM API call). Every `/ask` query costs money. The users who upload the most and ask the most are your most expensive users. Your best customers are your worst margin.

**How the design addresses it:** Tier 1 has near-zero marginal cost — storage and indexing only. This makes a free or cheap tier economically viable. Tier 2 costs need to be passed to the user, which means the AI features need to be valuable enough to justify a subscription. Per-query LLM costs are the key number to model before building Tier 2.

---

## 5. Users & Journeys

### 5.1 Core Principle: The Home Outlives Its Users

The home is the first-class entity. Its knowledge — every manual, receipt, schedule, annotation, and service record — belongs to the home, not to any individual user. Users bind to a home with a role. When ownership changes, the knowledge stays. A home sale is not a data export — it's a transfer of the owner binding.

This means:
- A buyer inherits the full knowledge base the previous owner built up
- Service history, maintenance schedules, and spatial annotations persist across ownership
- The new owner can review, extend, or clean up — but the baseline is everything that came before
- A home's value in nyumspace compounds over time, regardless of who holds the title

### 5.2 User Journeys

**Owner — Buying a house (inheritance)**

The previous owner used nyumspace. At closing, they initiate a transfer. The buyer creates an account (or already has one) and accepts. Instantly, the buyer has the full home profile: every appliance with its history, every manual, every maintenance schedule, spatial annotations for shutoffs and valves, service provider contacts. The house "remembers" everything. The new owner picks up where the previous owner left off — they can update, extend, or reorganize, but they start with years of accumulated knowledge instead of a blank slate.

**Owner — Starting from scratch (onboarding)**

No previous owner data. The owner creates a home, names the rooms, and starts uploading. The first 10 minutes matter: they drop in 3-4 PDFs (furnace manual, dishwasher receipt, home inspection report), and within minutes the system has extracted appliance details, warranty dates, and maintenance schedules. The home already "knows" things. That moment — uploading a PDF and seeing structured data appear — is what hooks them.

**Owner — Something breaks**

The washer throws error code F21. The owner opens nyumspace and asks "what does error code F21 mean?" The system already knows the washer make and model (extracted from the manual uploaded months ago), retrieves the relevant section, and gives the answer with a reference to the source document. No Googling, no digging through drawers.

**Owner — Seasonal maintenance**

Spring arrives. nyumspace notifies the owner: turn on the irrigation system (here's where the valve is — photo attached), schedule HVAC service (last serviced 11 months ago by ABC Heating, their number is 555-1234), inspect the trees (the arborist recommended trimming the oak in spring). The owner didn't have to remember any of this — the system derived it from uploaded documents and previous service records.

**Owner — Renting out the property**

The owner invites a tenant and assigns the Renter role. The renter immediately has access to everything they need: WiFi credentials, air filter location and size, appliance operating basics, emergency shutoff locations, and house rules. The owner can scope what's visible — the renter doesn't see purchase prices, warranty claim details, or the owner's service provider contracts unless the owner chooses to share them.

**Renter — Day-to-day living**

The renter needs to change the air filter but doesn't know where it is or what size to buy. They ask nyumspace and get the answer (spatial annotation from the owner + specs extracted from the HVAC manual). Later, a friend wants to stay for a month — the renter asks "does my lease allow long-term guests?" and gets an answer with a citation from the lease document.

**Guest — Arriving at a short-term rental**

The guest receives a link or QR code from the host. No account required. They see a curated view: WiFi password, house rules, checkout instructions, appliance basics ("here's how the coffee maker works"), emergency contacts, and local recommendations. Nothing more. The experience is instant and disposable — they never need to log in again.

---

## 6. Architecture

### 6.1 High-Level Overview

```
┌─────────────┐     ┌──────────────┐     ┌──────────────────┐
│   Clients   │────▶│   API Layer  │────▶│  Business Logic   │
│ (Web / App) │     │   (REST/gRPC)│     │  (Go services)    │
└─────────────┘     └──────┬───────┘     └────────┬──────────┘
                           │                      │
                    ┌──────▼───────┐        ┌─────▼──────────┐
                    │   Auth /     │        │   Temporal      │
                    │   Sessions   │        │   Workflows     │
                    └──────────────┘        └────────┬────────┘
                                                     │
                    ┌────────────────────────────────┐│
                    │         Data Layer             ││
                    │  ┌──────────┐  ┌────────────┐  ││
                    │  │ Postgres │  │ Object     │  ││
                    │  │ (struct) │  │ Storage    │  │◀
                    │  └──────────┘  │ (S3/R2)   │  │
                    │                └────────────┘  │
                    │  ┌──────────────────────────┐  │
                    │  │ Vector Store (pgvector)  │  │
                    │  └──────────────────────────┘  │
                    └────────────────────────────────┘
```

### 6.2 Key Components

1. **API Service** — HTTP REST API (chi router), handles auth, request validation, routing
2. **Temporal Workers** — async workflow execution for document processing, notifications, scheduled tasks
3. **Document Ingestion Pipeline** — Temporal workflow that:
   - Accepts uploaded files (PDF, images, etc.)
   - Stores originals in object storage
   - Runs OCR / text extraction
   - Sends content to LLM for structured data extraction
   - Stores extracted entities in Postgres
   - Generates embeddings and stores in pgvector
4. **Query Engine** — handles natural language questions by:
   - Generating embeddings for the query
   - Retrieving relevant document chunks via vector similarity
   - Passing context + query to LLM for answer generation
5. **Notification Service** — Temporal scheduled workflows for maintenance reminders, warranty alerts, etc.

### 6.3 Architectural Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Already in use, strong concurrency, good for services |
| Workflow engine | Temporal | Already integrated, ideal for long-running document pipelines |
| Primary DB | PostgreSQL | Already in use, mature, supports pgvector |
| Vector search | pgvector extension | Avoids a separate vector DB, good enough for per-home scale |
| Object storage | S3-compatible (R2/MinIO/LocalStack) | Already in use, cost-effective for documents |
| LLM provider | Anthropic Claude API | Best-in-class for document understanding and extraction |
| Auth | TBD — sessions vs JWT vs OAuth | Needs decision |
| Frontend | TBD | Needs decision — web app, mobile, or both |
| API style | REST (primary) + gRPC (internal) | REST for clients, gRPC for internal service communication |

---

## 7. Data Model

### 7.1 Core Entities

```
User
├── id, email, name, created_at
└── roles[] (per-home)

Home
├── id, name, address, owner_id, created_at
├── members[] (user_id + role)
├── rooms[]
├── appliances[]
├── documents[]
├── codes[] (WiFi, alarm, gate)
└── schedules[]

Room
├── id, home_id, name, type, floor, notes
└── appliances[]

Appliance
├── id, home_id, room_id
├── name, brand, model, serial_number
├── category (HVAC, plumbing, electrical, kitchen, etc.)
├── purchase_date, warranty_expiry
├── specs (JSONB)
└── documents[] (linked manuals, receipts)

Document
├── id, home_id, uploaded_by
├── original_file_key (S3 reference)
├── file_type, file_name, file_size
├── processing_status (pending, processing, completed, failed)
├── extracted_text
├── extracted_entities (JSONB)
├── embeddings (pgvector)
└── linked_appliances[], linked_rooms[]

Schedule
├── id, home_id, appliance_id (optional)
├── title, description
├── type (maintenance, inspection, replacement, custom)
├── recurrence (cron expression or interval)
├── next_due, last_completed
└── notifications_enabled

Code
├── id, home_id
├── type (wifi, alarm, gate, lockbox, etc.)
├── label, value (encrypted)
└── visible_to_roles[]
```

### 7.2 Extraction Schema

When the LLM processes a document, it extracts structured data into a known schema:

```json
{
  "document_type": "manual | receipt | warranty | floor_plan | photo | other",
  "appliances": [
    {
      "name": "...",
      "brand": "...",
      "model": "...",
      "serial_number": "...",
      "specs": {},
      "maintenance_schedule": [
        { "task": "Replace filter", "interval_months": 3 }
      ],
      "warranty": { "expires": "2027-01-15", "provider": "..." }
    }
  ],
  "key_information": ["...", "..."],
  "dates": [{ "label": "...", "date": "..." }],
  "costs": [{ "label": "...", "amount": 0.00, "currency": "USD" }]
}
```

---

## 8. Document Ingestion Pipeline (Temporal Workflow)

```
Upload → Store Original → Extract Text → LLM Extraction → Store Entities → Generate Embeddings → Index
                                              │
                                              ▼
                                    Link to Appliances/Rooms
                                    Create/Update Schedules
```

### Workflow Steps

1. **UploadActivity** — validate file, store in S3, create Document record with `status=pending`
2. **TextExtractionActivity** — OCR for images/scanned PDFs, text extraction for digital PDFs
3. **LLMExtractionActivity** — send extracted text to Claude API with extraction prompt, parse structured response
4. **EntityStorageActivity** — upsert appliances, schedules, codes based on extracted data, link to document
5. **EmbeddingActivity** — chunk text, generate embeddings, store in pgvector
6. **IndexingActivity** — update search indices, mark document as `status=completed`

### Error Handling

- Each activity is independently retryable (Temporal handles this)
- LLM extraction failures → retry with exponential backoff, fall back to raw text storage
- Partial extraction → store what was extracted, flag for human review

---

## 9. API Design

### 9.1 Authentication

```
POST   /auth/register          — create account
POST   /auth/login             — login, receive session/token
POST   /auth/logout            — invalidate session
POST   /auth/refresh           — refresh token
```

### 9.2 Homes

```
POST   /homes                  — create a home
GET    /homes                  — list user's homes
GET    /homes/:id              — get home details
PUT    /homes/:id              — update home
DELETE /homes/:id              — delete home (owner only)

POST   /homes/:id/members      — invite member (with role)
PUT    /homes/:id/members/:uid  — change member role
DELETE /homes/:id/members/:uid  — remove member
```

### 9.3 Documents

```
POST   /homes/:id/documents            — upload document(s)
GET    /homes/:id/documents             — list documents
GET    /homes/:id/documents/:docId      — get document details + extracted data
GET    /homes/:id/documents/:docId/file — download original file
DELETE /homes/:id/documents/:docId      — delete document
POST   /homes/:id/documents/:docId/reprocess — re-run extraction
```

### 9.4 Appliances

```
GET    /homes/:id/appliances            — list appliances (auto-extracted + manual)
POST   /homes/:id/appliances            — manually add appliance
GET    /homes/:id/appliances/:appId     — get appliance details
PUT    /homes/:id/appliances/:appId     — update appliance
DELETE /homes/:id/appliances/:appId     — delete appliance
GET    /homes/:id/appliances/:appId/documents — documents linked to appliance
```

### 9.5 Rooms

```
GET    /homes/:id/rooms                 — list rooms
POST   /homes/:id/rooms                 — add room
PUT    /homes/:id/rooms/:roomId         — update room
DELETE /homes/:id/rooms/:roomId         — delete room
```

### 9.6 Schedules & Maintenance

```
GET    /homes/:id/schedules             — list schedules
POST   /homes/:id/schedules             — create schedule
PUT    /homes/:id/schedules/:schId      — update schedule
DELETE /homes/:id/schedules/:schId      — delete schedule
POST   /homes/:id/schedules/:schId/complete — mark as completed
```

### 9.7 Query / Ask

```
POST   /homes/:id/ask                   — natural language query about the home
       Body: { "question": "When should I replace the HVAC filter?" }
       Response: { "answer": "...", "sources": [...document references...] }
```

### 9.8 Codes

```
GET    /homes/:id/codes                 — list access codes (filtered by role)
POST   /homes/:id/codes                 — add code
PUT    /homes/:id/codes/:codeId         — update code
DELETE /homes/:id/codes/:codeId         — delete code
```

---

## 10. Infrastructure & Deployment

### 10.1 Local Development

Everything runs on minikube, deployed via a Helm chart:

```
  - PostgreSQL 16 StatefulSet (pgvector extension added when Tier 2 needs it)
  - LocalStack (S3-compatible storage)
  - API service
  - Temporal Server + worker(s) — introduced in Phase 1
```

### 10.2 Production Hosting — Options

| Option | Pros | Cons | Cost Estimate |
|--------|------|------|---------------|
| **Fly.io** | Simple deploys, global edge, built-in Postgres | Limited Temporal support, need to self-host Temporal | $-$$ |
| **Railway** | Easy Docker deploys, managed Postgres | Smaller ecosystem | $-$$ |
| **AWS (ECS/Fargate)** | Full control, Temporal Cloud available | Complex setup, higher ops burden | $$-$$$ |
| **GCP (Cloud Run)** | Serverless, scales to zero | Cold starts, Temporal self-hosted | $-$$ |
| **Hetzner + k3s** | Cheapest, full control | Most ops burden | $ |

### 10.3 Recommended Stack (Initial)

**Target: simple, low-cost, production-ready**

- **Compute:** Fly.io or Railway (containerized Go services)
- **Database:** Managed Postgres with pgvector (Neon, Supabase, or provider-managed)
- **Object Storage:** Cloudflare R2 (S3-compatible, no egress fees)
- **Temporal:** Temporal Cloud (managed) or self-hosted on same infra
- **LLM:** Anthropic Claude API (direct)
- **DNS/CDN:** Cloudflare

### 10.4 Deployment Strategy

- **Containerized** — single Dockerfile per service (API, worker)
- **CI/CD** — GitHub Actions → build → test → push image → deploy
- **Environments:** `dev` (local minikube) → `staging` → `production`
- **Database migrations** — goose, run as init container or pre-deploy step
- **Secrets** — provider's secret management (Fly secrets, Railway variables, AWS SSM)

---

## 11. AI / LLM Integration

### 11.1 Document Extraction

- **Model:** Claude (via Anthropic API)
- **Approach:** Structured extraction with tool use / JSON mode
- **Prompt design:** System prompt defines the extraction schema, user message contains the document text
- **Cost management:** Cache common system prompts, batch where possible, use Haiku for simpler extractions and Sonnet/Opus for complex documents

### 11.2 Natural Language Query (RAG)

- **Embedding model:** TBD (OpenAI ada-002, Cohere, or open-source via API)
- **Vector store:** pgvector (same Postgres instance)
- **Retrieval:** similarity search on embeddings, filtered by home_id
- **Generation:** Claude with retrieved context chunks as grounding
- **Guardrails:** only answer from retrieved context, cite sources

### 11.3 Smart Features (Future)

- Auto-detect duplicate appliances across documents
- Cross-reference warranty dates with purchase receipts
- Suggest maintenance schedules based on appliance type
- Seasonal maintenance checklists generated from home's appliance inventory

---

## 12. Security Considerations

- **Encryption at rest** — Postgres encryption, S3 server-side encryption
- **Encryption in transit** — TLS everywhere
- **Access codes** — encrypted in DB, decrypted only on read with proper role
- **Document access** — pre-signed S3 URLs with short TTL
- **Role enforcement** — middleware-level authorization checks
- **LLM data handling** — documents sent to Claude API are subject to Anthropic's data policy; consider implications for sensitive documents
- **Rate limiting** — per-user, per-home limits on API and LLM calls
- **Input validation** — file type restrictions, size limits, sanitization

---

## 13. Roadmap

### Phase 0 — MVP: Get Documents In, Find Them Again — [detailed plan](phase0.md)

The first deliverable. Validates the two core assumptions: does the indexing work, and do people find it useful enough to keep using? Everything else waits until this is answered.

**What's in:**
- [ ] Auth — single user, email/password, sessions. Nothing fancy.
- [ ] One home per user — create it, name it, done
- [ ] Document upload → OCR → full-text index
- [ ] Add manual by URL → downloads the PDF → indexes it (make/model lookup agent deferred to a later phase)
- [ ] Keyword search against the index, returns results with document references
- [ ] Local dev environment (minikube + Helm)

**What's explicitly out:**
- Rooms, structured appliance entities, schedules
- Roles and access control beyond single-user login
- Multiple homes
- Any LLM calls (no extraction, no RAG, no AI at query time)
- Production deployment

**What this teaches us:**
- Do people actually onboard? (paste-a-URL manual fetch lowers the friction; the make/model agent comes later)
- What do they search for? (informs Tier 2 design)
- Is OCR quality good enough for real documents?

### Phase 1 — Tier 1 Complete

Full Tier 1 product. Builds on the MVP foundation once the core assumptions are validated.

- [ ] Multiple homes per user
- [ ] Roles and access control (owner, renter, guest)
- [ ] Spatial knowledge — photo capture with captions, pinned to home
- [ ] Document management — delete, re-index, organize by type
- [ ] Production deployment — hosting, CI/CD, monitoring
- [ ] Guest access via link/QR code (no account required)

### Phase 2 — Tier 2: Intelligence Layer

The AI layer on top of the indexed repository.

- [ ] LLM-powered structured extraction on upload (appliance details, warranty dates, maintenance intervals)
- [ ] Human-in-the-loop review — AI extracts, user confirms or corrects
- [ ] Natural language query (`/ask`) via RAG
- [ ] Maintenance schedules derived from extracted data
- [ ] Notification system for upcoming tasks and warranty expirations

### Phase 3 — Tier 2: Expanded Intelligence

- [ ] Home transfer — reassign ownership, knowledge stays with the home
- [ ] Smart home platform integrations
- [ ] MCP exposed as a tool for the home assistant — same MCP from Phase 0 now callable by the AI layer
- [ ] Multi-home dashboard
- [ ] Mobile app or PWA

---

## 14. Open Questions

- [ ] **Auth strategy** — session-based vs JWT vs OAuth (Google/Apple sign-in)?
- [ ] **Frontend framework** — React, Next.js, SvelteKit, or mobile-first?
- [ ] **Embedding model** — which provider, self-hosted vs API?
- [ ] **Temporal hosting** — Temporal Cloud vs self-hosted? Cost?
- [ ] **Multi-tenancy model** — shared DB with row-level security vs separate schemas?
- [ ] **Sensitive documents** — do we need on-premise LLM option for users who don't want docs sent to API?
- [ ] **Pricing model** — free tier + paid? Per-home? Per-document?
- [ ] **Offline access** — do guests/renters need offline access to home info?
