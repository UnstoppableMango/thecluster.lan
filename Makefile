BUN       ?= bun
BUN2NIX   ?= bun2nix
GO        ?= go
GOMOD2NIX ?= gomod2nix
HELM      ?= helm
NIX       ?= nix

build: build-web build-api
lint: chart-lint

test:
	cd src/api && $(GO) test ./...

run: build-web
	$(GO) -C src/api run ./cmd/thecluster-api

update:
	$(NIX) flake update

check: test build-web chart-lint
	$(NIX) flake check --all-systems

clean:
	rm -rf src/web/dist src/api/thecluster-api result result-*

web-deps:
	$(BUN) install --cwd src/web

build-web: web-deps
	$(BUN) run --cwd src/web build
build-api:
	cd src/api && $(GO) build ./cmd/thecluster-api

chart-lint:
	$(HELM) lint charts/thecluster

nix-build:
	$(NIX) build .#web .#api .#app .#ctr --no-link

nix-deps: web-deps
	$(GOMOD2NIX) generate --dir src/api
	$(BUN2NIX) -l src/web/bun.lockb -o src/web/bun.nix
