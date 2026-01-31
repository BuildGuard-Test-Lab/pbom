# CLAUDE.md - Project Memory

> **AI Instruction:** You are responsible for maintaining this file. Whenever we change the tech stack, add new tools, or complete major milestones, update this file automatically.

## Project Context
- **Project:** PBOM — Pipeline Bill of Materials. Tracks *how* artifacts reach production (not *what's inside* them — that's SBOM).
- **Tech Stack:** Go 1.23, Cobra (CLI), ORAS (OCI artifact storage — pending), GitHub Actions (org-level webhooks + required workflows), Kargo (promotion tracking — deferred)
- **Build Commands:** `go build ./...` | Binary: `go build -o pbom ./cmd/pbom`
- **Target Environment:** GitHub Actions + Kargo. Must scale to 1,000+ repos with zero developer friction.

## Architecture Decisions
- **Zero-Touch:** No opt-in. Developers never modify their workflows.
  - **Org Webhooks:** Listen to `workflow_run.completed` events to scrape metadata.
  - **Required Workflows:** Org-level GHA workflows run alongside dev workflows automatically.
  - **Kargo Controller:** K8s watcher on Freight/Promotion CRDs (deferred to later phase).
- **Tool-Agnostic:** No dependency on Docker Buildx or any specific build tool. Uses ORAS "Referrer" artifacts to attach PBOM metadata to any artifact type.
- **Interim Storage:** OCI Registry (via ORAS referrer artifacts). No dedicated database yet — registry acts as the source of truth until CAIOS or a DB is provisioned.
- **Async Collection:** Fire-and-forget from GHA runners. Heavy processing (CVE cross-ref, document generation) happens in the backend, not on runners.

## Collection Responsibility Split

The Required Workflow runs on its own runner (not the developer's), so it can only capture universal env var data. Runner-specific and API-dependent data is collected by the webhook listener.

| Data | Collected by | Why |
|---|---|---|
| Commit SHA, branch, actor, ref | Required Workflow | Same across all runners — GITHUB_* env vars are reliable |
| Run ID, workflow name, trigger | Required Workflow | Same across all runners |
| Runner OS/arch, tool versions | Webhook Listener | Must query the *developer's* actual workflow run via GitHub API |
| Secrets accessed | Webhook Listener | Requires Audit Log API (org-level access) |
| Artifact digests, vulns | Webhook Listener | API calls after the build completes |
| Kargo freight/promotion | Kargo Controller | K8s CRD watcher (deferred) |

## PBOM Data Model
A PBOM links: `{GHA run_id, Commit SHA, Image/Artifact Digest, Kargo Freight ID}`

### Phase A — Source & Build (GitHub Actions)
- Commit SHA, Workflow Run ID, Build environment (runner OS, tool versions), Secret fingerprints (accessed, not values)

### Phase B — Artifact & Security
- Artifact digest (SHA256), SLSA provenance attestations, Vulnerability snapshot (Critical/High CVE counts at build time)

### Phase C — Promotion (Kargo) — Deferred
- Freight ID, Stage history (timestamps + actors), Environment diff (co-deployed service versions)

## Guidelines & Rules
- **Full File Generation:** ALWAYS generate the entire code file for any change. No snippets.
- **Conciseness:** Keep prose and explanations short and direct.
- **Confirmation:** Discuss and wait for a "ready" signal before starting a new task.
- **Session End:** When I am "done for the day," generate a recap and update the "Current Status" below.

## Project Structure
```
cmd/pbom/main.go              ← CLI entrypoint
internal/cli/
  root.go                      ← cobra root + subcommand wiring
  generate.go                  ← reads GITHUB_* env vars, emits PBOM JSON
  validate.go                  ← checks required fields, SHA/digest formats
  inspect.go                   ← human-readable lineage summary
  filter.go                    ← evaluates repo filter rules (GitHub custom properties)
  push.go                      ← stub for ORAS registry push (not yet implemented)
  version.go                   ← prints CLI + schema version
internal/detect/
  tools.go                     ← auto-detects 14 build tools on PATH (for webhook listener use)
internal/filter/
  config.go                    ← YAML config types + LoadConfig parser
  eval.go                      ← rule evaluation engine (first-match-wins)
  filter_test.go               ← full test coverage for filter package
pkg/schema/
  pbom.go                      ← Go struct types for the PBOM data model
schema/
  pbom.schema.json             ← formal JSON Schema (Draft 2020-12)
  example.pbom.json            ← realistic sample PBOM document
workflows/
  pbom-collector.yml           ← Required Workflow template (with repo filtering)
examples/
  pbom-config.yml              ← reference filtering config for org .github repo
```

## Current Status & Memories
- **Last Milestone:** Repo filtering via GitHub custom properties. CLI `filter` command + `internal/filter` package with full test coverage. Required Workflow updated with filter-before-generate flow.
- **Session Recap (Jan 28):**
  1. Defined requirements from scratch — zero-touch, tool-agnostic, 1000+ repo scale.
  2. Chose custom JSON schema over CycloneDX.
  3. Built schema (JSON Schema + Go types + example doc).
  4. Scaffolded Go CLI with cobra (generate, validate, inspect, push stub, version).
  5. Built Required Workflow YAML for org-level deployment.
  6. Discovered that Required Workflow runs on its own runner, not the dev's — split collection responsibility between the workflow (env vars) and webhook listener (API-dependent data).
  7. `internal/detect/tools.go` exists but is unused by the CLI now — it will be used by the webhook listener to enrich PBOMs with tool versions from the dev's actual runner.
- **Session Recap (Jan 30):**
  1. Added repo filtering using GitHub custom properties as the selection mechanism.
  2. Created `internal/filter/` package — YAML config parsing (`config.go`) and first-match-wins rule evaluation engine (`eval.go`).
  3. Added `pbom filter` CLI command — takes `--config` (YAML path) and `--properties` (JSON string), exits 0 (include) or 1 (exclude).
  4. Updated Required Workflow to fetch `pbom-config.yml` from org `.github` repo, fetch repo custom properties via `gh api`, and gate all downstream steps on filter result.
  5. Safe default: if no config file exists, nothing runs on any repo.
  6. Added `gopkg.in/yaml.v3` dependency.
  7. Created `examples/pbom-config.yml` as a reference config.
- **Next Session — Start Here:**
  1. **Build the GitHub Org Webhook listener** — the enrichment service that listens to `workflow_run.completed`, queries the GitHub API for runner/tool/secret data, and enriches the PBOM. This is the highest-priority next step.
  2. Implement the ORAS-based `push` command.
  3. Build the Registry Indexer (crawl registry for PBOM referrer artifacts).
  4. (Later) Kargo Controller watcher for promotion events.
- **Key Context for Next Session:**
  - Developers do NOT have local access to runners. Everything runs on self-hosted or GitHub-hosted runners.
  - The CLI's `generate` command is headless — invoked only by the Required Workflow, never by a human.
  - The CLI's `filter` command is also headless — invoked only by the Required Workflow, never by a human.
  - The CLI's `inspect` and `validate` commands are for the platform/SRE team.
  - `inspect` should eventually query the registry directly (`pbom inspect ghcr.io/org/app@sha256:...`) instead of reading local files.
  - Repo filtering uses GitHub custom properties — set these on repos via org settings or API, then reference them in `pbom-config.yml`.
