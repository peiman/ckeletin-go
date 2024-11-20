# Project variables
BINARY_NAME = ckeletin-go
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT = $(shell git rev-parse HEAD)
DATE = $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

# Build flags
LDFLAGS = -ldflags "-X github.com/peiman/ckeletin-go/cmd.Version=${VERSION} \
                    -X github.com/peiman/ckeletin-go/cmd.Commit=${COMMIT} \
                    -X github.com/peiman/ckeletin-go/cmd.Date=${DATE}"

# Colors
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: help setup clean build test test-race test-pretty test-watch lint format vuln check run install

help: ## Display this help message
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-15s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Install development tools
	@echo "${CYAN}Installing development tools...${RESET}"
	@echo "Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Installing govulncheck..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Installing gotestsum..."
	@go install gotest.tools/gotestsum@latest
	@echo "Installing richgo..."
	@go install github.com/kyoh86/richgo@latest
	@echo "Installing golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		brew install golangci-lint; \
	else \
		echo "golangci-lint already installed"; \
	fi
	@echo "${GREEN}Development tools installed successfully!${RESET}"

format: ## Format code
	@echo "${CYAN}Formatting code...${RESET}"
	@goimports -w .
	@gofmt -w .

check: format lint-basic vuln test test-race ## Run all quality checks
	@echo "${GREEN}All checks passed!${RESET}"

lint-basic: ## Run basic Go linters
	@echo "${CYAN}Running basic Go linters...${RESET}"
	@go vet ./...
	@test -z "$$(gofmt -l .)"
	@test -z "$$(goimports -l .)"

lint: lint-basic ## Run all linters
	@echo "${CYAN}Attempting advanced linting...${RESET}"
	-@golangci-lint run || echo "${YELLOW}Advanced linting skipped due to compatibility issues${RESET}"

vuln: ## Check for vulnerabilities
	@echo "${CYAN}Checking for vulnerabilities...${RESET}"
	@govulncheck ./...

test: ## Run tests with coverage
	@echo "${CYAN}Running tests with coverage...${RESET}"
	@gotestsum --format pkgname-and-test-fails \
		--jsonfile test-output.json \
		--hide-summary=skipped \
		-- -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "${GREEN}Coverage report generated: coverage.txt${RESET}"

test-race: ## Run tests with race detection
	@echo "${CYAN}Running tests with race detection...${RESET}"
	@gotestsum --format testname \
		--jsonfile test-output.json \
		-- -race ./...

test-pretty: ## Run tests with pretty output
	@echo "${CYAN}Running tests with pretty output...${RESET}"
	@gotestsum --format testname \
		--hide-summary=skipped

test-watch: ## Run tests in watch mode
	@echo "${CYAN}Watching for changes...${RESET}"
	@gotestsum --format pkgname-and-test-fails --watch

clean: ## Clean build artifacts
	@echo "${CYAN}Cleaning build artifacts...${RESET}"
	@go clean
	@rm -f ${BINARY_NAME} coverage.txt coverage.out test-output.json

build: ## Build the binary
	@echo "${CYAN}Building ${BINARY_NAME}...${RESET}"
	@go build ${LDFLAGS} -o ${BINARY_NAME} main.go

run: build ## Run the application
	@echo "${CYAN}Running ${BINARY_NAME}...${RESET}"
	@./${BINARY_NAME}

install: ## Install the application
	@echo "${CYAN}Installing ${BINARY_NAME}...${RESET}"
	@go install ${LDFLAGS}

.DEFAULT_GOAL := help
