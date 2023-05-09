MAKEFLAG += --warn-undefined-variables
PROJECT_BRANCH ?="$(shell git rev-parse --abbrev-ref HEAD)"
PROJECT_SHA ?= $(shell git rev-parse HEAD)

BIN = netboard

PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin

.PHONY: netboard

netboard:
	CGO_ENABLED=1 go build \
		-ldflags "-X main.version=$(PROJECT_BRANCH) -X main.commit=$(PROJECT_SHA)" \
		-trimpath .

default: netboard

install: netboard
	install -d $(BINDIR)
	install -s $(BIN) $(BINDIR)

lint:
	golangci-lint run \
		--timeout=5m \
		--disable-all \
		--exclude-use-default=false \
		--exclude=package-comments \
		--exclude=unused-parameter \
		--enable=errcheck \
		--enable=goimports \
		--enable=ineffassign \
		--enable=revive \
		--enable=unused \
		--enable=staticcheck \
		--enable=unconvert \
		--enable=misspell \
		--enable=prealloc \
		--enable=nakedret \
		--enable=typecheck \
		--enable=unparam \
		--enable=gosimple \
		--enable=nilerr \
		./...
