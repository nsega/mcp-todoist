# Go parameters
BINARY_NAME=mcp-todoist
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_CLEAN=$(GO_CMD) clean
GO_TEST=$(GO_CMD) test
GO_GET=$(GO_CMD) get
GO_MOD=$(GO_CMD) mod
GO_VET=$(GO_CMD) vet
GO_FMT=$(GO_CMD) fmt

# Tool versions. This is the single source of truth for both linter pins: make lint and
# make staticcheck install and run exactly these versions, locally and in CI, so the two
# cannot diverge. No leading "v": the install recipe adds it. They are pinned rather than
# @latest because setup-go v6+ exports GOTOOLCHAIN=local, so an upstream go directive
# above ours would fail the build.
#
# Why the pins move with the go directive (canonical explanation; README and CLAUDE.md
# point here):
#  - golangci-lint refuses a module whose language version (major.minor) is newer than
#    the Go it was built with: "can't load config: the Go language version (goX) used
#    to build golangci-lint is lower than the targeted Go version (Y)".
#  - golangci-lint and staticcheck embed golang.org/x/tools' export-data reader, so a Go
#    release that changes the export data format needs tools built against a newer
#    x/tools, even across patch releases. go1.27.2 moved to export data version 5:
#    staticcheck fails with "export data version 5 is greater than maximum supported
#    version 4", golangci-lint with spurious typecheck errors. golangci-lint 2.13.x
#    fails on it (2.14.0 happened to work when tested). The pins track the current
#    release, not the oldest release that works.
GOLANGCI_LINT_VERSION=2.15.0
STATICCHECK_VERSION=0.8.2

GOLANGCI_LINT_PKG=github.com/golangci/golangci-lint/v2/cmd/golangci-lint
STATICCHECK_PKG=honnef.co/go/tools/cmd/staticcheck

# Pinned linters live in a repo-local, git-ignored directory, never on PATH. Each binary's
# name carries its pin and the Go release that built it, so changing either one misses
# the existing file and reinstalls: a linter built by an older Go fails for the reasons
# above even when its own pin has not moved.
TOOLS_BIN=bin
GO_VERSION:=$(shell $(GO_CMD) env GOVERSION)
GOLANGCI_LINT=$(TOOLS_BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(GO_VERSION)
STATICCHECK=$(TOOLS_BIN)/staticcheck-$(STATICCHECK_VERSION)-$(GO_VERSION)

# install_tool installs a pinned linter into TOOLS_BIN and drops any older build of it.
# A failed install fails the target, so a missing linter can never pass make check.
# Arguments: tool name, package, pinned version.
define install_tool
	@echo "Installing $(1) $(3) into $(TOOLS_BIN)/..."
	@mkdir -p $(TOOLS_BIN)
	@GOBIN="$(CURDIR)/$(TOOLS_BIN)" $(GO_CMD) install $(2)@v$(3) || { \
		echo "Could not install $(1) $(3), so it cannot run and this target fails rather than skip it."; \
		echo "Check that the Go module proxy is reachable, then retry."; \
		exit 1; \
	}
	@rm -f $(TOOLS_BIN)/$(1)-*
	@mv $(TOOLS_BIN)/$(1) $@
endef

# Build parameters
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Directories
BUILD_DIR=build
COVERAGE_DIR=coverage

.PHONY: all build clean test lint staticcheck fmt vet deps help install run check coverage tidy verify

# Default target
all: check build

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  make build        - Build the binary"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make coverage     - Run tests with coverage report"
	@echo "  make lint         - Run the pinned golangci-lint"
	@echo "  make staticcheck  - Run the pinned staticcheck"
	@echo "  make fmt          - Format code"
	@echo "  make vet          - Run go vet"
	@echo "  make check        - Run all checks (fmt, vet, lint, staticcheck, test)"
	@echo "  make deps         - Download dependencies"
	@echo "  make tidy         - Tidy and verify dependencies"
	@echo "  make verify       - Verify dependencies"
	@echo "  make install      - Install the binary"
	@echo "  make run          - Run the application (requires TODOIST_API_TOKEN environment variable)"
	@echo "  make all          - Run checks and build"

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

## clean: Remove build artifacts and test cache
clean:
	@echo "Cleaning..."
	@$(GO_CLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	@$(GO_TEST) -v -race -timeout 30s ./...

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO_TEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@$(GO_CMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"
	@$(GO_CMD) tool cover -func=$(COVERAGE_DIR)/coverage.out

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GO_FMT) ./...
	@echo "Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GO_VET) ./...
	@echo "go vet passed"

## lint: Run the pinned golangci-lint
lint: $(GOLANGCI_LINT)
	@echo "Running golangci-lint..."
	@$(GOLANGCI_LINT) run --timeout=5m ./...

$(GOLANGCI_LINT):
	$(call install_tool,golangci-lint,$(GOLANGCI_LINT_PKG),$(GOLANGCI_LINT_VERSION))

## staticcheck: Run the pinned staticcheck
staticcheck: $(STATICCHECK)
	@echo "Running staticcheck..."
	@$(STATICCHECK) ./...

$(STATICCHECK):
	$(call install_tool,staticcheck,$(STATICCHECK_PKG),$(STATICCHECK_VERSION))

## check: Run all checks (fmt, vet, lint, staticcheck, test)
check: fmt vet lint staticcheck test
	@echo "All checks passed!"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GO_GET) -v -t -d ./...
	@echo "Dependencies downloaded"

## tidy: Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	@$(GO_MOD) tidy
	@$(GO_MOD) verify
	@echo "Dependencies tidied and verified"

## verify: Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@$(GO_MOD) verify
	@echo "Dependencies verified"

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@$(GO_CMD) install $(LDFLAGS)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## run: Run the application (requires TODOIST_API_TOKEN environment variable)
run: build
	@if [ -z "$(TODOIST_API_TOKEN)" ]; then \
		echo "Error: TODOIST_API_TOKEN environment variable is required"; \
		echo "Usage: make run TODOIST_API_TOKEN=your_token_here"; \
		exit 1; \
	fi
	@echo "Running $(BINARY_NAME)..."
	@TODOIST_API_TOKEN=$(TODOIST_API_TOKEN) $(BUILD_DIR)/$(BINARY_NAME)

# Advanced targets

## build-all: Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Multi-platform builds complete"

## release: Build optimized release binaries
release:
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO_BUILD) $(LDFLAGS) -a -installsuffix cgo -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"
