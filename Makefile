
# Make does not offer a recursive wildcard function, so here's one:
rwildcard=$(wildcard $1$2) $(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2))

SHELL := /bin/bash
NAME := github-jira-bot
ORG := Digital-Transformation-Sandbox
ORG_REPO := $(ORG)/$(NAME)
RELEASE_ORG_REPO := $(ORG_REPO)

BUILD_TARGET = build
MAIN_SRC_FILE=main.go
GO := GO111MODULE=on go
GO_NOMOD :=GO111MODULE=off go
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES := $(call rwildcard,pkg/,*.go) $(call rwildcard,*.go)

REV := $(shell git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
ROOT_PACKAGE := github.com/$(ORG_REPO)

BRANCH     := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null  || echo 'unknown')
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
CGO_ENABLED = 0

REPORTS_DIR=$(BUILD_TARGET)/reports

GOTEST := $(GO) test
# If available, use gotestsum which provides more comprehensive output
# This is used in the CI builds
ifneq (, $(shell which gotestsum 2> /dev/null))
GOTESTSUM_FORMAT ?= standard-quiet
GOTEST := GO111MODULE=on gotestsum --junitfile $(REPORTS_DIR)/integration.junit.xml --format $(GOTESTSUM_FORMAT) --
endif

# set dev version unless VERSION is explicitly set via environment
VERSION ?= $(shell echo "$$(git describe --abbrev=0 --tags 2>/dev/null)-dev+$(REV)" | sed 's/^v//')

BUILDFLAGS :=  -ldflags \
  " -X $(ROOT_PACKAGE)/pkg/version.Version=$(VERSION)\
		-X $(ROOT_PACKAGE)/pkg/version.Revision='$(REV)'\
		-X $(ROOT_PACKAGE)/pkg/version.Branch='$(BRANCH)'\
		-X $(ROOT_PACKAGE)/pkg/version.BuildDate='$(BUILD_DATE)'\
		-X $(ROOT_PACKAGE)/pkg/version.GoVersion='$(GO_VERSION)'"

ifdef DEBUG
BUILDFLAGS := -gcflags "all=-N -l" $(BUILDFLAGS)
endif

ifdef PARALLEL_BUILDS
BUILDFLAGS += -p $(PARALLEL_BUILDS)
GOTEST += -p $(PARALLEL_BUILDS)
else
# -p 4 seems to work well for people
GOTEST += -p 4
endif

ifdef DISABLE_TEST_CACHING
GOTEST += -count=1
endif


.PHONY: list
list: ## List all make targets
	@$(MAKE) -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: build ## Build the binary

print-version: ## Print version
	@echo $(VERSION)

build: $(GO_DEPENDENCIES) ## Build binary for current OS
	CGO_ENABLED=$(CGO_ENABLED) $(GO) $(BUILD_TARGET) $(BUILDFLAGS) -o build/$(NAME) $(MAIN_SRC_FILE)

build-all: $(GO_DEPENDENCIES) build make-reports-dir ## Build all files - runtime, all tests etc.
	CGO_ENABLED=$(CGO_ENABLED) $(GOTEST) -run=nope -tags=integration -failfast -short ./...

.PHONY: build-covered
build-covered: $(GO_DEPENDENCIES) ## Build binary for current OS with coverage instrumentation to build/$(NAME).covered
	CGO_ENABLED=$(CGO_ENABLED) $(GO) $(COVERAGE_BUILD_TARGET) $(BUILDFLAGS) $(COVERAGE_BUILDFLAGS) -o build/$(NAME).covered $(COVERED_MAIN_SRC_FILE)

tidy-deps: ## Cleans up dependencies
	$(GO) mod tidy
	# mod tidy only takes compile dependencies into account, let's make sure we capture tooling dependencies as well
	@$(MAKE) install-generate-deps

install: $(GO_DEPENDENCIES) ## Install the binary
	GOBIN=${GOPATH}/bin $(GO) install $(BUILDFLAGS) $(MAIN_SRC_FILE)

linux: ## Build for Linux
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) $(BUILD_TARGET) $(BUILDFLAGS) -o build/linux/$(NAME)-linux-amd64 $(MAIN_SRC_FILE)
	chmod +x build/linux/$(NAME)-linux-amd64

arm: ## Build for ARM
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm $(GO) $(BUILD_TARGET) $(BUILDFLAGS) -o build/arm/$(NAME)-linux-arm $(MAIN_SRC_FILE)
	chmod +x build/arm/$(NAME)-linux-arm

win: ## Build for Windows
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 $(GO) $(BUILD_TARGET) $(BUILDFLAGS) -o build/win/$(NAME)-windows-amd64.exe $(MAIN_SRC_FILE)

win32:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=386 $(GO) $(BUILD_TARGET) $(BUILDFLAGS) -o build/win32/$(NAME)-win-386.exe $(MAIN_SRC_FILE)

darwin: ## Build for OSX
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 $(GO) $(BUILD_TARGET) $(BUILDFLAGS) -o build/darwin/$(NAME)-macOS $(MAIN_SRC_FILE)
	chmod +x build/darwin/$(NAME)-macOS

.PHONY: clean
clean: ## Clean the generated artifacts
	rm -rf build release dist

fmt: ## Format the code
	$(eval FORMATTED = $(shell $(GO) fmt ./...))
	@if [ "$(FORMATTED)" == "" ]; \
      	then \
      	    echo "All Go files properly formatted"; \
      	else \
      		echo "Fixed formatting for: $(FORMATTED)"; \
      	fi

.PHONY: lint
lint: ## Lint the code
	./hack/gofmt.sh
	./hack/linter.sh
	./hack/generate.sh

all: linux arm win win32 darwin
