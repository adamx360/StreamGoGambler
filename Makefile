.PHONY: build test test-race test-cover fuzz clean run help lint lint-fix fmt pre-commit install-tools docker docker-build docker-run docker-stop version mutate mutate-dry mutate-full mock-gen mock-clean

BINARY=streamgogambler
BINARY_WIN=$(BINARY).exe

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOFMT=gofmt

GOLANGCI_LINT=golangci-lint

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Default target
all: lint test build

## build: Build the binary with version info
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY) ./cmd/streamgogambler

## build-dev: Build without version info (faster)
build-dev:
	$(GOBUILD) -o $(BINARY) ./cmd/streamgogambler

## build-windows: Build Windows executable
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WIN) ./cmd/streamgogambler

## build-linux: Build Linux executable
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY)-linux ./cmd/streamgogambler

## build-all: Build for all platforms
build-all: build-windows build-linux
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY)-darwin-amd64 ./cmd/streamgogambler
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY)-darwin-arm64 ./cmd/streamgogambler

## version: Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(BUILD_DATE)"

## test: Run all tests
test:
	$(GOTEST) -v ./...

## test-short: Run tests in short mode
test-short:
	$(GOTEST) -short ./...

## test-race: Run tests with race detector (requires CGO)
test-race:
	CGO_ENABLED=1 $(GOTEST) -race -v ./...

## test-cover: Run tests with coverage report
test-cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-cover-func: Show coverage by function
test-cover-func:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

## fuzz: Run fuzz tests (10s each)
fuzz:
	$(GOTEST) -run=^$$ -fuzz=FuzzParseBombs -fuzztime=10s ./internal/domain/parsing
	$(GOTEST) -run=^$$ -fuzz=FuzzParseSlotsDelta -fuzztime=10s ./internal/domain/parsing
	$(GOTEST) -run=^$$ -fuzz=FuzzParsePoints -fuzztime=10s ./internal/domain/parsing

## fuzz-long: Run extended fuzz tests (60s each)
fuzz-long:
	$(GOTEST) -run=^$$ -fuzz=FuzzParseBombs -fuzztime=60s ./internal/domain/parsing
	$(GOTEST) -run=^$$ -fuzz=FuzzParseSlotsDelta -fuzztime=60s ./internal/domain/parsing
	$(GOTEST) -run=^$$ -fuzz=FuzzParsePoints -fuzztime=60s ./internal/domain/parsing

## lint: Run golangci-lint
lint:
	$(GOLANGCI_LINT) run --timeout=5m

## lint-fix: Run golangci-lint with auto-fix
lint-fix:
	$(GOLANGCI_LINT) run --fix --timeout=5m

## fmt: Format code with gofmt
fmt:
	$(GOFMT) -s -w .

## vet: Run go vet
vet:
	$(GOCMD) vet ./...

## pre-commit: Run pre-commit hooks on all files
pre-commit:
	pre-commit run --all-files

## pre-commit-install: Install pre-commit hooks
pre-commit-install:
	pre-commit install

## install-tools: Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
	go install github.com/vektra/mockery/v2@latest
	@echo "Install pre-commit: pip install pre-commit"
	@echo "Then run: make pre-commit-install"

## mutate: Run mutation testing (quick)
mutate:
	gremlins unleash ./...

## mutate-dry: Analyze mutations without running tests
mutate-dry:
	gremlins unleash --dry-run ./...

## mutate-full: Run full mutation testing with all mutant types
mutate-full:
	gremlins unleash \
		--invert-assignments \
		--invert-bitwise \
		--invert-logical \
		--invert-loopctrl \
		--remove-self-assignments \
		./...

## mock-gen: Generate mocks using mockery
mock-gen:
	mockery

## mock-clean: Remove generated mocks
mock-clean:
	rm -rf internal/mocks

## docker-build: Build Docker image
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t streamgogambler:$(VERSION) \
		-t streamgogambler:latest .

## docker-run: Run bot in Docker container
docker-run:
	docker compose up -d

## docker-stop: Stop Docker container
docker-stop:
	docker compose down

## docker-logs: Show Docker container logs
docker-logs:
	docker compose logs -f

## clean: Remove build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY) $(BINARY_WIN) $(BINARY)-linux $(BINARY)-darwin-*
	rm -f coverage.out coverage.html

## run: Build and run the bot
run: build
	./$(BINARY)

## deps: Verify and tidy dependencies
deps:
	$(GOMOD) verify
	$(GOMOD) tidy

## check: Run all checks (lint, test, build)
check: lint test build
	@echo "All checks passed!"

## ci: Run CI pipeline locally
ci: deps lint test-race test-cover build
	@echo "CI pipeline completed!"

## help: Show this help message
help:
	@echo "StreamGoGambler Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
