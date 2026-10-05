# AGENTS.md

Guidance for AI agents working in this repository.

## Overview

Internal dashboard for thecluster.lan. Vue 3 frontend + Go API backend, packaged with Nix Flakes and deployed via Helm to Kubernetes.

## Development Environment

```bash
nix develop     # Enter shell with all tools (bun, go, helm, kubectl, etc.)
# or: direnv auto-loads via .envrc (use flake)
```

## Common Commands

```bash
make build          # Build web + API
make test           # Run Go tests (go test ./...)
make run            # Start API at localhost:8080
make dev            # Vite on 0.0.0.0:5173 (HMR, proxies /ping and /api) + API under air on :8080
make check          # Full check: test + build-web + chart-lint + nix flake check
make lint           # Helm chart lint
make clean          # Remove dist/, api/thecluster-api, api/tmp/, result
make gen            # tdl fmt + tdl gen: model/cluster.tdl -> api/internal/model
```

**Run a single Go test:**
```bash
cd api && go test ./internal/server/ -run TestPing
```

**Frontend dev (hot reload):**
```bash
cd web && bun run dev
```

**Nix builds:**
```bash
nix build .#api     # Go binary
nix build .#web     # Vue static assets
nix build .#app     # Combined binary + assets
nix build .#ctr     # Docker image (stream layered)
```

## Architecture

### Model (`model/`)
- `model/cluster.tdl` defines the `/api/nodes` response types in [tdl](https://github.com/UnstoppableMango/tdl)
- Its `target go` block generates `api/internal/model/` (do not edit; run `make gen`). JSON tags and Go names come from `tag`/`name` directives in that block
- A field needing two directives uses a nested block (`cpuPct { name(...) tag(...) }`); `field => a(...) b(...)` attaches only the first
- `out(...)` resolves against the working directory, so run `tdl gen` from the repository root
- `nix flake check` runs `tdl-check`, `tdl-fmt`, and `tdl-gen` (`tdl gen --verify`, fails when generated code is stale)

### Go API (`api/`)
- Entry: `cmd/thecluster-api/main.go`
- Logic: `internal/server/server.go` — `GET /ping`, `GET /api/nodes`, and static file serving via `chi` + `github.com/olivere/vite`
- `internal/metrics/`: minimal Prometheus HTTP client and node snapshot builder, returning `internal/model` types (kube-state-metrics for node list/readiness/cordon/role, node-exporter for CPU/mem/disk/net/root FS). Health thresholds live in `health.go`
- Prometheus base URL from `PROMETHEUS_URL` (default `http://kube-prometheus-stack-prometheus.monitoring:9090`)
- Serves Vue build output from `../web/dist` (configurable via `STATIC_DIR` env var)
- Routing: `GET /` → Vite index, `GET /assets/*` → Vite assets, all other paths → `404.html` with HTTP 404
- Path traversal: handled by `path.Clean` (vite handler) + `os.DirFS` + `http.FileServerFS` (stdlib)
- Dependencies: `github.com/go-chi/chi/v5`, `github.com/olivere/vite`

### Vue Frontend (`web/`)
- Vue 3 Composition API (`<script setup>`), Vite, Tailwind CSS v4
- `App.vue`: rack-monitor node health board (header + 4-column tile grid), polls `/api/nodes` every 15s
- `components/`: `ClusterHeader.vue`, `NodeTile.vue`, `Sparkline.vue` (inline SVG, no chart library)
- `composables/useNodes.ts`: polling + stale detection (overlay after 60s without a successful fetch)
- Sizing is rem-based with `1rem = min(100vw/120, 100vh/67.5)`, so the board fills 1080p and 4K alike
- No 404 page — Go embeds `api/internal/server/404.html` at compile time and serves it directly
- Build output → `dist/` (consumed by Go API for static serving)

### Nix Packaging (`nix/`)
- `api.nix` — `buildGoApplication` (uses `gomod2nix.toml`)
- `web.nix` — `bun2nix.mkDerivation` (uses `bun.nix`, auto-generated from `bun.lock`)
- `app.nix` — combines Go binary + web dist into single derivation
- `ctr.nix` — `dockerTools.streamLayeredImage`

### Helm Chart (`charts/thecluster/`)
- Deploys the combined app container to Kubernetes
- CI publishes chart updates to `ghcr.io` on push to main

## Dependency Lock File Regeneration

When Go deps change:
```bash
cd api && gomod2nix
```

When Bun deps change (postinstall hook auto-runs `bun2nix`):
```bash
cd web && bun install
```

Or regenerate both:
```bash
make nix-deps
```

## CI Pipeline

Five parallel jobs: `web` (bun build), `api` (go test), `nix` (flake check + nix build), `helm` (lint + publish), `docker` (depends on nix — loads and pushes container).

Images published to `ghcr.io` on push to main with git SHA tag.
