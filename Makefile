.PHONY: build test test-tparse lint clean install uninstall check qa help version version-major version-minor version-patch release release-tag changelog-update coverage-summary cs

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

## test: Run all tests with race detection and coverage
test:
	go test -v -race -cover -coverprofile=coverage.out ./...

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

## vuln: Check for known vulnerabilities
vuln:
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	@echo "Running vulnerability check (excluding GO-2024-3295 - Codespace-only issue)..."
	@govulncheck ./... 2>&1 | tee /tmp/vuln-output.txt || true
	@if grep "GO-2024-3295" /tmp/vuln-output.txt > /dev/null 2>&1; then \
		echo "  ℹ Found GO-2024-3295 (Codespace-only, accepted risk - see .github/SECURITY_EXCEPTIONS.md)"; \
	fi
	@VULN_COUNT=$$(grep -c "^Vulnerability #" /tmp/vuln-output.txt 2>/dev/null || echo "0"); \
	GO_2024_3295_COUNT=$$(grep -c "GO-2024-3295" /tmp/vuln-output.txt 2>/dev/null || echo "0"); \
	OTHER_VULNS=$$((VULN_COUNT - GO_2024_3295_COUNT)); \
	if [ $$OTHER_VULNS -gt 0 ]; then \
		echo ""; \
		echo "ERROR: $$OTHER_VULNS critical vulnerabilities found (excluding GO-2024-3295)"; \
		cat /tmp/vuln-output.txt; \
		rm -f /tmp/vuln-output.txt; \
		exit 1; \
	fi
	@rm -f /tmp/vuln-output.txt
	@echo "✓ No critical vulnerabilities found (GO-2024-3295 excluded - Codespace-only)"

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

## coverage-summary: Display coverage summary with threshold report
coverage-summary:
	@/bin/bash -c ' \
	echo ""; \
	echo "══════════════════════════════════════════════════════════"; \
	echo "Coverage Summary"; \
	echo "══════════════════════════════════════════════════════════"; \
	if [ ! -f coverage.out ]; then \
		echo "Error: coverage.out not found"; \
		exit 1; \
	fi; \
	COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk "{print \$$3}" | sed "s/%//"); \
	THRESHOLD=80.0; \
	echo ""; \
	printf "  Total Coverage:     %6.1f%%\n" $$COVERAGE; \
	printf "  Required Threshold: %6.1f%%\n" $$THRESHOLD; \
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

## test-tparse: Run tests with tparse for formatted output
test-tparse:
	@command -v tparse >/dev/null 2>&1 || { echo "Installing tparse..."; go install github.com/mfridman/tparse@latest; }
	@set -o pipefail && go test -json -race -cover -coverprofile=coverage.out ./... | tparse -all -progress -slow 10

## coverage: Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## check-coverage: Verify test coverage meets 80% threshold
check-coverage:
	@if [ ! -f coverage.out ]; then \
		echo "Error: coverage.out not found. Run 'make test' first."; \
		exit 1; \
	fi
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	THRESHOLD=80.0; \
	echo "Coverage: $${COVERAGE}% (threshold: $${THRESHOLD}%)"; \
	if [ "$$(echo "$${COVERAGE} < $${THRESHOLD}" | bc)" -eq 1 ]; then \
		echo ""; \
		echo "ERROR: Test coverage below threshold"; \
		echo "  Current:   $${COVERAGE}%"; \
		echo "  Required:  $${THRESHOLD}%"; \
		echo "  Shortfall: $$(echo "$${THRESHOLD} - $${COVERAGE}" | bc)%"; \
		echo ""; \
		echo "Add tests to reach 80% coverage."; \
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
	@echo "Next version (patch): v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1)"

## changelog: Generate CHANGELOG.md from git commits
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

## changelog-update: Update changelog and commit it (internal target)
changelog-update:
	@command -v git-chglog >/dev/null 2>&1 || { echo "Installing git-chglog..."; go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest; }
	@echo "Generating changelog..."
	@git-chglog -o CHANGELOG.md
	@echo "Committing changelog and configuration..."
	@git add CHANGELOG.md .chglog/
	@git commit -m "docs(changelog): update for $(VERSION) release" || true

## version-major: Bump major version and create release
version-major:
	$(eval NEW_VERSION := v$(shell expr $(MAJOR) + 1).0.0)
	@echo "Bumping major version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release-tag VERSION=$(NEW_VERSION)

## version-minor: Bump minor version and create release
version-minor:
	$(eval NEW_VERSION := v$(MAJOR).$(shell expr $(MINOR) + 1).0)
	@echo "Bumping minor version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release-tag VERSION=$(NEW_VERSION)

## version-patch: Bump patch version and create release
version-patch:
	$(eval NEW_VERSION := v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1))
	@echo "Bumping patch version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release-tag VERSION=$(NEW_VERSION)

## release: Verify release readiness (requires VERSION variable)
release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION not set. Use: make release VERSION=v1.2.3"; exit 1; fi
	@echo "Verifying release $(VERSION) readiness..."
	@$(MAKE) check
	@echo "✓ All quality checks passed"
	@echo "Release $(VERSION) verified and ready"

## release-tag: Complete release workflow with changelog and tagging
release-tag:
	@if [ -z "$(VERSION)" ]; then echo "VERSION not set. Use: make release-tag VERSION=v1.2.3"; exit 1; fi
	@echo "Starting release workflow for $(VERSION)..."
	@echo ""
	@echo "Step 1: Running quality checks..."
	@$(MAKE) check
	@echo "✓ Quality checks passed"
	@echo ""
	@echo "Step 2: Updating changelog..."
	@$(MAKE) changelog-update VERSION=$(VERSION)
	@echo "✓ Changelog updated"
	@echo ""
	@echo "Step 3: Creating git tag..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "✓ Tag $(VERSION) created"
	@echo ""
	@echo "Step 4: Regenerating changelog with new tag..."
	@$(MAKE) changelog
	@git add CHANGELOG.md
	@git commit --amend --no-edit
	@echo "✓ Changelog finalized"
	@echo ""
	@echo "Step 5: Moving tag to amended commit..."
	@git tag -f $(VERSION)
	@echo "✓ Tag moved to final commit"
	@echo ""
	@echo "══════════════════════════════════════════════════════════"
	@echo "Release $(VERSION) ready to push!"
	@echo "══════════════════════════════════════════════════════════"
	@echo ""
	@echo "Push with:"
	@echo "  git push origin main"
	@echo "  git push --force origin $(VERSION)"
	@echo ""

## build-all: Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	GOOS=linux GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	GOOS=windows GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)

