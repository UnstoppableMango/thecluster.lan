BUN       ?= bun
BUN2NIX   ?= bun2nix
GO        ?= go
GOMOD2NIX ?= gomod2nix
HELM      ?= helm
NIX       ?= nix

dev:
	$(NIX) develop

web-deps:
	$(BUN) install --cwd src/web

build-web: web-deps
	$(BUN) run --cwd src/web build

build-api:
	cd src/api && $(GO) build ./cmd/thecluster-api

build: build-web build-api

test:
	cd src/api && $(GO) test ./...

run: build-web
	$(GO) run ./src/api/cmd/thecluster-api

chart-lint:
	$(HELM) lint charts/thecluster

lint: chart-lint

check: test build-web chart-lint
	$(NIX) flake check --all-systems

nix-build:
	$(NIX) build .#web .#api .#app .#docker

nix-deps: web-deps
	$(BUN2NIX) --lock-file src/web/bun.lock --output-file src/web/bun.nix
	$(GOMOD2NIX) generate --dir src/api

flake-update:
	$(NIX) flake update

clean:
	rm -rf src/web/dist src/api/thecluster-api result result-*
