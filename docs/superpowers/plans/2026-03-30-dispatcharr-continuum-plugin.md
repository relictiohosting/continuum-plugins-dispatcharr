# Dispatcharr Continuum Plugin Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Dispatcharr-specific Continuum plugin that exposes one shared **Live TV** source, supports Xtream-first connectivity with M3U/XMLTV fallback, plays live streams, and renders EPG.

**Architecture:** Build a Go plugin around the Continuum plugin SDK with a thin host-facing surface and explicit internal boundaries for config, upstream clients, mapping, cache/state, sync, and playback resolution. Implement Xtream mode first, prove the full demo path end-to-end, then add M3U/XMLTV fallback using the same canonical channel/program model and cache contract.

Assume the plugin integrates with Continuum so the resulting `Live TV` source is exposed via Continuum's Jellyfin-compatible API surface, not via a separate bespoke API.

**Tech Stack:** Go, Continuum plugin SDK, net/http, encoding/xml, testing package, optional test fixtures, JSON/protobuf structures required by the SDK.

---

## File Structure

### Planned files and responsibilities

**Bootstrap / packaging**
- Create: `go.mod` — module definition and SDK dependency
- Create: `go.sum` — dependency lockfile
- Create: `main.go` — plugin bootstrap, capability registration, manifest wiring
- Create: `manifest.json` — Continuum plugin manifest and admin config schema
- Create: `README.md` — local development and plugin usage notes

**Internal config + schema**
- Create: `internal/config/settings.go` — typed plugin settings, defaults, validation helpers
- Create: `internal/config/schema.go` — manifest/config schema generation for Xtream vs M3U/XMLTV form
- Create: `internal/config/secrets.go` — secret masking and config read/write helpers

**Canonical model**
- Create: `internal/model/source.go` — source mode enums and source identity constants
- Create: `internal/model/channel.go` — channel model and stable ID rules
- Create: `internal/model/program.go` — program model and stable ID rules
- Create: `internal/model/state.go` — cached catalog, guide, and sync-health state

**Upstream clients/parsers**
- Create: `internal/upstream/httpclient/client.go` — shared HTTP client, timeouts, request helpers
- Create: `internal/upstream/xtream/client.go` — Xtream auth/catalog/EPG/playback resolution calls
- Create: `internal/upstream/xtream/types.go` — Xtream response DTOs
- Create: `internal/upstream/m3u/parser.go` — M3U playlist parsing
- Create: `internal/upstream/xmltv/parser.go` — XMLTV parsing

**Mapping + cache**
- Create: `internal/mapping/channel_mapper.go` — upstream-to-canonical channel mapping
- Create: `internal/mapping/program_mapper.go` — upstream-to-canonical program mapping
- Create: `internal/matching/xmltv_matcher.go` — deterministic M3U↔XMLTV join rules
- Create: `internal/cache/store.go` — in-memory or SDK-backed cached state abstraction
- Create: `internal/cache/snapshot.go` — snapshot read/write helpers and stale-state semantics

**Plugin capabilities / application layer**
- Create: `internal/app/service.go` — orchestration façade used by SDK entrypoints
- Create: `internal/app/connection.go` — test-connection workflow
- Create: `internal/app/sync.go` — scheduled + on-demand sync orchestration
- Create: `internal/app/playback.go` — fresh stream resolution at play time
- Create: `internal/plugin/provider.go` — source/provider registration with Continuum SDK
- Create: `internal/plugin/routes.go` — HTTP routes for admin actions and playback handoff if needed
- Create: `internal/plugin/tasks.go` — scheduled task registration/execution
- Create: `internal/plugin/health.go` — admin-visible health/status payloads

**Tests / fixtures**
- Create: `testdata/xtream/live_categories.json`
- Create: `testdata/xtream/live_streams.json`
- Create: `testdata/xtream/short_epg.json`
- Create: `testdata/m3u/sample.m3u`
- Create: `testdata/xmltv/sample.xml`
- Create: `internal/model/channel_test.go`
- Create: `internal/model/program_test.go`
- Create: `internal/config/settings_test.go`
- Create: `internal/upstream/xtream/client_test.go`
- Create: `internal/upstream/m3u/parser_test.go`
- Create: `internal/upstream/xmltv/parser_test.go`
- Create: `internal/matching/xmltv_matcher_test.go`
- Create: `internal/mapping/channel_mapper_test.go`
- Create: `internal/mapping/program_mapper_test.go`
- Create: `internal/app/connection_test.go`
- Create: `internal/app/sync_test.go`
- Create: `internal/app/playback_test.go`
- Create: `internal/app/mode_switch_test.go`

**Plan-time notes**
- If the SDK requires different package names or entrypoint layout, adapt the file structure during Task 1 before writing feature code.
- If the SDK lacks secret/config persistence or provider/source registration primitives, document the exact adaptation before proceeding beyond Chunk 1.

---

## Chunk 1: Verify Continuum SDK Fit and Lock Skeleton

### Task 1: Verify required SDK primitives and record adaptations

**Files:**
- Modify: `docs/superpowers/specs/2026-03-30-dispatcharr-continuum-plugin-design.md`
- Create: `docs/sdk-fit-notes.md`

- [ ] **Step 1: Inspect the Continuum SDK capability surface against the approved spec**

Confirm whether the SDK supports:
- plugin configuration persistence
- masked secret fields
- scheduled tasks
- provider/source registration for one visible source named `Live TV`
- playback resolution or HTTP route handoff
- admin-visible status/error reporting

Inspect these SDK references directly:
- `README.md`
- `docs/compatibility.md`
- `examples/hello-scheduled-task/main.go`
- `examples/hello-scheduled-task/manifest.json`
- `proto/continuum/plugin/v1/auth_provider.proto`
- `proto/continuum/plugin/v1/http_routes.proto`
- `pkg/pluginsdk/runtime/runtime.go`

- [ ] **Step 2: Write the fit note with one row per required primitive**

Use this table format in `docs/sdk-fit-notes.md`:

```md
| Primitive | SDK support | Evidence | Adaptation needed? |
|-----------|-------------|----------|--------------------|
| Scheduled task | Yes/No | <doc/file> | <notes> |
```

- [ ] **Step 3: Update the spec if an assumption changes**

If any primitive is missing or different than expected, add a short “Planning confirmation” note to the spec before implementation starts.

- [ ] **Step 4: Review the fit note for blockers**

Decision rule:
- If a primitive is missing but has a viable workaround, record the workaround and continue.
- If a primitive is missing and blocks the architecture, stop and revise the plan before coding.

### Task 2: Create the plugin scaffold and manifest shell

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `manifest.json`
- Create: `README.md`

- [ ] **Step 1: Write a failing bootstrap test if the scaffold has testable helpers**

If the SDK bootstrap can be tested, add a minimal test first. Otherwise skip directly to the scaffold.

- [ ] **Step 2: Initialize the Go module and add the SDK dependency**

Run:
```bash
go mod init github.com/relictiohosting/continuum-plugins/dispatcharr
go get github.com/ContinuumApp/continuum-plugin-sdk@<verified-version-from-task-1>
```

Expected: `go.mod` and `go.sum` created without version conflicts.

- [ ] **Step 3: Create the minimal plugin entrypoint**

The initial `main.go` should:
- boot the SDK runtime
- register only the capabilities confirmed in Task 1
- compile cleanly

- [ ] **Step 4: Create the initial manifest shell**

Manifest must include:
- plugin name / id
- version
- supported capabilities
- admin configuration form stub

- [ ] **Step 5: Build the empty plugin**

Run:
```bash
go build ./...
```

Expected: successful compile with placeholder capability implementations.

### Task 3: Establish config schema and validation contract

**Files:**
- Create: `internal/config/settings.go`
- Create: `internal/config/schema.go`
- Create: `internal/config/secrets.go`
- Create: `internal/config/settings_test.go`

- [ ] **Step 1: Write failing tests for settings validation**

Cover:
- Xtream requires base URL + username + password
- M3U mode requires M3U URL + EPG XML URL
- EPG is required in v1
- source mode switching requires reset warning metadata

Example:

```go
func TestValidate_XtreamRequiresCredentials(t *testing.T) {
    cfg := Settings{SourceMode: SourceModeXtream}
    err := cfg.Validate()
    if err == nil {
        t.Fatal("expected validation error")
    }
}
```

- [ ] **Step 2: Run the config tests and verify failure**

Run:
```bash
go test ./internal/config -run TestValidate -v
```

Expected: FAIL because validation and types do not exist yet.

- [ ] **Step 3: Implement minimal typed settings and schema helpers**

Include:
- source mode enum
- Live TV toggle
- VOD status-only field
- optional refresh intervals behind defaults
- masked secret metadata

- [ ] **Step 4: Verify the generated schema/manifest shape matches the spec**

Check explicitly:
- conditional Xtream vs M3U/XMLTV fields
- masked password field
- non-interactive VOD status row
- `Live TV` naming and one-source assumptions where represented in config

- [ ] **Step 5: Run config tests again**

Run:
```bash
go test ./internal/config -v
```

Expected: PASS.

### Task 4: Review Chunk 1 plan assumptions before moving on

**Files:**
- Modify: `docs/superpowers/plans/2026-03-30-dispatcharr-continuum-plugin.md`

- [ ] **Step 1: Compare scaffold reality to the plan**
- [ ] **Step 2: Adjust later file paths/tasks if the SDK forced layout changes**
- [ ] **Step 3: Commit the scaffold milestone**

```bash
git add go.mod go.sum main.go manifest.json README.md internal/config docs/sdk-fit-notes.md docs/superpowers/specs/2026-03-30-dispatcharr-continuum-plugin-design.md docs/superpowers/plans/2026-03-30-dispatcharr-continuum-plugin.md
git commit -m "chore: scaffold dispatcharr continuum plugin"
```

---

## Chunk 2: Canonical Model, Cache Contract, and Xtream Client

### Task 5: Implement canonical models and stable ID rules

**Files:**
- Create: `internal/model/source.go`
- Create: `internal/model/channel.go`
- Create: `internal/model/program.go`
- Create: `internal/model/state.go`
- Create: `internal/model/channel_test.go`
- Create: `internal/model/program_test.go`

- [ ] **Step 1: Write failing tests for channel/program ID stability**

Cover:
- identical upstream records keep the same IDs across refreshes
- Xtream IDs pass through unchanged when available
- M3U IDs use deterministic fallback hashing
- program IDs remain stable for unchanged schedule entries

- [ ] **Step 2: Run the model tests and verify failure**

Run:
```bash
go test ./internal/model -run 'Test.*ID' -v
```

- [ ] **Step 3: Implement minimal model types and ID helpers**

- [ ] **Step 4: Run model tests again**

Run:
```bash
go test ./internal/model -v
```

Expected: PASS.

### Task 6: Implement the cache/state abstraction

**Files:**
- Create: `internal/cache/store.go`
- Create: `internal/cache/snapshot.go`
- Create: `internal/app/sync.go`
- Modify: `internal/model/state.go`
- Create: `internal/app/sync_test.go`

- [ ] **Step 1: Write failing tests for stale-state behavior**

Cover:
- last successful snapshot remains readable after refresh failure
- health state tracks last success and last failure
- stream URLs are not treated as durable cache fields

- [ ] **Step 2: Run the cache tests and verify failure**

Run:
```bash
go test ./internal/app -run TestSyncState -v
```

- [ ] **Step 3: Implement the snapshot/store abstraction**

Choose the simplest viable backing store supported by the SDK fit notes.

- [ ] **Step 4: Run the cache tests again**

Run:
```bash
go test ./internal/app -run TestSyncState -v
```

Expected: PASS.

### Task 7: Build the Xtream client and fixtures

**Files:**
- Create: `internal/upstream/httpclient/client.go`
- Create: `internal/upstream/xtream/client.go`
- Create: `internal/upstream/xtream/types.go`
- Create: `internal/upstream/xtream/client_test.go`
- Create: `testdata/xtream/live_categories.json`
- Create: `testdata/xtream/live_streams.json`
- Create: `testdata/xtream/short_epg.json`

- [ ] **Step 1: Write failing Xtream client tests against fixtures or httptest servers**

Cover:
- successful connection test
- fetching live categories/channels
- fetching EPG availability/data
- resolving a fresh playback target

- [ ] **Step 2: Run Xtream client tests and verify failure**

Run:
```bash
go test ./internal/upstream/xtream -v
```

- [ ] **Step 3: Implement the shared HTTP client with timeouts**

- [ ] **Step 4: Implement Xtream request methods and DTO parsing**

- [ ] **Step 5: Re-run Xtream client tests**

Run:
```bash
go test ./internal/upstream/xtream -v
```

Expected: PASS.

### Task 8: Map Xtream payloads into canonical channels/programs

**Files:**
- Create: `internal/mapping/channel_mapper.go`
- Create: `internal/mapping/program_mapper.go`
- Modify: `internal/mapping/channel_mapper_test.go`
- Modify: `internal/mapping/program_mapper_test.go`

- [ ] **Step 1: Write failing mapper tests using Xtream fixtures**

Cover:
- source-of-truth name/logo/number preservation
- canonical ID preservation
- EPG program mapping into the stable program model

- [ ] **Step 2: Run mapper tests and verify failure**

Run:
```bash
go test ./internal/mapping -run 'TestMapXtream' -v
```

- [ ] **Step 3: Implement minimal Xtream mapping logic**

- [ ] **Step 4: Re-run mapper tests**

Run:
```bash
go test ./internal/mapping -v
```

Expected: PASS.

### Task 9: Commit the Xtream data layer milestone

**Files:**
- Modify: relevant files from Tasks 5-8

- [ ] **Step 1: Run focused tests for models, cache, and Xtream layers**

Run:
```bash
go test ./internal/model ./internal/cache ./internal/mapping ./internal/upstream/xtream ./internal/app -v
```

- [ ] **Step 2: Commit the milestone**

```bash
git add internal/model internal/cache internal/mapping internal/upstream testdata/xtream docs/superpowers/plans/2026-03-30-dispatcharr-continuum-plugin.md
git commit -m "feat: add xtream data model and cache contract"
```

---

## Chunk 3: Plugin Workflows for Connection Test, Sync, Playback, and Live TV Source

### Task 10: Implement the app service and connection test workflow

**Files:**
- Create: `internal/app/service.go`
- Create: `internal/app/connection.go`
- Create: `internal/app/connection_test.go`

- [ ] **Step 1: Write failing tests for the admin test-connection path**

Cover:
- Xtream mode success requires credentials + channels + EPG
- validation errors surface meaningful admin status
- source-mode-specific branching is wired cleanly for future M3U/XMLTV support

- [ ] **Step 2: Run connection tests and verify failure**

Run:
```bash
go test ./internal/app -run TestConnection -v
```

- [ ] **Step 3: Implement the minimal service façade and connection workflow**

- [ ] **Step 4: Re-run connection tests**

Run:
```bash
go test ./internal/app -run TestConnection -v
```

Expected: PASS.

### Task 11: Implement scheduled sync and on-demand refresh orchestration

**Files:**
- Modify: `internal/app/sync.go`
- Modify: `internal/app/sync_test.go`
- Create: `internal/plugin/tasks.go`

- [ ] **Step 1: Write failing tests for sync orchestration**

Cover:
- daily channel refresh defaults
- EPG refresh every few hours
- on-demand refresh fallback
- stale snapshot survives upstream outage

- [ ] **Step 2: Run sync tests and verify failure**

Run:
```bash
go test ./internal/app -run TestSync -v
```

- [ ] **Step 3: Implement sync scheduling and orchestration**

- [ ] **Step 4: Re-run sync tests**

Run:
```bash
go test ./internal/app -run TestSync -v
```

Expected: PASS.

### Task 12: Implement fresh playback resolution

**Files:**
- Create: `internal/app/playback.go`
- Create: `internal/app/playback_test.go`
- Modify: `internal/upstream/xtream/client.go`

- [ ] **Step 1: Write failing tests for play-time resolution**

Cover:
- playback does not depend on stale cached stream URLs
- current upstream target is resolved at play time
- Dispatcharr failover remains upstream-owned

- [ ] **Step 2: Run playback tests and verify failure**

Run:
```bash
go test ./internal/app -run TestPlayback -v
```

- [ ] **Step 3: Implement minimal play-time resolution logic**

- [ ] **Step 4: Re-run playback tests**

Run:
```bash
go test ./internal/app -run TestPlayback -v
```

Expected: PASS.

### Task 13: Register the `Live TV` source and required plugin routes

**Files:**
- Create: `internal/plugin/provider.go`
- Create: `internal/plugin/routes.go`
- Create: `internal/plugin/health.go`
- Modify: `main.go`
- Modify: `manifest.json`

- [ ] **Step 1: Write failing integration-style tests around host-facing registration helpers if practical**

If the SDK is not test-friendly here, create narrow unit tests around the data these handlers emit.

- [ ] **Step 2: Implement the provider/source registration**

Requirements:
- expose one user-visible source named `Live TV`
- expose searchable channels and guide data
- expose admin-visible health state

- [ ] **Step 3: Implement admin action routes/status endpoints if the SDK uses routes for this**

- [ ] **Step 4: Build the plugin and run the relevant tests**

Run:
```bash
go test ./internal/app ./internal/plugin -v
go build ./...
```

Expected: PASS + successful build.

### Task 14: Commit the Xtream-first vertical slice

**Files:**
- Modify: relevant files from Tasks 10-13

- [ ] **Step 1: Run the full automated suite available so far**

Run:
```bash
go test ./... -v
go build ./...
```

- [ ] **Step 2: Commit the milestone**

```bash
git add .
git commit -m "feat: add xtream-backed live tv source"
```

---

## Chunk 4: M3U/XMLTV Fallback Mode

### Task 15: Implement the M3U parser

**Files:**
- Create: `internal/upstream/m3u/parser.go`
- Create: `internal/upstream/m3u/parser_test.go`
- Create: `testdata/m3u/sample.m3u`

- [ ] **Step 1: Write failing tests for playlist parsing**

Cover:
- title extraction
- tvg-id extraction
- logo extraction
- stream URL extraction
- deterministic fallback ID inputs

- [ ] **Step 2: Run parser tests and verify failure**

Run:
```bash
go test ./internal/upstream/m3u -v
```

- [ ] **Step 3: Implement the M3U parser**

- [ ] **Step 4: Re-run parser tests**

Run:
```bash
go test ./internal/upstream/m3u -v
```

Expected: PASS.

### Task 16: Implement the XMLTV parser and matcher

**Files:**
- Create: `internal/upstream/xmltv/parser.go`
- Create: `internal/upstream/xmltv/parser_test.go`
- Create: `internal/matching/xmltv_matcher.go`
- Create: `internal/matching/xmltv_matcher_test.go`
- Create: `testdata/xmltv/sample.xml`

- [ ] **Step 1: Write failing tests for XMLTV parsing and matching**

Cover:
- channel parsing
- programme parsing
- join order: XMLTV id, tvg-id, normalized channel name
- unmatched channels still show without guide data

- [ ] **Step 2: Run parser/matcher tests and verify failure**

Run:
```bash
go test ./internal/upstream/xmltv ./internal/matching -v
```

- [ ] **Step 3: Implement parser and deterministic matcher**

- [ ] **Step 4: Re-run parser/matcher tests**

Run:
```bash
go test ./internal/upstream/xmltv ./internal/matching -v
```

Expected: PASS.

### Task 17: Add fallback-mode mapping and connection/sync wiring

**Files:**
- Modify: `internal/mapping/channel_mapper.go`
- Modify: `internal/mapping/program_mapper.go`
- Modify: `internal/app/connection.go`
- Modify: `internal/app/sync.go`
- Modify: `internal/app/playback.go`
- Modify: `internal/app/connection_test.go`
- Modify: `internal/app/sync_test.go`
- Modify: `internal/app/playback_test.go`

- [ ] **Step 1: Write failing tests for M3U/XMLTV end-to-end fallback behavior**

Cover:
- setup validation succeeds with playlist + EPG
- channels are mapped into canonical `Live TV` items
- guide attaches only when deterministic rules match
- playback resolves from current playlist data

- [ ] **Step 2: Run fallback tests and verify failure**

Run:
```bash
go test ./internal/app ./internal/mapping -run 'Test.*M3U|Test.*XMLTV' -v
```

- [ ] **Step 3: Implement the fallback-mode wiring**

- [ ] **Step 4: Re-run fallback tests**

Run:
```bash
go test ./internal/app ./internal/mapping -v
```

Expected: PASS.

### Task 18: Commit the fallback milestone

**Files:**
- Modify: relevant files from Tasks 15-17

- [ ] **Step 1: Run the full test suite and build**

Run:
```bash
go test ./... -v
go build ./...
```

- [ ] **Step 2: Commit the milestone**

```bash
git add .
git commit -m "feat: add m3u xmltv fallback mode"
```

### Task 19: Implement source-mode switching reset and rebuild behavior

**Files:**
- Modify: `internal/app/connection.go`
- Modify: `internal/app/sync.go`
- Modify: `internal/cache/store.go`
- Modify: `internal/cache/snapshot.go`
- Create: `internal/app/mode_switch_test.go`

- [ ] **Step 1: Write failing tests for source-mode switching**

Cover:
- changing source mode emits the required reset warning metadata
- cached channel/program state is cleared and rebuilt from the new source mode
- the top-level `Live TV` source identity remains stable across the switch

- [ ] **Step 2: Run mode-switch tests and verify failure**

Run:
```bash
go test ./internal/app -run TestSourceModeSwitch -v
```

- [ ] **Step 3: Implement the reset and rebuild behavior**

- [ ] **Step 4: Re-run mode-switch tests**

Run:
```bash
go test ./internal/app -run TestSourceModeSwitch -v
```

Expected: PASS.

---

## Chunk 5: Demo Readiness and Manual Verification

### Task 20: Add README usage notes and known limitations

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Document setup for Xtream mode**
- [ ] **Step 2: Document setup for M3U/XMLTV mode**
- [ ] **Step 3: Document v1 limitations**

Must include:
- one source only
- no VOD playback
- no filtering controls
- EPG required for setup

### Task 21: Run the demo checklist manually

**Files:**
- Create: `docs/demo-checklist.md`

- [ ] **Step 1: Write the manual verification checklist**

Checklist must include:
- configure Xtream mode
- test connection succeeds
- `Live TV` appears in Continuum
- channels search works
- stream playback works
- guide data renders
- simulate upstream outage and confirm stale metadata remains visible
- switch to M3U/XMLTV and confirm reset warning + successful rebuild

- [ ] **Step 2: Execute the checklist and capture results**

Run the app/plugin manually in the target environment and record pass/fail notes next to each item.

### Task 22: Final verification and release-ready checkpoint

**Files:**
- Modify: `docs/demo-checklist.md`
- Modify: `README.md`

- [ ] **Step 1: Run final automated verification**

Run:
```bash
go test ./... -v
go build ./...
```

Expected: PASS.

- [ ] **Step 2: Confirm the acceptance targets from the design spec**

Match against:
- Xtream mode connects successfully
- channels appear under `Live TV`
- streams play
- guide renders
- M3U/XMLTV fallback works

- [ ] **Step 3: Commit the release-ready state**

```bash
git add README.md docs/demo-checklist.md docs/superpowers/plans/2026-03-30-dispatcharr-continuum-plugin.md
git commit -m "docs: finalize dispatcharr plugin implementation notes"
```

---

## Planning Notes for the Implementer

- Do not start with M3U/XMLTV. The Xtream vertical slice is the fastest path to a meaningful demo.
- Keep playback resolution separate from catalog caching.
- Do not add category filtering, multiple sources, or VOD playback during v1.
- If the SDK does not offer a first-class provider/source abstraction, adapt through the nearest supported capability and update the plan before proceeding.
- Prefer table-driven tests and fixtures over ad hoc parsing logic.
