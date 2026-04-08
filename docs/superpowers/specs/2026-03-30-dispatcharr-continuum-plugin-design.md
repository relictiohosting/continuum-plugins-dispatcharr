# Dispatcharr Continuum Plugin Design

Date: 2026-03-30
Status: Partially implemented; blocked on Continuum host capability gap

## Summary

Build a Dispatcharr-specific Continuum plugin that exposes a single shared source named **Live TV**. The plugin is configured once by an admin against one Dispatcharr server and lets Continuum users browse and play IPTV streams with guide data. The primary integration path is Dispatcharr's Xtream-compatible interface, with M3U + XMLTV as a fallback mode.

Original implementation assumption: the plugin would feed **Live TV** into Continuum's media surface so the resulting source is available through Continuum's Jellyfin-compatible API layer.

Implementation update: after reviewing the actual Continuum host code, the current plugin architecture appears to support plugin-driven metadata enrichment and plugin HTTP routes, but not creation of a brand-new Jellyfin-visible Live TV catalog/source. That means the design goal remains valid as a product target, but achieving it likely requires Continuum host changes that are explicitly out of scope for the current work.

## Goals

- Let an admin connect one Dispatcharr server to Continuum.
- Expose a single end-user source named **Live TV**.
- Support live playback and EPG in v1.
- Prefer Xtream configuration, but support M3U + XMLTV fallback.
- Preserve Dispatcharr as the source of truth for naming, logos, ordering, numbering, and stream failover behavior.
- Keep stale cached data available when Dispatcharr is temporarily unavailable.

## Non-Goals

- Per-user login or account binding inside Continuum.
- VOD playback in v1.
- Category or channel filtering in v1.
- Multiple Dispatcharr server instances in v1.
- Aggressive metadata normalization or remapping.

## Product Shape

### Plugin identity

- Product type: Dispatcharr-specific media/provider plugin
- End-user label: **Live TV**
- Instance model: exactly one configured Dispatcharr-backed source in v1

### Supported source modes

1. **Xtream (default)**
   - Admin enters base URL, username, and password.
   - Plugin uses the Dispatcharr Xtream-compatible interface for live channel and EPG access.

2. **M3U + XMLTV (fallback)**
   - Admin enters M3U URL and EPG XML URL.
   - Used when Xtream is unavailable or undesirable.

The plugin treats Xtream as the recommended path and M3U/XMLTV as compatibility mode.

## Admin Experience

### Configuration form

Use a single form with conditional fields.

Shared fields:
- Source type selector
- Test Connection action
- Live TV enabled toggle
- VOD status row labeled **Coming soon** (non-interactive)

Xtream fields:
- Base URL
- Username
- Password (masked)

M3U/XMLTV fields:
- M3U URL
- EPG XML URL

Optional v1.1-if-cheap fields:
- Channel refresh interval
- EPG refresh interval

### Test Connection requirements

The setup validation should confirm:
- credentials or URLs are valid
- channels can be fetched
- EPG is available

The setup flow does **not** need to probe real playback streams.

For v1, EPG availability is required for a valid setup in both source modes.

### Source mode switching

If an admin changes the source type after setup:
- show a warning that cached/imported data will be reset
- replace the current cached data with a rebuild from the new source mode
- keep the same single **Live TV** source identity if Continuum allows it

## End-User Experience

Users see one source named **Live TV**.

v1 user capabilities:
- browse channels
- search channels
- play live streams
- view guide/EPG data

v1 exclusions:
- no VOD items presented to users
- no category filtering controls
- no per-user authentication step

## Source of Truth Rules

Dispatcharr remains authoritative for:
- channel names
- channel ordering
- channel numbers
- logos/artwork
- stream URLs and failover behavior

The plugin should mirror Dispatcharr as-is rather than normalize or enhance the presentation.

## Canonical Data Model and Stable IDs

The plugin needs stable identities even when catalog data refreshes.

- **Plugin source ID**: one stable internal source ID representing the single **Live TV** source.
- **Channel ID**: derived from the upstream provider identity when available.
  - Xtream mode: use the upstream stream/channel identifier from the Xtream-compatible API as the canonical channel ID.
  - M3U mode: use a deterministic ID derived from the playlist entry, preferring a stable provider-specific identifier when present, otherwise a hash of normalized source attributes.
- **Program ID**: derived from the EPG item identity if present, otherwise from a deterministic combination of channel ID + program start time + normalized title.

Refreshes must preserve IDs for unchanged channels/programs so search, caching, and guide associations remain stable.

If the admin changes source mode, the plugin keeps the same top-level **Live TV** source identity but rebuilds channel/program identities from the new upstream source.

## M3U and XMLTV Matching Rules

For M3U/XMLTV fallback mode, the plugin must associate playlist channels with guide data explicitly.

Primary join order:
1. Exact XMLTV channel ID match when the playlist entry exposes one.
2. Exact match on normalized tvg-id / guide identifier fields.
3. Exact match on normalized channel name.

Fallback behavior:
- If no guide match is found, the channel still appears in **Live TV** without guide data.
- If guide entries do not match any channel, ignore them.
- The plugin must not invent fuzzy matches beyond the deterministic rules above in v1.

## Data Flow

### Xtream mode

1. Admin saves Dispatcharr base URL, username, and password.
2. Plugin validates connectivity and retrieves channel and EPG availability.
3. Plugin stores masked credentials in plugin configuration.
4. Background sync refreshes channel catalog and guide data on schedule.
5. Continuum reads cached channel and guide data from the plugin.
6. Playback requests resolve stream information fresh at play time rather than relying on stale cached stream URLs.
7. The plugin uses Dispatcharr-provided stream information without implementing separate failover logic in the plugin.

### M3U/XMLTV mode

1. Admin saves M3U URL and EPG XML URL.
2. Plugin validates playlist readability and EPG availability.
3. Background sync refreshes cached channel and guide data.
4. Continuum reads the cached data and presents the same **Live TV** source.

## Cache and Playback Contract

The plugin caches:
- channel catalog data
- guide/EPG data
- admin-visible health and last-sync state

The plugin does **not** treat stream URLs as durable cached data for playback.

Playback resolution rule:
- when a user starts playback, the plugin should resolve the current upstream stream target at request time
- cached catalog data may point to playable items, but final stream resolution should be fresh

This avoids relying on stale or credential-bound stream URLs while still allowing stale metadata and guide data to remain visible during temporary outages.

## Sync and Resilience

Use both scheduled refresh and on-demand retrieval.

Default refresh policy:
- channels: daily
- EPG: every few hours

Resilience behavior:
- if Dispatcharr is unavailable, continue serving stale cached channels and guide data
- surface an admin-visible error state indicating the last refresh failure
- recover automatically on the next successful sync

## VOD Strategy

VOD is not shipped to end users in v1.

v1 admin behavior:
- show a VOD status area only
- label it clearly as **Coming soon**

This keeps the design extensible without committing to a metadata or playback model before the live TV path is stable.

## Architecture Notes from Reference Research

- Continuum's early SDK appears to support plugin capabilities such as HTTP routes and scheduled tasks, which suits sync-based provider integration.
- Dispatcharr explicitly supports Xtream Codes API compatibility as well as M3U/XMLTV output, which aligns with the two-mode configuration model.
- Jellyfin Xtream-style integrations suggest Xtream is a practical default for catalog + guide retrieval, but Continuum should treat this plugin as a shared media source, not a user login provider.

## Continuum Capability Assumptions to Confirm During Planning

The implementation plan must verify these Continuum-side primitives before coding begins:

- plugin configuration persistence, including masked secret storage for credentials
- scheduled task execution for background refresh
- a provider/source registration model that can expose one user-visible source named **Live TV**
- a playback/request path that allows fresh stream resolution at play time
- a way to surface admin-visible connection and sync health errors

If any of these primitives are missing in the SDK, the plan must define the closest viable adaptation before implementation starts.

## Current blocker

- Continuum auto-registers `metadata_provider.v1` plugin capabilities into its metadata provider system.
- That provider path is used for metadata search and enrichment on existing movie/series library chains.
- The observed host code does not show a plugin capability that creates a new Jellyfin-facing Live TV source, channel catalog, or guide model.
- Result: this repository can align with Continuum's real SDK and config/runtime contracts, but it cannot complete the original Live TV integration goal without host-side Continuum work.

## Error Handling

Admin-facing errors should cover:
- invalid credentials
- unreachable base URL or source URLs
- channels unavailable
- EPG unavailable
- sync failure after prior success

User-facing behavior should stay simple:
- keep the **Live TV** source visible when cached data exists
- avoid exposing raw provider configuration details to end users

## Testing Expectations

Acceptance targets for the first working demo:

1. Xtream mode connects successfully.
2. Channels appear in Continuum under **Live TV**.
3. Streams play.
4. Guide data renders.
5. M3U/XMLTV fallback also works.

## Future Expansion

After v1, likely next additions are:
- category filtering/group controls
- multiple Dispatcharr sources
- user-visible VOD support
- admin-configurable sync intervals if omitted from v1

## Decisions Made

- Primary job: display IPTV streams in Continuum.
- Auth model: admin-configured shared server, not per-user login.
- Plugin count: one plugin.
- Source modes: Xtream default, M3U/XMLTV fallback.
- End-user source name: **Live TV**.
- EPG: required for a valid v1 setup in both source modes.
- Filtering: no filtering in v1.
- Failover: Dispatcharr handles it.
- VOD: admin-only placeholder in v1.
- Refresh: scheduled sync plus on-demand fallback.
- Resilience: stale cache stays visible on failures.
- Presentation: mirror Dispatcharr exactly.
