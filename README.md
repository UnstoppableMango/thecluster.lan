# THECLUSTER

Internal dashboard for `thecluster.lan`.

## Stack

- Vue + Vite + Tailwind CSS for the web UI
- Go for the API and static asset serving
- Nix for dependency management and builds
- `bun2nix` for frontend dependency locking
- `gomod2nix` for Go module locking
- Helm for Kubernetes deployment

## Repository layout

- `src/web`: Vue application built with Bun and served by the Go API
- `src/api`: Go service with the `GET /ping` endpoint
- `charts/thecluster`: Helm chart for Kubernetes deployment
- `default.nix`, `flake.nix`: Nix build and development entrypoints

## Local development

```bash
nix develop
bun install --cwd src/web
go test ./src/api/...
bun run --cwd src/web build
go run ./src/api/cmd/thecluster-api
```

The Go service serves static files from `src/web/dist` by default, so build the web app before starting the API locally.

## Regenerating Nix lock material

When frontend dependencies change:

```bash
bun install --cwd src/web
bun2nix --lock-file src/web/bun.lock --output-file src/web/bun.nix
```

When Go dependencies change:

```bash
gomod2nix generate --dir src/api
```

## Nix builds

```bash
nix build .#web
nix build .#api
nix build .#app
nix build .#docker
```

The Docker image is produced with `dockerTools.streamLayeredImage`.
