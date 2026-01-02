# Variables
APP_NAME=notinoteapp
BINARY_DIR=bin
MAIN_PATH=cmd/server/main.go
MIGRATION_DIR=internal/adapters/secondary/database/postgres/migrations

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test coverage deps run dev migrate-up migrate-down migrate-create help

all: clean deps build

## help: Display this help message
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application binary"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with live reload (requires air)"
	@echo "  make test           - Run all tests"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests only"
	@echo "  make test-domain    - Run domain layer tests"
	@echo "  make test-service   - Run service layer tests"
	@echo "  make test-handler   - Run handler layer tests"
	@echo "  make test-repo      - Run repository tests"
	@echo "  make test-oauth     - Run OAuth provider tests"
	@echo "  make test-utils     - Run utility tests (JWT, password)"
	@echo "  make coverage       - Run tests with coverage report"
	@echo "  make coverage-html  - Generate HTML coverage report"
	@echo "  make bench          - Run benchmarks"
	@echo "  make deps           - Download dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make migrate-up     - Run database migrations up"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make migrate-create - Create new migration (use NAME=migration_name)"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start Docker Compose stack"
	@echo "  make docker-down    - Stop Docker Compose stack"

## build: Build the application binary
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_DIR)/$(APP_NAME)"

## run: Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

## dev: Run with live reload (requires air)
dev:
	@echo "Starting development server with live reload..."
	@air

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## test: Run all tests
test:
	@echo "Running all tests..."
	$(GOTEST) -v ./...

## test-unit: Run unit tests only
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./tests/unit/...

## test-integration: Run integration tests only
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./tests/integration/...

## test-domain: Run domain layer tests
test-domain:
	@echo "Running domain layer tests..."
	$(GOTEST) -v ./internal/core/domain/

## test-service: Run service layer tests
test-service:
	@echo "Running service layer tests..."
	$(GOTEST) -v ./internal/application/services/

## test-handler: Run handler layer tests
test-handler:
	@echo "Running handler layer tests..."
	$(GOTEST) -v ./internal/adapters/primary/http/handlers/

## test-repo: Run repository tests
test-repo:
	@echo "Running repository tests..."
	$(GOTEST) -v ./internal/adapters/secondary/database/postgres/repositories/

## test-oauth: Run OAuth provider tests
test-oauth:
	@echo "Running OAuth provider tests..."
	$(GOTEST) -v ./internal/adapters/secondary/oauth/

## test-utils: Run utility tests
test-utils:
	@echo "Running utility tests..."
	$(GOTEST) -v ./pkg/utils/

## test-race: Run tests with race detector (requires CGO)
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -v -race ./...

## coverage: Run tests with coverage report
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Coverage summary:"
	@$(GOCMD) tool cover -func=coverage.out | tail -n 1

## coverage-html: Generate HTML coverage report
coverage-html:
	@echo "Generating HTML coverage report..."
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Open coverage.html in your browser to view the report"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies downloaded"

## lint: Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Code formatted"

## migrate-up: Run database migrations up
migrate-up:
	@echo "Running migrations..."
	@migrate -path $(MIGRATION_DIR) -database "postgres://postgres:postgres@localhost:5432/notinoteapp?sslmode=disable" up
	@echo "Migrations complete"

## migrate-down: Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path $(MIGRATION_DIR) -database "postgres://postgres:postgres@localhost:5432/notinoteapp?sslmode=disable" down
	@echo "Rollback complete"

## migrate-create: Create new migration
migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Error: NAME is required. Use: make migrate-create NAME=migration_name"; exit 1; fi
	@echo "Creating migration: $(NAME)..."
	@migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)
	@echo "Migration created"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .
	@echo "Docker image built"

## docker-up: Start Docker Compose stack
docker-up:
	@echo "Starting Docker Compose stack..."
	docker-compose up -d
	@echo "Docker stack started"

## docker-down: Stop Docker Compose stack
docker-down:
	@echo "Stopping Docker Compose stack..."
	docker-compose down
	@echo "Docker stack stopped"

## docker-logs: View Docker Compose logs
docker-logs:
	docker-compose logs -f

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"
