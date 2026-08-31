# JobRadar — Living Technical Specification

Status: v0 design complete — implementation ready



## 1\. Project Goal

Build a personal job-market monitoring system that continuously discovers, collects, normalizes, deduplicates, filters, and tracks backend vacancies relevant to the owner, initially focused on Go/Golang and PHP roles.

The project must be useful as a real personal tool and, at the same time, provide practice with realistic backend engineering problems: external integrations, scheduled work, concurrency, retries, observability, data modeling, failure handling, and later AI-assisted analysis and application tracking.

The owner should implement the project independently. This document defines what to build, why each part exists, expected behavior, and important constraints. It intentionally avoids prescribing exact implementation code unless needed later.

## 2\. Core Principle

Do not build the system around LinkedIn scraping or anti-bot bypasses.

LinkedIn may be used as a discovery or benchmark channel through legitimate user-facing mechanisms such as job alerts, but direct scraping and automated interaction with LinkedIn should not be a dependency of the system.

Prefer direct company/ATS sources and official/public APIs where available.

## 3\. Phase 0 Goal

Phase 0 must answer one business/technical question:

> Can the system automatically discover and collect enough relevant vacancies without relying on unsafe LinkedIn scraping?

Success is not defined only by "the API works". The system should run continuously for a test period and produce measurable source coverage, overlap, freshness, discovery latency, and collector reliability.

## 4\. Phase 0 Scope

Included:

* Source/company discovery
* Source registry
* ATS detection and verification
* Job collection
* Raw payload storage
* Normalization
* Basic deduplication
* Rule-based relevance filtering
* Vacancy lifecycle tracking
* Source health tracking
* Metrics and observability
* Query API for collected jobs

Deferred:

* Automatic applications
* Browser automation for job applications
* AI matching/scoring
* Cover-letter generation
* Salary estimation
* Gmail-based application status tracking
* CV generation
* Company research
* Interview tracking
* Recommendation engine
* Kubernetes
* Kafka/event broker unless a real need emerges

## 5\. Initial Source Strategy

### 5.1 Direct ATS sources

Initial ATS families to support:

* Greenhouse
* Ashby
* Lever
* SmartRecruiters

Purpose:
Direct ATS sources are preferred because they are close to the employer's own publication source and can provide more reliable freshness than aggregators.

### 5.2 Discovery sources

Initial discovery candidates:

* Adzuna
* Jobicy
* LinkedIn Job Alerts as a benchmark/discovery channel

Discovery sources are not necessarily trusted as canonical truth. Their main purpose is to reveal new companies, vacancies, or ATS endpoints that the system does not already know.

## 6\. Discovery vs Ingestion

The system has two separate responsibilities.

### Discovery

Input: unknown companies, job URLs, aggregator results, job-alert results.

Goal: identify the employer and a reusable direct job source.

Flow:

Discovery input -> Company discovery -> Identity resolution -> Career-page resolution -> ATS detection -> Source verification -> JobSource Registry

### Ingestion

Input: verified JobSource records.

Goal: regularly collect job postings and transform them into normalized internal data.

Flow:

JobSource Registry -> Scheduler -> Collector -> RawJob -> Normalization -> Deduplication -> Filtering -> Job storage

These two flows should remain conceptually separate even if they live in one deployable application in Phase 0.

## 7\. Company and Source Discovery

Do not manually maintain hundreds of companies as the long-term strategy.

A small manual seed list is acceptable only for bootstrapping and testing collectors.

The system should gradually build its own registry of companies and direct sources.

### 7.1 Discovery lifecycle

Suggested states:

* DISCOVERED
* RESOLVING
* RESOLVED
* SOURCE\_FOUND
* NO\_SOURCE\_FOUND

### 7.2 Company identity

A canonical Company should be separate from source-specific names.

Suggested concepts:

Company:

* id
* canonical\_name
* website
* careers\_url
* created\_at
* updated\_at

CompanyAlias:

* company\_id
* name
* source

Normalization may initially remove common legal suffixes/punctuation and normalize case. Domain identity is a stronger signal than company-name similarity.

### 7.3 ATS detection

ATS detection should initially be rule-based, not LLM-based.

Examples of useful URL/domain fingerprints:

* Greenhouse
* Ashby
* Lever
* SmartRecruiters

Detection must not immediately create a trusted source.

The detected source must be verified by making a provider-specific request and validating that the response is structurally correct.

### 7.4 SourceCandidate

A detected source that has not yet been fully trusted.

Suggested fields:

* id
* company\_id
* detected\_type
* identifier
* url
* discovered\_from
* confidence
* status
* detected\_at
* verified\_at

Only verified candidates should become active JobSource records.

### 7.5 Fallback strategy for custom career sites

Preferred order:

1. Known ATS / public API
2. Structured JobPosting data (for example JSON-LD)
3. Stable JSON endpoint exposed by the careers site
4. Generic HTML parsing as the last fallback

HTML selectors should not be the default ingestion model because they are fragile and site-specific.

## 8\. Source Registry

Suggested JobSource fields:

* id
* company\_id
* type
* identifier
* url
* enabled
* status
* last\_sync\_at
* last\_success\_at
* last\_error\_at
* created\_at
* updated\_at

A company may have multiple sources over time or, in some cases, multiple simultaneous sources.

Source lifecycle should support failures and migrations, for example:
ACTIVE -> FAILING -> DISABLED

A disabled or repeatedly failing source may trigger source rediscovery for that company.

## 9\. Raw Ingestion

Always preserve provider payloads before normalization.

Suggested RawJob fields:

* id
* source\_id
* external\_id
* payload JSONB
* fetched\_at

Why:

* allows reprocessing if normalization logic changes
* helps debug provider-specific mapping bugs
* preserves fields that the first version of the normalizer may ignore
* makes migrations of internal schemas safer

## 10\. Normalized Job Model

Suggested initial fields:

* id
* company\_id
* title
* description
* location\_text
* country
* city
* remote\_type
* employment\_type
* salary\_min
* salary\_max
* salary\_currency
* salary\_period
* seniority
* published\_at
* first\_seen\_at
* last\_seen\_at
* status
* created\_at
* updated\_at

Unknown values must remain representable; do not invent data during normalization.

Possible remote\_type values:

* REMOTE
* HYBRID
* ONSITE
* UNKNOWN

Possible seniority values:

* JUNIOR
* MIDDLE
* SENIOR
* STAFF
* PRINCIPAL
* LEAD
* MANAGER
* UNKNOWN

## 11\. Job Occurrences / Provenance

A normalized job must be separate from the places where it was observed.

Suggested JobOccurrence fields:

* id
* job\_id
* source\_id
* external\_id
* source\_url
* published\_at
* first\_seen\_at
* last\_seen\_at
* raw\_job\_id

Reason:
The same vacancy can appear in an employer ATS, an aggregator, and a job alert. The system should preserve this provenance rather than create separate logical jobs for every occurrence.

## 12\. Deduplication

Phase 0 should prefer false negatives over false positives. Accidentally merging two different openings is worse than temporarily storing a duplicate.

Initial levels:

Exact:

* same source
* same external\_id

Strong candidate:

* same normalized company
* same normalized title
* compatible location

Probable candidate:

* same company
* similar title
* high description similarity

Probable matches should initially be marked for later review rather than automatically merged.

## 13\. Rule-Based Relevance Filtering

Do not use an LLM in Phase 0.

Initial target technologies:

* Go
* Golang
* PHP

Relevant titles may include:

* Backend Engineer
* Backend Developer
* Software Engineer
* Software Developer
* Go Developer
* Golang Developer
* PHP Developer

Title alone must not determine relevance. A generic "Software Engineer" role may be relevant if the description clearly uses target backend technologies.

Initially exclude clearly out-of-scope levels such as:

* Staff
* Principal
* Engineering Manager
* Director
* Head
* VP

Do not automatically exclude every Senior or Lead vacancy during Phase 0; title-based seniority is inconsistent across companies.

## 14\. Job Evaluation

Do not store relevance only as a boolean on Job.

Suggested JobEvaluation fields:

* job\_id
* relevant
* reason
* matched\_keywords
* excluded\_keywords
* evaluated\_at
* rules\_version

Reason:
Filtering rules will change. Versioning evaluations allows old jobs to be reprocessed and results compared across rule versions.

## 15\. Vacancy Lifecycle

A job disappearing from a source should not immediately be treated as permanently closed.

Suggested states:
ACTIVE -> MISSING -> CLOSED

Use a grace period to protect against temporary API outages or incomplete provider responses.

first\_seen\_at and last\_seen\_at must be preserved.

## 16\. Source Synchronization History

Suggested SourceSync fields:

* source\_id
* started\_at
* finished\_at
* status
* jobs\_received
* jobs\_new
* jobs\_updated
* duration\_ms
* error

This provides both operational visibility and input for source-quality metrics.

## 17\. Phase 0 Metrics

### Coverage

How many relevant jobs the system finds compared with the observed union of all benchmark/discovery channels.

### Overlap

Intersection between sources, for example ATS vs LinkedIn alerts or aggregator vs ATS.

### Unique contribution

Jobs contributed by a source that no other observed source discovered.

### Discovery latency

published\_at -> first\_seen\_at

### Freshness

How long sources continue to expose jobs that are already absent/closed at a more authoritative direct source.

### Duplicate rate

Occurrences compared with normalized unique jobs.

### Collector reliability

Successful syncs / total syncs.

### Relevant yield

Relevant jobs / total jobs fetched for each source or source family.

### Source discovery funnel

Companies discovered -> website resolved -> careers page found -> source detected -> source verified.

## 18\. Initial Operational Shape

Phase 0 should begin as one Go application or one codebase with clearly separated modules/processes, rather than artificial microservices.

Logical processes:

* Discovery
* Ingestion
* Processing
* API

Potential runtime roles may later be separated if real operational needs appear.

Initial infrastructure:

* Go
* PostgreSQL
* Docker / Docker Compose
* REST API
* structured logging
* Prometheus
* Grafana

Do not introduce Redis, Kafka, or Kubernetes until a concrete requirement justifies them.

## 19\. Initial Query Capability

The user should be able to inspect collected relevant jobs, for example by technology, relevance, state, and recency.

A first UI is optional; API access is sufficient for Phase 0.

## 20\. Decision Principles

* Prefer official/public APIs over scraping.
* Prefer direct employer/ATS data over aggregators when resolving freshness or canonical state.
* Preserve raw external data.
* Preserve provenance for every observed vacancy.
* Avoid premature distributed-system complexity.
* Measure source quality instead of assuming it.
* Implement only enough automation to validate the next uncertainty.
* The owner writes the implementation; this specification describes requirements, rationale, constraints, and expected behavior.

## 21\. Open Questions

* Exact process for resolving an employer website from an aggregator-only company name
* Initial supported countries/regions and remote rules
* Exact rules for MISSING -> CLOSED
* How to ingest LinkedIn Job Alert emails into the benchmark dataset
* Which custom career-site families are worth supporting after ATS coverage is measured



## 22\. Scheduling, Polling, Concurrency, and Retries

### 22.1 Scheduler

A scheduler is the component that decides when a known JobSource should be synchronized again. It does not know how a provider API works; it only decides which source is due and when it should run next.

The scheduler should initially be backed by PostgreSQL rather than by a separate queueing system.

Useful JobSource scheduling concepts:

* sync\_interval
* last\_sync\_at
* last\_success\_at
* next\_sync\_at
* consecutive\_failures
* enabled
* status

A basic query can select enabled sources whose `next\\\_sync\\\_at` is due.

### 22.2 Polling

Polling means periodically asking an external source whether vacancies have appeared, changed, or disappeared.

Initial polling should be configurable per provider/source family rather than hard-coded globally. Exact intervals should be chosen after checking provider rules, rate limits, source volume, and real vacancy activity.

Adaptive polling is deferred. A future version may poll high-activity sources more frequently and low-activity sources less frequently based on observed history.

### 22.3 Worker Pool

A worker pool is a bounded set of goroutines that consume synchronization tasks from a queue. It prevents both fully sequential collection and uncontrolled creation of one goroutine per source.

The system should have:

* a global concurrency limit
* provider-specific concurrency limits where necessary

The exact numbers are configuration values, not domain rules.

### 22.4 Rate Limiting

Rate limiting controls how many requests the application sends to an external service over time.

Each provider integration should respect the provider's documented rules and observable response behavior. A provider-specific limiter may be required even when the application still has unused global worker capacity.

HTTP 429 (Too Many Requests) must not be retried immediately in a tight loop. If the provider supplies `Retry-After` or equivalent guidance, it should be respected.

### 22.5 Retries

Retries are only for failures that are likely to be temporary.

Typical retry candidates:

* request timeout
* connection reset / temporary network failure
* HTTP 502, 503, 504
* HTTP 429, following provider guidance

Typical non-immediate-retry cases:

* invalid source identifier
* HTTP 400
* HTTP 401
* HTTP 403
* HTTP 404
* structurally invalid provider response

Exact classification may be provider-specific.

Retries must be bounded. The application must never retry a failed synchronization forever inside one run.

### 22.6 Exponential Backoff and Jitter

Backoff means waiting longer between repeated attempts after temporary failures. Exponential backoff grows the delay across attempts.

Jitter is a small random variation added to retry delays so that many workers that fail at the same time do not retry in lockstep and create another traffic spike.

### 22.7 Source Failure Lifecycle

Repeated failure should affect scheduling and source health.

Conceptual lifecycle:
ACTIVE -> FAILING -> DISABLED

The exact failure thresholds remain configurable.

A repeatedly failing or disabled source may trigger source rediscovery for the company, because the employer may have migrated from one ATS to another.

### 22.8 SourceSync Audit History

Each source synchronization attempt should be recorded.

Suggested SourceSync fields:

* id
* source\_id
* started\_at
* finished\_at
* status
* jobs\_received
* jobs\_new
* jobs\_updated
* attempt\_count
* http\_status
* error\_type
* error\_message
* duration\_ms

Purpose:

* debugging
* source-health monitoring
* reliability metrics
* identifying provider changes
* understanding why a source did not produce data

### 22.9 Idempotency

Idempotency means that safely repeating the same synchronization should not create incorrect duplicates or corrupt state.

At minimum, the pair `source\\\_id + external\\\_id` should make repeated observations of the same provider-side vacancy recognizable.

A partially completed synchronization followed by a retry must be safe.

### 22.10 HTTP Timeouts

Every external request must have a timeout. A stalled provider must not occupy a worker indefinitely.

Timeout values may differ by provider and should be configurable.

### 22.11 Graceful Shutdown

On process shutdown, the application should stop accepting new scheduled work, cancel or finish in-flight work according to context cancellation, shut down the HTTP server, and close database resources cleanly.

An interrupted synchronization must never be recorded as successful.

### 22.12 Deferred Reliability Mechanisms

Do not add these in Phase 0 unless evidence shows they are needed:

* distributed queue
* Redis-backed jobs
* Kafka/NATS
* circuit breaker
* distributed workers
* adaptive polling

## 23\. Operational Metrics for Scheduling and Collection

Initial useful metrics include:

* source\_sync\_total
* source\_sync\_failed\_total
* source\_sync\_duration\_seconds
* collector\_jobs\_received\_total
* collector\_jobs\_new\_total
* http\_requests\_total
* http\_request\_failures\_total
* rate\_limit\_hits\_total
* retry\_total
* sources\_due
* sources\_failing
* sources\_disabled

Metric names are illustrative; implementation naming may follow Prometheus conventions later.

## 24\. Data Model v0

The Phase 0 model separates external observations from the canonical vacancy used by JobRadar.

### 24.1 Company

Represents the canonical employer inside JobRadar.

Suggested fields:

* id
* canonical\_name
* website\_url
* careers\_url
* created\_at
* updated\_at

Do not add company enrichment such as employee count, funding, industry classification, logo, or long-form company research in Phase 0.

### 24.2 CompanyAlias

Stores alternative names used by external sources for the same company.

Example:

* Cloudflare
* Cloudflare Inc.
* Cloudflare, Inc

Suggested fields:

* id
* company\_id
* name
* normalized\_name
* source\_type
* created\_at

The purpose is company identity resolution and cross-source deduplication.

### 24.3 SourceCandidate

Represents a potential job source discovered but not yet trusted by the operational collector system.

Example: discovery finds `jobs.ashbyhq.com/acme`, but the application has not yet verified that the job board is valid.

Suggested states:
DISCOVERED -> VERIFYING -> VERIFIED / REJECTED

A verified candidate is promoted into JobSource. Keeping SourceCandidate separate prevents the operational registry from filling with broken or incorrectly detected sources.

### 24.4 JobSource

Represents a verified source that the scheduler is allowed to synchronize.

Suggested fields:

* id
* company\_id
* type
* identifier
* url
* status
* enabled
* sync\_interval
* next\_sync\_at
* last\_sync\_at
* last\_success\_at
* consecutive\_failures
* created\_at
* updated\_at

Likely unique key:
`company\\\_id + type + identifier`

### 24.5 SourceSync

Represents one synchronization attempt for a JobSource. It is operational history rather than the source itself.

Relationship:
JobSource 1 -> N SourceSync

The detailed history is useful for debugging, source reliability, provider-change detection, and later retention/aggregation policies.

### 24.6 RawJob

Stores provider-specific vacancy data before normalization.

Purpose:

* preserve original data
* debug mapping errors
* allow reprocessing when normalization logic changes
* detect provider schema changes

Repeated polling must not create unlimited copies of identical payloads. A future optimization may store versioned snapshots using a payload hash. Phase 0 may begin simpler, but the design should not assume every poll deserves a permanent duplicate JSON snapshot.

### 24.7 Job

Represents one logical vacancy in the market, independent of where JobRadar observed it.

Example: the same `Backend Engineer` position may be visible through the company's Greenhouse board, an aggregator, and a LinkedIn alert but still correspond to one Job.

Suggested fields:

* id
* company\_id
* title
* description
* location\_text
* country
* city
* remote\_type
* employment\_type
* seniority
* salary\_min
* salary\_max
* salary\_currency
* salary\_period
* published\_at
* first\_seen\_at
* last\_seen\_at
* status
* created\_at
* updated\_at

Do not create a uniqueness rule such as `company\\\_id + title`; a company may legitimately have multiple simultaneous positions with the same title.

### 24.8 JobOccurrence

Represents one source-specific observation of a logical Job.

Relationship:
JobSource -> JobOccurrence <- Job

Suggested fields:

* id
* job\_id
* source\_id
* external\_id
* source\_url
* published\_at
* first\_seen\_at
* last\_seen\_at
* status
* latest\_raw\_job\_id

Critical unique key:
`source\\\_id + external\\\_id`

This model preserves both facts: one logical vacancy exists, and it may have been observed through multiple sources.

### 24.9 Source Authority

Source authority describes how trustworthy a source is as the primary representation of a vacancy.

Conceptual priority:

1. Company ATS / official company careers source: HIGH
2. Aggregator: MEDIUM
3. Alert/discovery-only source: DISCOVERY

When conflicting descriptions or status information exist, the canonical Job should prefer higher-authority direct sources where available.

### 24.10 JobEvaluation

Represents JobRadar's current interpretation of a Job under a specific rules version.

Suggested fields:

* id
* job\_id
* rules\_version
* relevant
* matched\_keywords
* excluded\_keywords
* reason
* evaluated\_at

Likely unique key:
`job\\\_id + rules\\\_version`

Separating Job from JobEvaluation allows filtering logic to change without rewriting the historical fact that the vacancy existed.

### 24.11 Intentionally Deferred Entities

Do not add in Phase 0 unless requirements change:

* User
* Application
* CV
* CoverLetter
* Email
* Interview
* Technology / JobTechnology
* CompanyResearch
* SalaryEstimate
* AIAnalysis

Phase 0 is single-user and focuses on market acquisition quality.

## 25\. Vacancy Lifecycle and Consistency

This section defines how JobRadar behaves when the same vacancy is observed repeatedly, changes over time, appears through multiple sources, disappears, or is affected by collection failures.

### 25.1 First Observation

A collector fetches an external vacancy that JobRadar has never seen for that source.

Flow:

1. persist or register RawJob data
2. normalize provider-specific fields
3. check whether `source\\\_id + external\\\_id` already exists
4. if not, attempt cross-source deduplication against existing logical Jobs
5. create a new Job if no sufficiently strong match exists
6. create JobOccurrence linking the source observation to the Job
7. evaluate the Job using the current rules version
8. update source synchronization metrics

`first\\\_seen\\\_at` records when JobRadar first observed the vacancy. It must not be confused with provider `published\\\_at`, which may be missing, inaccurate, or older.

### 25.2 Repeated Observation Without Changes

If a later synchronization returns the same `source\\\_id + external\\\_id` and the normalized content has not materially changed:

* do not create a new Job
* do not create another JobOccurrence
* update occurrence `last\\\_seen\\\_at`
* update Job `last\\\_seen\\\_at` as appropriate
* preserve idempotency

Repeated polling is an observation of the same vacancy, not a new vacancy.

### 25.3 Existing Vacancy Changes

A provider may edit title, description, location, salary, or other fields while keeping the same external vacancy ID.

The occurrence remains the same because provider identity did not change.

JobRadar should:

1. detect material change
2. preserve provider-side raw information sufficiently for debugging/history
3. update normalized occurrence/canonical Job fields according to source authority
4. re-run JobEvaluation because relevance may have changed

Example: a generic `Software Engineer` vacancy may later add `Go` to its description and become relevant.

### 25.4 Same Vacancy Appears in a Second Source

Example:

* Greenhouse contains Acme / Backend Engineer / Warsaw
* Adzuna later contains a matching Acme / Backend Engineer / Warsaw vacancy

JobRadar should not immediately create a second logical Job if there is a strong identity match.

Expected result:

* one Job
* two JobOccurrences

The direct ATS occurrence should have higher authority than the aggregator occurrence for canonical description/status when they conflict.

### 25.5 Uncertain Duplicate

When similarity is plausible but insufficiently strong, Phase 0 should prefer a false negative over a false positive.

In other words, temporary duplicate Jobs are safer than incorrectly merging two genuinely different positions.

A future state or record may mark a `possible\\\_duplicate` for inspection, but Phase 0 must not require ML or LLM-based deduplication.

### 25.6 Missing from One Poll

A vacancy disappearing from one synchronization does not immediately mean that it is closed.

Possible causes:

* provider temporary inconsistency
* pagination or collection bug
* partial API response
* network/provider incident
* actual vacancy closure

Therefore absence from one successful poll should transition the source-specific occurrence toward MISSING rather than immediately closing the logical Job.

### 25.7 Collector Failure Is Not Vacancy Disappearance

This is a critical consistency rule.

If a source synchronization fails, JobRadar must not mark all previously active vacancies from that source as missing or closed.

Only a sufficiently complete successful synchronization may be used as evidence that a previously observed vacancy was absent from the source.

This prevents an API outage from falsely closing hundreds of vacancies.

### 25.8 Occurrence Lifecycle

Conceptual JobOccurrence lifecycle:
ACTIVE -> MISSING -> CLOSED

ACTIVE: the vacancy was observed in the latest trustworthy source state.

MISSING: the vacancy was absent from one or more successful source snapshots but closure is not yet considered reliable.

CLOSED: enough evidence exists that this source no longer exposes the vacancy, or the authoritative source explicitly indicates closure.

Exact grace-period rules should be configurable and finalized after real source behavior is observed.

### 25.9 Job Lifecycle

Logical Job state is derived from its occurrences rather than copied blindly from a single aggregator.

Conceptual Job lifecycle:
ACTIVE -> CLOSED

A Job remains ACTIVE while there is reliable evidence that at least one authoritative occurrence is active.

If an authoritative direct ATS occurrence closes while a lower-authority aggregator still lists the vacancy, JobRadar should prefer the direct source and treat the aggregator record as potentially stale.

When only lower-authority sources exist, closure should use a grace period and multiple successful observations rather than a single absence.

### 25.10 Source Migration

If a JobSource repeatedly fails or becomes invalid, this does not imply that the company stopped hiring.

Example:
Greenhouse source -> repeatedly invalid
-> source rediscovery
-> Ashby source discovered and verified

The old JobSource remains historical and may be disabled. The new source becomes operational. Existing Company identity remains unchanged.

### 25.11 Partial Processing Failure

Example:

* collector fetches 100 vacancies
* 70 are persisted/processed
* process crashes
* source is synchronized again

The repeated synchronization must safely process the same provider jobs without creating incorrect duplicate Jobs or JobOccurrences.

This requirement is the reason provider identity (`source\\\_id + external\\\_id`) and transactional boundaries are important.

Do not record SourceSync as SUCCESS until the synchronization has reached the defined successful completion point.

### 25.12 Transaction Boundaries

External HTTP requests must not be performed while a long-lived PostgreSQL transaction is held open.

Preferred conceptual separation:

1. perform provider request with timeout
2. obtain and validate response
3. begin bounded database work
4. persist/update observations consistently
5. commit

The exact batch/transaction strategy will be chosen during implementation after observing response sizes and failure modes.

### 25.13 Evaluation Consistency

A Job should be re-evaluated when material canonical fields that affect relevance change, such as:

* title
* description
* location when geography rules use it
* seniority

Re-evaluation creates or updates the evaluation associated with the active rules version. Historical rules versions should remain distinguishable.

### 25.14 Important Phase 0 Consistency Principles

1. Provider observation is not the same as logical vacancy identity.
2. Collector failure is not evidence of vacancy closure.
3. One missing poll is not sufficient closure evidence.
4. Direct company sources outrank aggregators where data conflicts.
5. Retrying the same external data must be safe.
6. Incorrectly merging two jobs is worse than temporarily keeping a duplicate.
7. External network operations and database transactions should have clear boundaries.
8. Relevance is derived state and may be recalculated independently of Job identity.

## Deduplication Strategy

Deduplication is the process of determining when multiple external job records represent the same real-world vacancy.

### Core Principle

Prefer false negatives over false positives. It is safer to temporarily keep two duplicate Job records than to merge two genuinely different vacancies into one.

### Matching Layers

1. Exact occurrence match: same JobSource + same external\_id. Always the same JobOccurrence.
2. Strong cross-source match: same canonical company, highly similar normalized title, compatible location, and supporting evidence such as matching source URL, description similarity, or publication timing.
3. Probable match: enough similarity to suspect a duplicate, but not enough confidence for automatic merge. Store as a duplicate candidate for later review or more advanced processing.

### Normalization

Before comparison, normalize company names, titles, location strings, whitespace, punctuation, case, and common legal suffixes such as Inc., Ltd., LLC, GmbH, and similar terms. Keep the original values separately.

### Important Signals

* Canonical company identity
* Normalized job title
* Location compatibility
* Remote/onsite mode
* Source URL or destination URL
* Description similarity
* Published date proximity
* Department/team when available
* Salary range when available

No single weak signal should be treated as proof of identity.

### Cases that must not be merged automatically

* Same company and same title, but clearly different locations or teams
* Same title with materially different descriptions
* Generic titles such as Software Engineer without supporting evidence
* Jobs with conflicting employment type, seniority, or responsibilities
* Situations where source data is incomplete or ambiguous

### Duplicate Candidates

For uncertain cases, create a duplicate candidate relation containing the two Job IDs, match score/confidence, reasons/signals, status, and timestamps. Possible statuses: PENDING, CONFIRMED\_DUPLICATE, NOT\_DUPLICATE.

### Merge behavior

When two Jobs are confirmed as the same vacancy, keep one canonical Job and reassign JobOccurrences to it. Preserve history; do not silently delete the losing Job without trace. A merge event/history record may be added later if operationally useful.

### v0 Decision

Start with deterministic rules and conservative fuzzy matching. Do not use LLMs for automatic deduplication in v0. More advanced text similarity or embeddings can be evaluated after collecting real duplicate examples.

## Filtering and Relevance (v0)

### Purpose

Filtering/relevance determines whether a normalized logical `Job` is potentially useful for the current search profile. It must not mutate or delete the `Job`; it creates an evaluation result that can be recomputed when rules change.

### Key distinction

* **Filtering**: cheap deterministic rules used to reject obviously irrelevant jobs.
* **Relevance evaluation**: rules that score or classify remaining jobs using multiple signals.

The v0 system uses deterministic rules only. LLM/ML-based matching is deferred.

### Search profile

The initial profile is single-user configuration, not a `User` domain model. It includes target technologies, allowed role families, excluded role/seniority signals, geographic preferences, and relocation/remote preferences.

### Positive signals

Examples: Go, Golang, PHP, backend, software engineer, software developer, backend engineer, backend developer.

A title alone is insufficient. `Software Engineer` may be relevant only when description/metadata contains target backend technologies.

### Negative signals

Examples of strong role-family exclusions: frontend-only, mobile-only, QA, data science, designer, product manager. Seniority exclusions for v0 should be conservative: Staff, Principal, Engineering Manager, Director, Head, VP are strong exclusions. `Senior` and `Lead` are not automatically rejected in v0.

### Evaluation stages

1. Validate minimum usable job data.
2. Apply hard exclusions where confidence is high.
3. Collect positive and negative signals from title, description, location, remote type, seniority and metadata.
4. Produce an evaluation result: `RELEVANT`, `NOT\\\_RELEVANT`, or optionally `REVIEW` for ambiguous cases.
5. Persist reasons and matched signals with a `rules\\\_version` so all jobs can be re-evaluated later.

### Why not use a boolean only

A bare `is\\\_relevant` loses reasoning and prevents safe rule evolution. `JobEvaluation` should preserve matched keywords, excluded keywords, reason and rule version.

### Title handling

Title is a high-value signal but not an identity or final relevance signal. Examples:

* `Go Developer` -> strong positive.
* `Backend Engineer` -> positive role family, requires technology evidence.
* `Software Engineer` -> ambiguous; description/metadata required.
* `Senior Backend Engineer` -> not rejected automatically in v0.

### Description handling

Description is used to detect target technologies and role context. Match whole tokens/phrases where practical to avoid false positives such as matching `go` inside unrelated words. HTML should be stripped or normalized before matching, while original text remains stored.

### Geography

Geography should not be an aggressive hard filter in v0 because relocation/sponsorship information is often incomplete. Keep explicit impossible regions configurable, but prefer classifications such as `MATCH`, `POSSIBLE`, `UNKNOWN`, `NO\\\_MATCH` rather than deleting jobs immediately.

### Remote and relocation

`remote\\\_type` is objective job data when known. `relocation` and `visa sponsorship` are often unknown and should not be guessed by deterministic rules. They can later be enriched by company/job analysis.

### Rules versioning

Every evaluation stores the rules version. When configuration changes, old jobs can be reprocessed and results compared without changing canonical job data.

### Example

`Backend Engineer`, description contains Go/Kafka/PostgreSQL, location Warsaw, seniority Senior -> likely `RELEVANT` in v0 even if title says Senior, because title seniority alone is not a hard exclusion.

`Principal Backend Engineer`, Go/PostgreSQL -> `NOT\\\_RELEVANT` because Principal is an explicit strong exclusion for the initial profile.

### Deferred

* LLM matching
* embeddings/vector search
* learned scoring model
* automatic relocation/sponsorship inference
* candidate-specific multi-user profiles



## REST API v0

### Purpose

The v0 REST API is an operational and inspection interface for the single-user JobRadar deployment. It exists so the owner can inspect discovered jobs, sources, source health, and evaluation results without connecting directly to PostgreSQL. It is not a public multi-user product API yet.

**REST API (Representational State Transfer Application Programming Interface)** — a conventional HTTP interface where clients read or change resources through endpoints such as `GET /jobs`.

### Scope principles

* Keep the API small and read-heavy in v0.
* Do not expose internal provider payloads by default.
* Do not design authentication, roles, tenants, or public API compatibility yet.
* Prefer filters/pagination over many specialized endpoints.
* Operational actions that can damage state should be omitted until a real need appears.

### Jobs

`GET /api/v1/jobs`

Purpose: list normalized logical jobs with filtering and pagination.

Initial useful filters:

* `evaluation=RELEVANT|REVIEW|NOT\\\_RELEVANT`
* `status=ACTIVE|CLOSED`
* `technology=go|php`
* `company\\\_id=<id>`
* `source\\\_type=ASHBY|GREENHOUSE|LEVER|...`
* `remote\\\_type=REMOTE|HYBRID|ONSITE|UNKNOWN`
* `first\\\_seen\\\_from=<timestamp>`
* `first\\\_seen\\\_to=<timestamp>`
* `page`
* `page\\\_size`

Default ordering for the personal workflow should favor newest discovered jobs (`first\\\_seen\\\_at DESC`).

`GET /api/v1/jobs/{id}`

Purpose: inspect one logical job in detail. Response should include canonical normalized data, current evaluation, source occurrences, important timestamps, and source URLs.

The endpoint should make it possible to answer:

* What is this vacancy?
* Why does JobRadar consider it relevant/review/not relevant?
* Where was it found?
* Which source is authoritative?
* When was it first and last seen?

### Companies

`GET /api/v1/companies`

Purpose: inspect companies discovered by the system. Keep filters minimal in v0 (name/search and pagination).

`GET /api/v1/companies/{id}`

Purpose: inspect canonical company identity, aliases, careers URL, and registered job sources.

### Job sources

`GET /api/v1/sources`

Purpose: inspect the source registry and operational health.

Useful filters:

* `type`
* `status`
* `enabled`
* `company\\\_id`
* `failing=true|false`

`GET /api/v1/sources/{id}`

Purpose: inspect one verified source including provider type, identifier, sync schedule, last success, consecutive failures, and recent sync history summary.

`GET /api/v1/sources/{id}/syncs`

Purpose: inspect SourceSync audit history for debugging provider failures and scheduler behavior.

### Discovery/source candidates

`GET /api/v1/source-candidates`

Purpose: inspect unresolved or rejected source candidates discovered by the discovery subsystem. This is primarily a debugging/quality endpoint for Phase 0.

Useful filters:

* `status=DISCOVERED|VERIFYING|VERIFIED|REJECTED`
* `detected\\\_type`
* `company\\\_id`

### Phase 0 metrics/reporting

`GET /api/v1/stats/overview`

Purpose: return a small operational summary such as jobs discovered, currently active jobs, relevant/review counts, known companies, verified sources, and failing sources.

More specialized coverage/overlap reports should be added only once the Phase 0 experiment data exists. The API should not invent a generic analytics platform before there are real questions to answer.

### Raw payload access

Raw external payloads are stored for debugging/reprocessing but should not be returned in normal job responses. If raw inspection becomes necessary, expose it only through a dedicated debug/admin endpoint or inspect it directly during development.

Reason: raw payloads can be large, provider-specific, unstable, and may contain fields not intended for the primary API contract.

### Pagination

All collection endpoints should be paginated. Exact implementation (`page/page\\\_size` versus cursor pagination) is an implementation decision. For v0 and one user, offset pagination is acceptable unless real data volume shows a problem.

### Error response

Use one consistent application error envelope instead of provider-specific errors leaking through HTTP. A response should communicate at least:

* stable application error code
* human-readable message
* optional details useful for debugging

Provider responses such as Ashby/Greenhouse HTTP errors should be translated into JobRadar-level errors/logs rather than returned verbatim to API clients.

### API versioning

Use `/api/v1` from the beginning. This does not mean the API is public or guaranteed stable; it avoids renaming every route later when a second contract is introduced.

### Mutating endpoints intentionally deferred

Do not add in v0 unless implementation/testing creates a real operational need:

* manual create/update/delete Job
* manual merge/unmerge Job
* arbitrary raw payload edits
* public source registration
* user/profile/auth management
* application/cover-letter endpoints

A narrowly scoped internal action such as manually triggering a source sync can be added later for development/operations, but it is not required for the first complete v0 specification.

## Observability v0

### Purpose

Observability is the ability to understand what the system is doing from the outside by using logs, metrics, traces/profiles, and operational history. In JobRadar this is not decoration: collectors depend on external services, scheduled work runs without a user watching it, and failures can otherwise silently reduce market coverage.

The v0 observability goal is to answer practical questions:

* Are collectors running?
* Which sources are failing or slow?
* Are jobs still being discovered?
* Did a provider start returning unexpected data?
* Is the application leaking goroutines or consuming unusual CPU/memory?
* How long do syncs and processing stages take?

### Structured logging

Use structured application logs rather than free-form text only. In Go, `log/slog` is sufficient for v0.

A useful log event should include stable fields where applicable:

* `component` (scheduler, collector, discovery, deduplicator, evaluator, api)
* `source\\\_id`
* `source\\\_type`
* `company\\\_id`
* `sync\\\_id`
* `job\\\_id` / `external\\\_id` when relevant
* `duration\\\_ms`
* `error\\\_type`
* request/correlation identifier for API requests

Do not log full job descriptions, full raw provider payloads, secrets, tokens, email addresses, authorization headers, or other unnecessary personal/sensitive data.

Logging levels:

* DEBUG: detailed development diagnostics; normally disabled in production
* INFO: important normal events such as application startup, source sync completion, source registration
* WARN: recoverable/suspicious situations such as unexpected empty responses, retries, repeated source failures
* ERROR: operations that failed after normal recovery/retry handling

Avoid logging every successfully processed job at INFO level. At scale this creates noise. Prefer one summary event per sync plus detailed DEBUG logs when needed.

### Metrics with Prometheus

Prometheus is a monitoring system that periodically reads numeric metrics exposed by an application and stores them as time series. JobRadar should expose a `/metrics` endpoint for Prometheus scraping.

Core v0 counters:

* `jobradar\\\_source\\\_sync\\\_total{source\\\_type,status}`
* `jobradar\\\_source\\\_sync\\\_failures\\\_total{source\\\_type,error\\\_type}`
* `jobradar\\\_jobs\\\_received\\\_total{source\\\_type}`
* `jobradar\\\_jobs\\\_created\\\_total{source\\\_type}`
* `jobradar\\\_jobs\\\_updated\\\_total{source\\\_type}`
* `jobradar\\\_retry\\\_total{source\\\_type,reason}`
* `jobradar\\\_rate\\\_limit\\\_hits\\\_total{source\\\_type}`
* `jobradar\\\_job\\\_evaluation\\\_total{result}`

Core gauges:

* `jobradar\\\_sources\\\_active`
* `jobradar\\\_sources\\\_failing`
* `jobradar\\\_sources\\\_disabled`
* `jobradar\\\_sources\\\_due`
* worker queue depth if an in-process queue exists
* active workers/in-flight source syncs

Core histograms:

* source sync duration
* provider HTTP request duration
* normalization/processing duration if useful
* API request duration
* discovered-job delay when `published\\\_at` is available (`first\\\_seen\\\_at - published\\\_at`)

Metrics must avoid high-cardinality labels. Do not use `job\\\_id`, `source\\\_id`, company name, URL, or arbitrary error text as Prometheus labels. These can create an unbounded number of time series. Use bounded dimensions such as provider type, status, result, HTTP class, or normalized error type.

### Grafana

Grafana is a dashboard/visualization system that can query Prometheus and display operational charts. It should be used in v0 because JobRadar is a background data pipeline and visual monitoring helps validate that it keeps working over days/weeks.

First dashboard should focus on a small operational view:

* successful vs failed syncs over time
* sync duration by provider
* jobs received/new over time
* active/failing sources
* retries and HTTP 429 rate-limit events
* relevant/review/not-relevant evaluations
* discovery latency where available

Do not create many decorative dashboards before real operational questions appear.

### SourceSync vs Prometheus

`SourceSync` records and Prometheus metrics serve different purposes.

`SourceSync` is durable business/operational history in PostgreSQL. It can answer: "What happened to source X at 10:02 on July 29?"

Prometheus is aggregated time-series monitoring. It can answer: "Did the failure rate for Ashby increase during the last six hours?"

Keep both. Prometheus must not replace durable sync history, and PostgreSQL should not be used as a substitute for all monitoring charts.

### Health and readiness endpoints

Expose small operational endpoints separate from the product REST API:

* `/health/live`: process is running
* `/health/ready`: application is ready to serve/work (for example, required initialization succeeded and PostgreSQL is reachable according to the chosen readiness policy)
* `/metrics`: Prometheus metrics

Liveness should not perform expensive external-provider checks. A temporary Ashby outage must not make the JobRadar process itself "dead".

### Profiling with pprof

`pprof` is Go's built-in profiling tooling for inspecting CPU usage, heap allocations, goroutines, mutex contention, and other runtime behavior.

JobRadar is a useful project for pprof because collectors and worker pools can expose practical problems such as:

* goroutine leaks
* excessive allocations while parsing large descriptions/raw JSON
* unexpectedly expensive deduplication
* excessive CPU usage during normalization
* blocked workers / mutex contention

Enable pprof for development and controlled production diagnostics, but do not expose profiling endpoints openly to the public internet. Bind them to an internal interface, protect access, or tunnel to them during diagnostics.

### Distributed tracing / OpenTelemetry

OpenTelemetry is a standard ecosystem for traces, metrics, and logs. A trace follows one operation through multiple components/services.

For v0, full distributed tracing is intentionally deferred because JobRadar is initially one deployable Go application, so there is little distributed call graph to inspect. Structured logs + `SourceSync` + Prometheus + pprof provide more value initially.

Keep boundaries trace-friendly: propagate `context.Context`, use request/sync identifiers, and isolate external HTTP calls. OpenTelemetry can be added later when the system splits into separate processes/services or LLM/email/application workflows make end-to-end traces valuable.

### Alerting

Do not build a complex alerting system for v0. The first production deployment may add a few useful alerts after real baselines exist, for example:

* no successful syncs for a long period
* unusually high source failure rate
* source registry has many failing/disabled entries
* discovery/job ingestion unexpectedly drops to zero

Exact thresholds must come from observed behavior rather than invented numbers.

### Observability decisions for v0

Required:

* structured logging using Go `log/slog` or equivalent
* durable `SourceSync` operational history
* Prometheus `/metrics`
* one useful Grafana operational dashboard
* liveness/readiness endpoints
* pprof available for controlled diagnostics
* correlation/sync identifiers in logs

Deferred:

* full OpenTelemetry tracing pipeline
* distributed trace backend such as Tempo/Jaeger
* complex alert manager rules before baselines exist
* log aggregation stack (Loki/ELK) unless deployment experience shows a need
* custom observability platform

## Testing Strategy v0

Testing focuses on business-risk boundaries rather than chasing a coverage percentage. Layers: unit tests for deterministic rules (normalization, relevance, retry/error classification, lifecycle decisions); repository integration tests against real PostgreSQL for constraints, transactions, migrations, idempotency, and queries; provider-adapter tests with local HTTP test servers and stored fixtures for valid responses, malformed payloads, timeouts, 429/5xx, empty responses, and schema drift; pipeline scenario tests for first discovery, repeated sync, changed job data, cross-source duplicates, failed syncs, missing/closed transitions, and restart/reprocessing safety; smoke tests after deployment for readiness, DB connectivity, scheduler startup, metrics exposure, and at least one controlled source sync.

External live APIs are not required for normal CI because they are nondeterministic and may be rate-limited or unavailable. A small optional manual/periodic contract check may verify that supported public provider endpoints still match expected schemas.

Tests must prioritize cases where a mistake would silently lose jobs, create duplicates, close live jobs, or misclassify relevant vacancies. Coverage percentage is secondary; critical-path behavior is mandatory.

## Phase 0 Experiment and Definition of Done

### Purpose

Phase 0 is not judged by whether the API starts or whether all planned components exist. Its purpose is to validate the product hypothesis: JobRadar can autonomously discover enough relevant backend vacancies from allowed sources, keep the data fresh, and do so reliably enough to become a useful daily tool.

### Experiment duration

After v0 is deployed and stable enough for normal operation, run a continuous observation period of approximately 7-14 days. The period must be long enough to include repeated polling cycles, new vacancy publication, source failures/recovery, duplicates across providers, and vacancy closures or edits.

### Target search scope

The first experiment focuses on backend/software roles with Go/Golang and PHP signals. Strong title exclusions include Staff, Principal, Engineering Manager, Director, Head, and VP. Senior and Lead are not automatic exclusions in v0. Geography includes remote roles and selected countries/markets where relocation or cross-border employment may be possible; unclear relocation/authorization is recorded as unknown or review, not guessed as false.

### Benchmark / reference set

LinkedIn is not scraped. It may be used as an external comparison channel through user-facing job alerts/email or manual observation. The benchmark is not treated as the complete market; it is one independent reference source used to measure overlap and missed opportunities.

The observed universe is the union of relevant vacancies found by JobRadar and the independent benchmark/reference channels during the same period.

### Required measurements

At minimum, record:

* total unique relevant/review jobs found by JobRadar
* unique relevant jobs seen in benchmark/reference channels
* intersection between JobRadar and benchmark
* JobRadar-only jobs
* benchmark-only jobs
* unique contribution per provider/source type
* discovery latency where a trustworthy `published\\\_at` exists
* source sync success/failure rate
* duplicate candidate / deduplication outcomes
* stale vacancy examples where an aggregator conflicts with a direct ATS
* source discovery funnel: companies discovered -> website/career source resolved -> source candidate detected -> source verified
* number of active/failing/disabled sources
* relevance result counts: RELEVANT / REVIEW / NOT\_RELEVANT

### Coverage metric

Do not claim coverage of the entire job market. Phase 0 can only measure coverage against the observed universe.

Example:

* JobRadar relevant unique jobs: 240
* benchmark relevant unique jobs: 200
* overlap: 160
* observed unique universe: 280
* JobRadar observed coverage: 240 / 280 = 85.7%

This number is evidence about the experiment, not proof that 85.7% of all real vacancies were found.

### Source discovery success

Measure how well the registry grows without manual curation.

Example funnel:

* 1,000 companies discovered
* 900 websites resolved
* 820 career pages or ATS links found
* 700 source candidates detected
* 670 verified usable sources

The exact target percentage is intentionally not fixed before real data exists. The important requirement is that source discovery produces a meaningful and growing direct-source registry rather than depending on a manually maintained list.

### Reliability requirements

Phase 0 is not successful if it finds many jobs but silently corrupts state. The following invariants must hold during the experiment:

* repeated processing of the same source/external ID does not create duplicate occurrences
* failed/invalid syncs do not mark all existing jobs missing or closed
* uncertain cross-source duplicates are not aggressively merged
* a temporary provider failure does not permanently disable valid jobs
* direct/authoritative ATS data has priority over stale aggregator data according to source policy
* invalid provider responses are recorded as failed/suspicious syncs rather than successful empty datasets
* restart/reprocessing does not materially change counts except for legitimate new/updated data

### Operational usefulness

The v0 should already be useful to the owner without direct SQL access. A normal daily workflow should allow the owner to open the API/UI client and answer:

* what new potentially relevant jobs appeared since the last check?
* where were they found?
* why were they classified RELEVANT or REVIEW?
* is the source still active/fresh?
* are collectors operating normally?

A polished frontend is not required; REST API plus a practical client/tool is sufficient for Phase 0.

### Definition of Done for v0

v0 is considered implemented when all of the following are true:

1. The project runs locally through documented setup using Go, PostgreSQL, and Docker-based dependencies.
2. Database migrations create the full v0 schema from an empty database.
3. At least the first planned direct provider collectors work end-to-end (initially Greenhouse, Ashby, Lever; additional providers may be added according to implementation roadmap and policy validation).
4. Discovery can create source candidates and verified JobSources without requiring every company to be manually entered.
5. Scheduler/worker processing performs repeated polling with bounded concurrency, timeouts, retries/backoff, and graceful shutdown.
6. Raw/provider data can be normalized into the canonical Job/JobOccurrence model.
7. Repeated syncs are idempotent.
8. Lifecycle rules distinguish successful absence from failed syncs and support ACTIVE/MISSING/CLOSED behavior without mass false closures.
9. Cross-source deduplication supports exact/high-confidence automatic matches and conservative possible-duplicate handling.
10. Deterministic relevance evaluation produces RELEVANT/REVIEW/NOT\_RELEVANT with reasons and rules versioning.
11. REST endpoints allow inspection of jobs, companies, sources, sync history, source candidates, and overview statistics.
12. Structured logs, Prometheus metrics, health/readiness, Grafana, and controlled pprof access are available.
13. Critical business invariants are covered by unit/integration/pipeline tests and `go test -race` is used for concurrency-sensitive code.
14. The application can be deployed as one Docker Compose stack on a VPS with PostgreSQL persistence, secrets kept outside images/repository, and internal-only DB/metrics/profiling surfaces.
15. A 7-14 day Phase 0 observation can be run and produce the coverage/source/reliability metrics described above.

### Success vs failure of the product hypothesis

Implementation DoD and product success are different.

The implementation can be technically complete even if the experiment shows poor market coverage. In that case Phase 0 is still valuable: it proves that acquisition/source coverage, not backend implementation, is the current bottleneck.

After the experiment, classify the result:

* **Promising:** JobRadar finds a substantial share of the observed relevant universe, direct-source registry grows automatically, and reliability is good enough for daily use. Continue to AI matching/application assistance.
* **Mixed:** the system is reliable but coverage is weak or discovery depends too much on aggregators/manual work. Improve acquisition/source discovery before building auto-apply features.
* **Not viable under current acquisition model:** coverage or allowed-source access is too poor for the system to save meaningful search effort. Revisit sources/product scope rather than adding more downstream features.

Do not invent a fixed percentage threshold before observing the market. The final go/no-go decision should consider coverage, unique JobRadar-only discoveries, discovery latency, source quality, and the amount of manual effort still required.

# Final v0 Review

Status after review: **v0 specification complete enough to begin implementation.**

The final review found one documentation gap: configuration/secrets, deployment/security, and external-source access policy had been discussed but were not present as dedicated final sections. They are consolidated here before implementation starts.

## Configuration and Secrets v0

Use environment-based configuration, safe defaults for non-secret operational values, and startup validation. Configuration covers application/logging, HTTP timeouts, PostgreSQL connection/pool settings, scheduler/worker limits, provider-specific timeout/rate/retry/polling settings, and observability settings.

Secrets include database credentials and provider/API credentials. Never commit them, copy them into Docker images, log them, or return them through APIs. Local development may use `.env`; commit only `.env.example`.

Fail fast on invalid required configuration. Provider credentials are required only when that provider is enabled. Search preferences such as technologies, seniority, geography, and relocation are domain/SearchProfile data, not infrastructure configuration long-term.

A JobSource may optionally override the provider default polling interval; otherwise it inherits that default.

## Deployment and Basic Security v0

Initial production shape: one VPS with Docker Compose running the Go application, PostgreSQL, Prometheus, Grafana, and optionally Caddy/reverse proxy for HTTPS.

PostgreSQL, Prometheus, and pprof are not publicly exposed. Prefer internal Docker networking plus VPN/SSH tunnel/private access. If the REST API is public, use HTTPS and appropriate access protection; a multi-user auth platform is not required for personal v0.

Use a multi-stage Docker build and a non-root runtime user. Keep secrets outside images. Persist PostgreSQL data and establish basic backups once the database contains useful market history.

Use versioned database migrations and identifiable application builds/commits. Graceful shutdown must stop new scheduled work, cancel/finish in-flight work under deadlines, stop HTTP, and close database resources.

External HTTP access uses timeouts, response-size limits, redirect limits, and validation. Generic crawling must defend against SSRF: reject loopback/private/link-local destinations and revalidate redirects.

Kubernetes, Kafka, service mesh, autoscaling, multi-region infrastructure, and Terraform are not v0 requirements.

## External Source Access Policy v0

Every provider/source family needs a documented access policy before automated collection is enabled.

Prefer documented public job APIs, public employer job boards, appropriate public career pages, and legitimate user-facing alerts/feeds for discovery or benchmarking.

Do not build authenticated scraping, anti-bot bypass, CAPTCHA solving, fake accounts, access-control circumvention, hidden/internal vacancy discovery, or private-profile scraping.

LinkedIn is not a direct scraper dependency in v0. User-facing alerts/email or manual observation may be used for discovery/benchmarking.

A CAPTCHA/access challenge or explicit automation restriction makes a source unsupported/manual unless an allowed official integration exists. Provider rate limits and published access constraints override JobRadar polling preferences.

JobRadar v0 is read-only toward vacancy providers. Application submission is a separate future problem.

For every new provider document: what it is, why JobRadar uses it, access method, authentication, obtained data, restrictions, rate-limit/polling considerations, allowed operations, and prohibited/deferred operations.

# v0 Architecture Baseline

The first implementation is a **modular monolith**, not microservices.

```text
Discovery inputs
      ↓
Company / source discovery
      ↓
SourceCandidate verification
      ↓
JobSource Registry
      ↓
Scheduler
      ↓
bounded worker pool
      ↓
provider Collector
      ↓
RawJob
      ↓
Normalizer
      ↓
identity / conservative deduplication
      ↓
Job + JobOccurrence
      ↓
Lifecycle processing
      ↓
Rule-based JobEvaluation
      ↓
REST API / metrics / experiment data
```

PostgreSQL is the durable source of truth. Redis, Kafka, a distributed queue, vector database, and LLM are not required to prove Phase 0.

# Implementation Roadmap v0

The roadmap defines outcomes and boundaries, not implementation code. Complete stages in order unless real evidence requires changing the specification.

## Stage 1 — Project Foundation

Goal: runnable Go service with clean startup boundaries.

Implement the Go module/project structure, configuration loading/validation, structured logging, HTTP server, `/health/live`, `/health/ready`, application startup/shutdown lifecycle, local Docker Compose with PostgreSQL, and initial CI build/test/format/lint checks.

Expected result: the service starts through the intended local workflow, validates config, connects to PostgreSQL, serves health endpoints, and shuts down cleanly.

Do not implement collectors or business workflows yet.

## Stage 2 — Database Foundation and Core Model

Goal: PostgreSQL becomes the durable market model.

Create versioned migrations and repositories for the minimum core entities needed by subsequent stages: Company, CompanyAlias where needed, SourceCandidate, JobSource, SourceSync, RawJob representation, Job, JobOccurrence, and JobEvaluation.

Add identity/integrity constraints, especially uniqueness of an occurrence by source + external identifier. Integration-test repositories and migrations against real PostgreSQL.

Expected result: an empty DB migrates to a valid schema and repeated writes cannot trivially violate core identity invariants.

## Stage 3 — First Direct Provider Collector

Goal: prove one real public ATS end-to-end before generalizing.

Choose one supported public ATS and implement its HTTP client, timeout/response-size protection, DTO parsing, raw capture, provider-neutral mapping, fixtures, and failure classification.

Expected result: given a known source, JobRadar retrieves and parses its current public vacancies reliably.

Build abstractions from this real implementation rather than designing four hypothetical collectors first.

## Stage 4 — Ingestion, Identity, and Idempotency

Goal: repeated observations become stable internal data.

Implement normalization, exact occurrence identity, Job/JobOccurrence creation/update, raw payload/update handling, first\_seen/last\_seen behavior, and idempotent reprocessing.

Expected result: repeated processing of the same source does not create duplicate jobs/occurrences, while material changes update the existing observation.

## Stage 5 — Scheduler and Reliable Synchronization

Goal: collection becomes autonomous.

Implement PostgreSQL-backed due-source selection, bounded worker pool, global/provider concurrency, provider rate limiting, HTTP timeouts, bounded retries, exponential backoff + jitter, SourceSync audit history, source failure state, and graceful cancellation.

Expected result: registered sources synchronize automatically; failures are visible and cannot corrupt lifecycle state.

## Stage 6 — Vacancy Lifecycle

Goal: safely distinguish active, missing, and closed observations.

Implement successful-complete-sync semantics, ACTIVE→MISSING behavior, grace-period closure, FAILED-sync protection, Job status derivation from occurrences/source authority, and a minimum safeguard against suspicious empty responses.

Expected result: tests prove temporary provider failure or one missing observation cannot mass-close live vacancies.

## Stage 7 — Additional Providers and Source Discovery

Goal: move from one known source to a self-growing registry.

Add remaining direct ATS adapters one by one. Then implement discovery inputs, company identity/resolution, career-page/ATS detection, SourceCandidate, verification, promotion to JobSource, and rejection/failure recording.

Expected result: the registry grows from discovery data rather than a permanently hand-maintained company list.

Do not build a generic anti-bot crawler.

## Stage 8 — Cross-Source Deduplication

Goal: represent one logical vacancy seen through several sources without aggressive false merges.

Implement normalized comparison fields, deterministic multi-signal matching, high-confidence attachment, uncertain duplicate representation/review, and source-authority/canonical-data rules.

Expected result: known duplicates converge to one Job with several occurrences while ambiguous cases remain separate.

No LLM/embeddings until real duplicate data demonstrates a need.

## Stage 9 — Relevance Evaluation

Goal: reduce market data to an explainable daily set.

Implement SearchProfile representation, role/title rules, Go/Golang/PHP detection, hard seniority exclusions, geography/remote signals, RELEVANT/REVIEW/NOT\_RELEVANT, reasons, rules versioning, and reevaluation after material changes.

Expected result: JobRadar produces an explainable shortlist without deleting non-relevant market data.

Do not invent a 0–100 score.

## Stage 10 — Query API

Goal: use and inspect JobRadar without querying PostgreSQL manually.

Implement agreed read-oriented `/api/v1` endpoints for jobs/detail, companies, sources/sync history, source candidates, and overview/Phase 0 statistics, with filters, pagination, stable errors, and provenance.

Expected result: the owner can see what was found, where it came from, why it was classified, and whether collection is healthy.

## Stage 11 — Observability and Profiling

Goal: operate the system for days without guessing.

Implement Prometheus metrics with bounded-cardinality labels, `/metrics`, one useful Grafana dashboard, controlled pprof access, correlation/sync IDs, and useful sync summary logs.

Expected result: failures, ingestion volume, latency, retries, source health, and relevance yield are observable, and real CPU/heap/goroutine problems can be profiled.

Full distributed tracing remains deferred.

## Stage 12 — Production-like Deployment

Goal: run continuously outside the development machine.

Implement the multi-stage non-root image, production Compose, persistent PostgreSQL, migration procedure, private DB/Prometheus/pprof exposure, appropriate HTTPS/private API/Grafana access, backups, build metadata, and smoke test.

Expected result: JobRadar survives restart/redeploy without losing market state and runs unattended on one VPS.

## Stage 13 — Phase 0 Experiment

Goal: decide whether the acquisition strategy justifies the next product phase.

Run continuously for approximately 7–14 days and collect the measurements defined above: unique jobs, source contribution/overlap, benchmark intersection/misses, observed coverage, discovery latency, reliability, duplicate quality, relevance yield, discovery funnel, and stale-data examples.

Fix correctness/reliability defects during the experiment, but do not hide weak acquisition by adding AI or auto-apply.

Expected result: a written Phase 0 report separately concludes whether the technical system is reliable enough and whether the acquisition/product hypothesis is supported.

# Start Decision

**v0 design is complete enough to implement.**

Further architecture work before code would mostly be speculation. Exact polling intervals, concurrency limits, lifecycle grace periods, batching strategy, dedup thresholds, and infrastructure resource limits should now be determined from experiments and observed data.

Start with **Stage 1 — Project Foundation**.

From this point the development loop is:

```text
stage requirement
    ↓
you implement independently
    ↓
we review implementation/results
    ↓
update the specification if evidence changes a decision
    ↓
next stage
```

AI job analysis, salary estimation, company research, cover-letter generation, application assistance, Gmail status tracking, and possible application automation remain outside v0 until Phase 0 proves useful market acquisition.

