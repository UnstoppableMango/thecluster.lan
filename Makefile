BUN       ?= bun
BUN2NIX   ?= bun2nix
DOCKER	  ?= docker
GO        ?= go
GOMOD2NIX ?= gomod2nix
HELM      ?= helm
NIX       ?= nix

GO_SRC := $(shell find src/api -type f -name '*.go')
TS_SRC := $(shell find src/web -type f -name '*.ts')
JS_SRC := $(shell find src/web -type f -name '*.js')
VUE_SRC := $(shell find src/web -type f -name '*.vue')

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

load: bin/stream-image.sh
	${CURDIR}/bin/stream-image.sh | $(DOCKER) load

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

bin/stream-image.sh: ${GO_SRC} ${TS_SRC} ${JS_SRC} ${VUE_SRC}
	nix build .#ctr --out-link $@
bin/image.tar: bin/stream-image.sh
	${CURDIR}/bin/stream-image.sh > $@
