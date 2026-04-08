# Dispatcharr Continuum Plugin

Dispatcharr-specific Continuum plugin for exposing a shared **Live TV** source that Continuum can surface through its Jellyfin-compatible API layer.

## Supported source modes

- **Xtream** (default/recommended)
  - Base URL
  - Username
  - Password
- **M3U/XMLTV** fallback
  - M3U URL
  - EPG XML URL

## Current behavior

- Validates admin configuration for Xtream and M3U/XMLTV modes
- Syncs channel and guide data into one stable `Live TV` source
- Resolves playback targets fresh at play time
- Keeps stale metadata visible when sync fails
- Exposes a plugin status route at `/dispatcharr/status`
- Exposes plugin bridge routes:
  - `/dispatcharr/player` (navigable page)
  - `/dispatcharr/channels`
  - `/dispatcharr/guide`
  - `/dispatcharr/stream?channel_id=...`
- Supports a scheduled sync task with key `dispatcharr-sync`

## v1 limitations

- Exactly one Dispatcharr-backed source
- No VOD playback
- No category/channel filtering controls
- EPG is required for setup in both source modes
- Source-mode changes reset cached channel/guide state before rebuilding
- Continuum host integration still needs real environment validation
- Current Continuum plugin host wiring appears to support plugin-driven **metadata enrichment**, not creation of a brand-new Jellyfin-visible **Live TV** catalog/source. A true Dispatcharr Live TV integration likely requires host-side capability work in Continuum first.

See also:
- `docs/continuum-host-gap.md`
- `docs/continuum-host-change-proposal.md`
- `docs/sdk-fit-notes.md`
- `docs/demo-checklist.md`

## Build

```bash
go build ./...
```

## GitLab CI builds

The repository includes `.gitlab-ci.yml` to run tests and produce versioned plugin binaries.

- Tagged builds (`vX.Y.Z`) use `X.Y.Z` as the plugin manifest version.
- Branch builds use a snapshot version `0.0.0-<shortsha>`.
- Artifacts include:
  - Linux binaries (`amd64`, `arm64`)
  - generated manifest JSON from each binary (`<binary>.manifest.json`)
  - SHA256 files (`<binary>.sha256`)

## GitHub Actions builds and releases

The repository also includes `.github/workflows/ci.yml` for GitHub-hosted runners.

- Runs tests on every pull request and push.
- Builds Linux binaries for `amd64` and `arm64`.
- Publishes a GitHub Release on every push:
  - `main` branch pushes publish prerelease snapshots (`snapshot-<sha>` tags).
  - `v*` tags publish normal releases.

## Test

```bash
go test ./... -v
```

## Inspect manifest

```bash
go run . manifest
```

## License

`continuum-plugin-dispatcharr` is licensed under `AGPL-3.0-or-later`. See `LICENSE`.
