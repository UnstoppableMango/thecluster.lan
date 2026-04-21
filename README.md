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
make build
make test
make run
```

The Go service serves static files from `src/web/dist` by default, so build the web app before starting the API locally.

## Make targets

```bash
make check
make nix-build
make nix-deps
make flake-update
```

## Regenerating Nix lock material

When frontend dependencies change:

```bash
cd src/web && bun install
```

When Go dependencies change:

```bash
make nix-deps
```

`src/web/package.json` runs `bun2nix -o bun.nix` as a `postinstall` script, so `bun install` keeps the Bun v2 lock material in sync automatically.

## Nix builds

```bash
nix build .#web
nix build .#api
nix build .#app
nix build .#ctr
```

The container image is produced with `dockerTools.streamLayeredImage`.
