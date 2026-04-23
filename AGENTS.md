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
make check          # Full check: test + build-web + chart-lint + nix flake check
make lint           # Helm chart lint
make clean          # Remove dist/, api/thecluster-api, result
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

### Go API (`api/`)
- Entry: `cmd/thecluster-api/main.go`
- Logic: `internal/server/server.go` — `GET /ping` endpoint + static file serving via `chi` + `github.com/olivere/vite`
- Serves Vue build output from `../web/dist` (configurable via `STATIC_DIR` env var)
- Routing: `GET /` → Vite index, `GET /assets/*` → Vite assets, all other paths → `404.html` with HTTP 404
- Path traversal: handled by `path.Clean` (vite handler) + `os.DirFS` + `http.FileServerFS` (stdlib)
- Dependencies: `github.com/go-chi/chi/v5`, `github.com/olivere/vite`

### Vue Frontend (`web/`)
- Vue 3 Composition API (`<script setup>`), Vite, Tailwind CSS v4
- `App.vue` — main page with `/ping` call + response display
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
