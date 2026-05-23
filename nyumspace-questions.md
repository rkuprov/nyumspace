# nyumspace — Design Questions Framework

> A first-principles catalog of every question this project needs to answer before (and during) implementation.

---

## 1. Problem & Product

### 1.1 Problem Definition
- What specific pain does a homeowner feel today? ("I can't find the manual for my furnace" vs "I don't know when my roof warranty expires" — these drive different solutions)
- How do people currently solve this? (Filing cabinets, Google Drive folders, binder in the kitchen, nothing)
- What's the cost of the status quo? (Missed maintenance → expensive repairs, lost receipts → no warranty claims, guests calling at 2am asking for WiFi password)

### 1.2 User Personas & Journeys
- **Owner moving in:** Has a stack of manuals, receipts, and closing docs. What's the onboarding experience? How fast do they see value?
- **Owner mid-tenancy:** Furnace breaks. They need the model number and warranty info NOW. How fast can they get it?
- **Landlord onboarding a tenant:** What info does the tenant need? How does handoff work?
- **Guest arriving at Airbnb:** Needs WiFi, house rules, checkout instructions. What's the simplest possible access path? (Link? QR code? No account needed?)
- **Owner selling the house:** Can the entire home profile transfer to the buyer?

### 1.3 Competitive Landscape
- What exists? (HomeZada, Centriq, Notion templates, Google Drive, smart home apps)
- Where do they fall short?
- What's our differentiator? (AI extraction is the bet — but is it enough?)

### 1.4 MVP Definition
- What is the absolute minimum that delivers value to a single homeowner?
- Can we validate the AI extraction value prop without building the full platform?
- What can we cut from Phase 1 and still have something worth using?

---

## 2. Domain Model

### 2.1 Core Entities
- What is a "home"? (Single-family house, apartment, condo, multi-unit building — does the model change?)
- What is an "appliance" vs a "system" vs a "feature"? (Is a deck an appliance? A sprinkler system? A built-in bookshelf?)
- How granular are rooms? (Does "basement" have sub-areas? What about outdoor spaces?)
- What's the taxonomy of documents? (Manual, receipt, warranty card, inspection report, floor plan, photo, note — flat list or hierarchical?)

### 2.2 Relationships & Invariants
- Can an appliance belong to multiple rooms? (Ducted HVAC serves the whole house)
- Can a document relate to multiple appliances? (A home inspection report covers everything)
- What happens when an appliance is replaced? (History? Linked to old AND new?)
- What happens when a room is renovated? (Versioning? Before/after?)

### 2.3 Lifecycle
- What is the lifecycle of a document? (Uploaded → processing → indexed → archived → deleted)
- What is the lifecycle of a home? (Created → active → transferred → archived)
- Do we ever delete data? What's the retention policy?

---

## 3. AI / LLM Pipeline

### 3.1 Extraction Accuracy
- What accuracy rate is acceptable for structured extraction? (90%? 99%?)
- What happens when the LLM gets it wrong? (Extracts wrong model number, wrong date)
- **How do users correct AI mistakes?** This is a critical UX question — do they edit extracted fields directly? Flag errors? Re-upload?
- How do we measure extraction quality over time?

### 3.2 Document Types & Complexity
- What document types do we actually need to handle on day one? (PDFs only? Photos of labels? Handwritten notes?)
- How do we handle multi-appliance documents? (A home inspection report mentioning 30 items)
- How do we handle non-English documents?
- What about scanned/photographed documents with poor quality?

### 3.3 Cost & Latency
- What does it cost to process one document through Claude? (Input tokens for a 20-page manual ≈ $X)
- What's the acceptable latency for document processing? (Seconds? Minutes? Background is fine?)
- What's the acceptable latency for a `/ask` query? (Must feel conversational — under 3-5 seconds?)
- What's the monthly LLM cost for a typical user? (10 documents uploaded, 20 queries/month)
- How do we manage costs at scale? (Caching, model tiering, rate limits)

### 3.4 RAG Design
- What embedding model? What dimensions? What chunk size?
- How do we handle queries that span multiple documents? ("Compare the warranty terms of all my kitchen appliances")
- How do we handle queries with no relevant context? ("What color should I paint my bedroom?" — out of scope)
- Do we need conversation memory for follow-up questions?

### 3.5 Human-in-the-Loop
- After extraction, does the user review/confirm before data is committed?
- Or do we auto-commit and let users correct later?
- How prominent is the "AI extracted this, verify it" signal in the UI?

---

## 4. Architecture

### 4.1 System Boundaries
- Monolith or microservices? (For a small team / solo dev, a modular monolith is likely right — but where are the boundaries?)
- What runs synchronously in the request path vs asynchronously via Temporal?
- What's the read/write ratio? (Mostly reads — query, browse. Writes are bursty — onboarding, uploading docs)

### 4.2 Failure Modes
- What happens if the LLM API is down? (Document sits in queue, user sees "processing")
- What happens if object storage is down? (Can we still serve extracted data without originals?)
- What happens if Temporal is down? (API still serves reads, writes queue locally?)
- What's the blast radius of a bad deployment?

### 4.3 Scalability
- How many homes per user? How many documents per home? What's the ceiling?
- How large is a typical vector index per home? (Hundreds of chunks, not millions — pgvector is fine)
- What's the concurrent user expectation? (Hundreds? Thousands? Tens of thousands?)

### 4.4 Modularity
- Can the document pipeline run independently of the API?
- Can we swap the LLM provider without rewriting the pipeline?
- Can we add new extraction schemas without code changes? (Config-driven extraction?)

---

## 5. API Design

### 5.1 Consumer Contracts
- Who consumes the API? (Our own frontend only? Third-party integrations? Public API?)
- REST vs GraphQL? (REST is simpler; GraphQL is better if the frontend needs flexible queries over nested home→room→appliance graphs)
- API versioning strategy? (URL path `/v1/` vs header-based?)

### 5.2 Conventions
- Pagination: cursor-based or offset-based?
- Filtering and sorting: query params? OData-style? Custom?
- Error format: RFC 7807 (Problem Details)? Custom?
- Bulk operations: needed for initial onboarding? (Upload 15 documents at once)

### 5.3 Real-time
- Do we need WebSockets or SSE? (Document processing progress, real-time notifications)
- Or is polling sufficient for MVP?

---

## 6. Auth & Multi-tenancy

### 6.1 Authentication
- Email/password? OAuth (Google, Apple)? Magic link? Passkeys?
- What's the session model? (Stateful server sessions vs stateless JWT)
- How long do sessions last? Refresh tokens?

### 6.2 Authorization
- RBAC is clear at the home level (owner, manager, renter, guest) — but what about edge cases?
- Can a renter upload documents? (Maybe — receipts for repairs they did)
- Can a guest see maintenance schedules? (Probably not)
- How do we handle "shared but sensitive" data? (Access codes visible to renters but not guests)

### 6.3 Multi-tenancy
- Row-level security in Postgres? Application-level enforcement? Both?
- How do we prevent data leakage between homes/users?
- What about the LLM — do we include home_id context in every call to prevent cross-contamination?

---

## 7. Security & Privacy

### 7.1 Sensitive Data
- Documents may contain financial info, addresses, account numbers. Classification?
- Access codes stored in DB — encryption at the field level? Vault?
- What PII do we store? What's the GDPR/CCPA exposure?

### 7.2 Third-party Data Flow
- Documents are sent to Anthropic's API for extraction. Users need to understand this.
- Do we need a consent flow? ("Your documents will be processed by AI. They are not stored by the provider.")
- Should we offer an "AI-free" mode for privacy-sensitive users?

### 7.3 Attack Surface
- File uploads are a classic attack vector. Validation? Sandboxing?
- Pre-signed URLs for document access — TTL? Scoping?
- What if someone uploads malicious content designed to manipulate the LLM? (Prompt injection via document text)

---

## 8. Observability & Operations

### 8.1 Monitoring
- What metrics matter? (API latency, document processing time, LLM call duration/cost, error rates)
- What's the alerting strategy? (PagerDuty? Email? For a solo project?)
- How do we track LLM spend in real time?

### 8.2 Logging
- Structured logging? What format?
- How do we log LLM interactions without storing sensitive document content?
- Log retention policy?

### 8.3 Debugging
- How do we debug a failed extraction? (Need to see: original doc, extracted text, LLM prompt, LLM response, parse result)
- Temporal UI covers workflow debugging — but what about the API layer?

---

## 9. Testing

### 9.1 Strategy
- Unit tests for business logic
- Integration tests against real Postgres (already in CI)
- How do we test the LLM extraction pipeline? (Golden files? Recorded responses? Mock?)
- How do we test RAG quality? (Benchmark question set with expected answers?)
- End-to-end tests? (API → upload → extraction → query → answer)

### 9.2 AI-Specific Testing
- Regression suite of documents with known-good extraction results
- What's the acceptable drift when the model updates?
- Cost of running the full test suite against the real API?

---

## 10. Deployment & Infrastructure

### 10.1 Hosting Decision Factors
- Team size (solo / small → simplicity wins)
- Budget (side project vs funded startup — very different answers)
- Geographic requirements (US only? Global?)
- Compliance requirements (data residency?)

### 10.2 Operational Complexity Budget
- How much ops work is acceptable? (Managed everything vs roll-your-own)
- Is Temporal Cloud worth the cost vs self-hosting?
- Database backups — automated? Tested?

### 10.3 CI/CD
- What gates a deploy? (Tests pass, lint clean — what else?)
- Blue/green? Rolling? Canary?
- Rollback strategy?

---

## 11. Business & Growth

### 11.1 Monetization
- Free tier? What limits? (1 home, 10 documents, 5 queries/month?)
- Paid tier? What unlocks? (Unlimited docs, more queries, multiple homes)
- Per-query LLM costs make this a margin question — can we charge enough to cover Claude API costs?

### 11.2 Growth Vectors
- Organic: homeowner tells another homeowner
- Real estate transactions: transfer home profile with sale
- Property management: landlords managing multiple units
- Short-term rentals: Airbnb hosts creating guest portals

### 11.3 Data Moat
- The more documents a user uploads, the harder it is to leave
- Extracted structured data IS the product — not the raw documents
- Can we offer data export? (Should we, for trust? Must we, for GDPR?)

---

## 12. Comparison to nyumspace.md

### What nyumspace.md covers well
- Architecture diagram and component breakdown
- Data model with concrete schemas
- API surface (comprehensive endpoint list)
- Infrastructure options with tradeoffs table
- Phased roadmap
- Temporal workflow steps

### What nyumspace.md is missing or underweight

| Gap | Why it matters |
|-----|----------------|
| **User journeys / personas** | We defined roles but not *what they actually do*. The API should be derived from journeys, not the other way around. |
| **MVP scoping** | Phase 1 is still large. What's the true MVP — the smallest thing we can ship and learn from? |
| **AI accuracy & correction flows** | The extraction pipeline is described mechanically but doesn't address what happens when it's wrong. Human-in-the-loop is a core UX question. |
| **LLM cost modeling** | No estimates for per-document or per-query costs. This drives hosting and pricing decisions. |
| **Competitive analysis** | No mention of alternatives. Need to know what we're displacing. |
| **Failure modes** | No discussion of what breaks and how we handle it. |
| **Observability** | No monitoring, logging, or alerting strategy. |
| **Testing strategy** | No plan for testing AI components, which are inherently non-deterministic. |
| **Privacy / compliance** | Security section exists but doesn't address GDPR, consent for AI processing, or data export. |
| **Business model** | Pricing briefly mentioned in open questions but not explored. LLM costs make this critical. |
| **Real-time / notifications delivery** | Mentioned alerts but no discussion of delivery mechanism (email, push, SMS, in-app). |
| **Onboarding flow** | The first-use experience (uploading initial batch of documents) is the make-or-break moment and isn't designed. |
