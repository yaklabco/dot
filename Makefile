.PHONY: build test test-tparse lint clean install uninstall check qa help version changelog-update coverage-summary cs build-all

# Build variables
BINARY_NAME := dot
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Parse semantic version components
MAJOR := $(shell echo $(CURRENT_VERSION) | sed 's/v//' | cut -d. -f1)
MINOR := $(shell echo $(CURRENT_VERSION) | sed 's/v//' | cut -d. -f2)
PATCH := $(shell echo $(CURRENT_VERSION) | sed 's/v//' | cut -d. -f3 | cut -d- -f1)

# LDFLAGS for version embedding
LDFLAGS := -ldflags "\
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)"

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application binary
build:
	go build -buildvcs=false $(LDFLAGS) -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

## test: Run all tests with race detection and coverage (parallelized across all CPUs)
test:
	go test -v -race -cover -coverprofile=coverage.out -parallel=$(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8) -p=$(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8) ./...

## lint: Run golangci-lint
lint:
	golangci-lint run --config .golangci.yml

## vet: Run go vet
vet:
	go vet ./...

## fmt: Check code formatting
fmt:
	@if [ "$(shell gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted:"; \
		gofmt -s -l .; \
		exit 1; \
	fi

## fmt-fix: Fix code formatting
fmt-fix:
	gofmt -s -w .
	goimports -w .

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -rf dist/

## install: Install the binary
install: build
	go install -buildvcs=false $(LDFLAGS) ./cmd/$(BINARY_NAME)

## uninstall: Remove the installed binary
uninstall:
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then \
		GOBIN=$$(go env GOPATH)/bin; \
	fi; \
	if [ -f "$$GOBIN/$(BINARY_NAME)" ]; then \
		rm -f "$$GOBIN/$(BINARY_NAME)"; \
		echo "✓ Removed $$GOBIN/$(BINARY_NAME)"; \
	else \
		echo "Binary not found at $$GOBIN/$(BINARY_NAME)"; \
	fi

## vuln: Check for known vulnerabilities (excludes GO-2024-3295 - see SECURITY.md)
vuln:
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	@echo "Running vulnerability check..."
	@govulncheck -json ./... 2>&1 | tee /tmp/vuln-output.json > /dev/null || true
	@VULN_IDS=$$(grep -A 1 '"finding"' /tmp/vuln-output.json 2>/dev/null | grep '"osv"' | grep -o 'GO-[0-9-]*' | sort -u || echo ""); \
	EXCLUDED="GO-2024-3295"; \
	if echo "$$VULN_IDS" | grep -q "$$EXCLUDED" 2>/dev/null; then \
		echo "  ℹ Found GO-2024-3295 (accepted risk - GitHub Codespaces only, see SECURITY.md)"; \
	fi; \
	OTHER_VULNS=$$(echo "$$VULN_IDS" | grep -v "^$$" | grep -v "$$EXCLUDED" || true); \
	if [ -n "$$OTHER_VULNS" ]; then \
		echo ""; \
		echo "ERROR: Unaccepted vulnerabilities found:"; \
		echo "$$OTHER_VULNS" | sed 's/^/  - /'; \
		echo ""; \
		govulncheck ./...; \
		rm -f /tmp/vuln-output.json; \
		exit 1; \
	fi
	@rm -f /tmp/vuln-output.json
	@echo "✓ No unaccepted vulnerabilities found (GO-2024-3295 excluded - see SECURITY.md)"

## fuzz: Run fuzzing tests (short duration)
fuzz:
	@echo "Running fuzzing tests for 30 seconds each..."
	@echo ""
	@echo "Fuzzing config parsing..."
	@go test -fuzz=FuzzLoadFromFile -fuzztime=30s ./internal/config/ -run=^$$
	@go test -fuzz=FuzzValidateExtended -fuzztime=30s ./internal/config/ -run=^$$
	@go test -fuzz=FuzzLoaderLoad -fuzztime=30s ./internal/config/ -run=^$$
	@echo ""
	@echo "Fuzzing ignore patterns..."
	@go test -fuzz=FuzzGlobToRegex -fuzztime=30s ./internal/ignore/ -run=^$$
	@go test -fuzz=FuzzPatternMatch -fuzztime=30s ./internal/ignore/ -run=^$$
	@go test -fuzz=FuzzIgnoreSetShouldIgnore -fuzztime=30s ./internal/ignore/ -run=^$$
	@echo ""
	@echo "Fuzzing domain path validation..."
	@go test -fuzz=FuzzNewPackagePath -fuzztime=30s ./internal/domain/ -run=^$$
	@go test -fuzz=FuzzNewTargetPath -fuzztime=30s ./internal/domain/ -run=^$$
	@go test -fuzz=FuzzNewFilePath -fuzztime=30s ./internal/domain/ -run=^$$
	@echo ""
	@echo "✓ All fuzzing tests passed"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./internal/planner/ ./internal/scanner/ ./internal/executor/

## bench-compare: Run benchmarks and compare with baseline
bench-compare:
	@if [ ! -f bench-baseline.txt ]; then \
		echo "Creating baseline..."; \
		go test -bench=. -benchmem -run=^$$ ./internal/planner/ ./internal/scanner/ ./internal/executor/ > bench-baseline.txt; \
		echo "✓ Baseline created in bench-baseline.txt"; \
		echo ""; \
		echo "Run 'make bench-compare' again after making changes to compare results."; \
	else \
		echo "Comparing against baseline..."; \
		go test -bench=. -benchmem -run=^$$ ./internal/planner/ ./internal/scanner/ ./internal/executor/ > bench-current.txt; \
		command -v benchstat >/dev/null 2>&1 || { echo "Installing benchstat..."; go install golang.org/x/perf/cmd/benchstat@latest; }; \
		benchstat bench-baseline.txt bench-current.txt; \
	fi

## check: Run tests and linting (machine-readable output for CI/AI agents)
check: test check-coverage lint vet vuln

## find-test-targets: Find functions/methods that need test coverage
find-test-targets:
	@go test ./internal/cli/pretty/... -coverprofile=pretty_cov.out
	@go tool cover -func=pretty_cov.out | grep -v "100.0%" | grep '\.go:'

## qa: Run tests with tparse, linting, vetting, and coverage summary (human-friendly output)
qa: test-tparse lint vet coverage-summary

## coverage-summary: Display coverage summary with threshold report (excludes UI files)
coverage-summary:
	@/bin/bash -c ' \
	echo ""; \
	echo "══════════════════════════════════════════════════════════"; \
	echo "Coverage Summary (excluding Bubble Tea UI files)"; \
	echo "══════════════════════════════════════════════════════════"; \
	if [ ! -f coverage.out ]; then \
		echo "Error: coverage.out not found"; \
		exit 1; \
	fi; \
	COVERAGE=$$(go tool cover -func=coverage.out | \
		grep -v "internal/cli/adopt/selector.go" | \
		grep -v "internal/cli/adopt/scanner.go" | \
		grep -v "internal/cli/adopt/interactive.go" | \
		grep -v "internal/cli/adopt/discovery.go" | \
		grep -v "^total:" | \
		awk "BEGIN {covered=0; total=0} \
			{if (NF==3) { \
				split(\$$3, a, \"%\"); \
				pct=a[1]; \
				covered += pct; \
				total++; \
			}} \
			END {if (total>0) printf \"%.1f\", covered/total; else print \"0\"}"); \
	THRESHOLD=60.0; \
	echo ""; \
	printf "  Total Coverage:     %6.1f%%\n" $$COVERAGE; \
	printf "  Required Threshold: %6.1f%%\n" $$THRESHOLD; \
	printf "  Excluded Files:     UI and interactive workflow files\n"; \
	echo ""; \
	if [ "$$(echo "$$COVERAGE < $$THRESHOLD" | bc)" -eq 1 ]; then \
		SHORTFALL=$$(echo "$$THRESHOLD - $$COVERAGE" | bc); \
		printf "  Status: ✗ BELOW THRESHOLD\n"; \
		printf "  Shortfall: %.1f%%\n" $$SHORTFALL; \
		echo ""; \
		echo "══════════════════════════════════════════════════════════"; \
		echo "Action Required: Add tests to reach $$THRESHOLD%% coverage"; \
		echo "══════════════════════════════════════════════════════════"; \
		exit 1; \
	else \
		SURPLUS=$$(echo "$$COVERAGE - $$THRESHOLD" | bc); \
		printf "  Status: ✓ PASSED\n"; \
		printf "  Surplus: +%.1f%%\n" $$SURPLUS; \
		echo ""; \
		echo "══════════════════════════════════════════════════════════"; \
	fi \
	'

## cs: Alias for coverage-summary
cs: coverage-summary

## test-tparse: Run tests with tparse for formatted output (parallelized across all CPUs)
test-tparse:
	@command -v tparse >/dev/null 2>&1 || { echo "Installing tparse..."; go install github.com/mfridman/tparse@latest; }
	@set -o pipefail && go test -json -race -cover -coverprofile=coverage.out -parallel=$(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8) -p=$(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8) ./... | tparse -all -progress -slow 10

## coverage: Generate coverage report (parallelized across all CPUs)
coverage:
	go test -coverprofile=coverage.out -parallel=$(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8) -p=$(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8) ./...
	go tool cover -html=coverage.out

## check-coverage: Verify test coverage meets 60% threshold (excludes UI files)
check-coverage:
	@if [ ! -f coverage.out ]; then \
		echo "Error: coverage.out not found. Run 'make test' first."; \
		exit 1; \
	fi
	@echo "Calculating coverage (excluding Bubble Tea UI files)..."; \
	COVERAGE=$$(go tool cover -func=coverage.out | \
		grep -v 'internal/cli/adopt/selector.go' | \
		grep -v 'internal/cli/adopt/scanner.go' | \
		grep -v 'internal/cli/adopt/interactive.go' | \
		grep -v 'internal/cli/adopt/discovery.go' | \
		grep -v '^total:' | \
		awk 'BEGIN {covered=0; total=0} \
			{if (NF==3) { \
				split($$3, a, "%"); \
				pct=a[1]; \
				covered += pct; \
				total++; \
			}} \
			END {if (total>0) printf "%.1f", covered/total; else print "0"}'); \
	THRESHOLD=60.0; \
	echo "Coverage: $${COVERAGE}% (threshold: $${THRESHOLD}%, UI files excluded)"; \
	if [ "$$(echo "$${COVERAGE} < $${THRESHOLD}" | bc)" -eq 1 ]; then \
		echo ""; \
		echo "ERROR: Test coverage below threshold"; \
		echo "  Current:   $${COVERAGE}%"; \
		echo "  Required:  $${THRESHOLD}%"; \
		echo "  Shortfall: $$(echo "$${THRESHOLD} - $${COVERAGE}" | bc)%"; \
		echo ""; \
		echo "Note: Bubble Tea UI and interactive workflow files are excluded from coverage"; \
		echo "Excluded files: selector.go, scanner.go, interactive.go, discovery.go"; \
		echo "Add tests to reach 60% coverage."; \
		echo "Run: go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out"; \
		exit 1; \
	fi; \
	echo "Coverage check passed: $${COVERAGE}%"

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## deps-update: Update all dependencies
deps-update:
	go get -u ./...
	go mod tidy

## deps-verify: Verify dependencies
deps-verify:
	go mod verify

## version: Display current version
version:
	@echo "Current version: $(CURRENT_VERSION)"

## changelog: Generate CHANGELOG.md locally for preview
changelog:
	@command -v git-chglog >/dev/null 2>&1 || { echo "Installing git-chglog..."; go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest; }
	git-chglog -o CHANGELOG.md

## changelog-next: Preview next version changelog
changelog-next:
	@command -v git-chglog >/dev/null 2>&1 || { echo "Installing git-chglog..."; go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest; }
	$(eval NEXT_VERSION := v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1))
	@echo "Preview of changelog for $(NEXT_VERSION):"
	@echo ""
	git-chglog --next-tag $(NEXT_VERSION) $(CURRENT_VERSION)..

## build-all: Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	GOOS=linux GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	GOOS=windows GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)

