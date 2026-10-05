# THECLUSTER

[![Hercules CI](https://hercules-ci.com/api/v1/site/github/account/UnstoppableMango/project/thecluster.lan/badge)](https://hercules-ci.com/github/UnstoppableMango/thecluster.lan)

Internal dashboard for `thecluster.lan`.

## Stack

- Vue + Vite + Tailwind CSS for the web UI
- Go for the API and static asset serving
- Nix for dependency management and builds
- `bun2nix` for frontend dependency locking
- `gomod2nix` for Go module locking
- Helm for Kubernetes deployment

## Repository layout

- `web`: Vue application built with Bun and served by the Go API
- `api`: Go service with `GET /ping` and `GET /api/nodes` (node health from Prometheus)
- `charts/thecluster`: Helm chart for Kubernetes deployment
- `flake.nix`, `flake.lock`: Nix build entrypoint and pinned inputs

## Local development

```bash
nix develop
make build
make test
make run
```

The Go service looks for static files in `../web/dist` and `web/dist` by default, so build the web app before starting the API locally.

### Dev loop

```bash
make dev
```

This runs the Vite dev server on port 5173 and the Go API on port 8080 under [air](https://github.com/air-verse/air).
Vite listens on all interfaces and proxies `/ping` to the API, so another machine on the LAN can open `http://<host-ip>:5173`.
Vue edits hot-reload in the browser.
Go edits rebuild and restart the API; a failed build leaves the last good binary running.
Browsing by hostname instead of IP requires adding the name to `server.allowedHosts` in `web/vite.config.ts`.

## Make targets

```bash
make check
make nix-build
make nix-deps
make update
```

## Regenerating Nix lock material

When frontend dependencies change:

```bash
cd web && bun install
```

When Go dependencies change:

```bash
make nix-deps
```

`web/package.json` runs `bun2nix bun.lock -o bun.nix` as a `postinstall` script, so `bun install` keeps the Bun v2 lock material in sync automatically.

## Nix builds

```bash
nix build .#web
nix build .#api
nix build .#app
nix build .#ctr
```

The container image is produced with `dockerTools.streamLayeredImage`.
