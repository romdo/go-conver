NAME := conver
ROOT := $(CURDIR)
SOURCES := $(shell find . -name "*.go" -or -name "go.mod" -or -name "go.sum" \
	-or -name "Makefile")

# Verbose output
ifdef VERBOSE
V = -v
endif

#
# Environment
#

BINDIR := bin
TOOLDIR := tools/bin

# Global environment variables for all targets
SHELL ?= /bin/bash
SHELL := env \
	GO111MODULE=on \
	GOBIN=$(CURDIR)/$(TOOLDIR) \
	CGO_ENABLED=1 \
	PATH='$(CURDIR)/$(BINDIR):$(CURDIR)/$(TOOLDIR):$(PATH)' \
	$(SHELL)

#
# Defaults
#

# Default target
.DEFAULT_GOAL := build

.PHONY: all
all: lint test build

#
# Tools
#

TOOLS += $(TOOLDIR)/gobin
gobin: $(TOOLDIR)/gobin
$(TOOLDIR)/gobin:
	GO111MODULE=off go get -u github.com/myitcv/gobin

# external tool
define tool # 1: binary-name, 2: go-import-path
TOOLS += $(TOOLDIR)/$(1)

.PHONY: $(1)
$(1): $(TOOLDIR)/$(1)

$(TOOLDIR)/$(1): $(TOOLDIR)/gobin Makefile
	gobin $(V) "$(2)"
endef

$(eval $(call tool,gofumports,mvdan.cc/gofumpt/gofumports))
$(eval $(call tool,golangci-lint,github.com/golangci/golangci-lint/cmd/golangci-lint@v1.31))

.PHONY: tools
tools: $(TOOLS)

#
# Build
#

BINARY=$(BINDIR)/$(NAME)
LDFLAGS := -w -s

VERSION ?= $(shell git describe --tags 2>/dev/null)
GIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE ?= $(shell date +%s)

ifeq ($(trim $(VERSION)),)
	VERSION = dev
endif

.PHONY: build
build: $(BINARY)

$(BINARY): $(SOURCES)
	go build $(V) -a -o "$@" -ldflags "$(LDFLAGS) \
		-X main.Version=$(VERSION) \
		-X main.Commit=$(GIT_SHA) \
		-X main.Date=$(DATE)"

#
# Development
#

.PHONY: clean
clean:
	rm -rf $(BINS) $(TOOLS)
	rm -f ./coverage.out ./go.mod.tidy-check ./go.sum.tidy-check

.PHONY: test
test:
	go test $(V) -count=1 --race ./...

.PHONY: lint
lint: golangci-lint
	$(info Running Go linters)
	GOGC=off golangci-lint $(V) run

.PHONY: format
format: gofumports
	gofumports -w .

#
# Coverage
#

.PHONY: cov
cov: coverage.out

.PHONY: cov-html
cov-html: coverage.out
	go tool cover -html=coverage.out

.PHONY: cov-func
cov-func: coverage.out
	go tool cover -func=coverage.out

coverage.out: $(SOURCES)
	go test $(V) -covermode=count -coverprofile=coverage.out ./...

#
# Dependencies
#

.PHONY: deps
deps:
	$(info Downloading dependencies)
	go mod download

.PHONY: tidy
tidy:
	go mod tidy $(V)

.PHONY: verify
verify:
	go mod verify

.SILENT: check-tidy
.PHONY: check-tidy
check-tidy:
	cp go.mod go.mod.tidy-check
	cp go.sum go.sum.tidy-check
	go mod tidy
	-diff go.mod go.mod.tidy-check
	-diff go.sum go.sum.tidy-check
	-rm -f go.mod go.sum
	-mv go.mod.tidy-check go.mod
	-mv go.sum.tidy-check go.sum
