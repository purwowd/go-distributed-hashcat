.PHONY: build build-server build-agent run-server run-agent clean test deps lint fmt vet tidy mod-verify
.PHONY: frontend-setup frontend-dev frontend-build frontend-install benchmark-api

# Go 1.24 build flags for performance optimization
BUILD_FLAGS := -ldflags="-s -w" -trimpath
GO_VERSION := 1.24

# Build targets
build: build-server build-agent

build-server:
	@echo "Building server with Go $(GO_VERSION) optimizations..."
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -o bin/server cmd/server/main.go

build-agent:
	@echo "Building agent with Go $(GO_VERSION) optimizations..."
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/agent cmd/agent/main.go

# Build for production with additional optimizations
build-prod: build-server-prod build-agent-prod frontend-build

build-server-prod:
	@echo "Building server for production..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/server-linux cmd/server/main.go

build-agent-prod:
	@echo "Building agent for production..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/agent-linux cmd/agent/main.go

# Frontend targets
frontend-install:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

frontend-setup:
	@echo "Setting up frontend development environment..."
	bash scripts/setup-frontend.sh

frontend-dev:
	@echo "Starting frontend development server..."
	cd frontend && npm run dev

frontend-build:
	@echo "Building frontend for production..."
	cd frontend && npm run build

frontend-preview:
	@echo "Previewing frontend production build..."
	cd frontend && npm run preview

# Full stack development
dev-full: dev-server-bg frontend-dev

dev-server-bg:
	@echo "Starting server in background..."
	mkdir -p data uploads
	./bin/server &

# Run targets
run-server:
	@echo "Starting server..."
	mkdir -p data uploads
	./bin/server

run-agent:
	@echo "Starting agent..."
	./bin/agent

# Development targets
dev-server:
	@echo "Running server in development mode..."
	mkdir -p data uploads
	go run cmd/server/main.go

dev-agent:
	@echo "Running agent in development mode..."
	go run cmd/agent/main.go

# Dependencies and module management
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

mod-verify:
	@echo "Verifying module dependencies..."
	go mod verify

tidy:
	@echo "Tidying go modules..."
	go mod tidy

# Code quality and formatting
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w -local go-distributed-hashcat .

vet:
	@echo "Running go vet..."
	go vet ./...

lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Install dev tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Testing - Modern Test Suite
test:
	@echo "🧪 Running comprehensive test suite..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh all

test-unit:
	@echo "🔧 Running unit tests..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh unit

test-integration:
	@echo "🔗 Running integration tests..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh integration

test-coverage:
	@echo "📊 Generating test coverage report..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh coverage

test-benchmark:
	@echo "🚀 Running benchmark tests..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh benchmark

test-lint:
	@echo "🔍 Running code quality checks..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh lint

test-security:
	@echo "🔒 Running security checks..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh security

test-clean:
	@echo "🧹 Cleaning test artifacts..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh clean

test-verbose:
	@echo "🔊 Running tests with verbose output..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh -v all

test-no-race:
	@echo "⚡ Running tests without race detection..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh --no-race all

test-quick:
	@echo "⚡ Running quick test suite (unit + lint)..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh unit && ./scripts/run_tests.sh lint

# Go testing suite - comprehensive test runner
test-go:
	@echo "Running Go test suite..."
	bash scripts/run_tests.sh --all

test-go-unit:
	@echo "Running Go unit tests..."
	bash scripts/run_tests.sh --unit --verbose

test-go-integration:
	@echo "Running Go integration tests..."
	bash scripts/run_tests.sh --integration --verbose

test-go-benchmarks:
	@echo "Running Go benchmarks..."
	bash scripts/run_tests.sh --benchmark

test-go-coverage:
	@echo "Running Go tests with coverage..."
	bash scripts/run_tests.sh --all --coverage

# API testing (both bash and Go)
test-api:
	@echo "Running comprehensive API tests..."
	bash scripts/test_api.sh

test-api-quick:
	@echo "Running quick API status check..."
	bash scripts/quick_test.sh

# Mock testing utilities
test-create-mocks:
	@echo "Creating mock test data..."
	@echo "Creating mock agents..."
	curl -s -X POST http://localhost:1337/api/v1/agents/ \
		-H "Content-Type: application/json" \
		-d '{"name":"Mock-Agent-1","ip_address":"192.168.1.100","port":8080,"capabilities":"Test GPU"}' > /dev/null
	@echo "✅ Mock test data created"

test-setup:
	@echo "Setting up test environment..."
	mkdir -p test-results
	chmod +x scripts/run_tests.sh scripts/test_api.sh scripts/quick_test.sh

# Combined testing workflow
test-all: test-setup test-go test-api
	@echo "🎉 All tests completed!"

benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Performance and testing
benchmark-api:
	@echo "Running API performance benchmark..."
	bash scripts/benchmark_api.sh

# Clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf data/
	rm -rf uploads/
	rm -f coverage.out coverage.html

# Docker targets
docker-build:
	@echo "Building Docker images..."
	docker build -t hashcat-server -f docker/Dockerfile.server .
	docker build -t hashcat-agent -f docker/Dockerfile.agent .

docker-run-server:
	@echo "Running server in Docker..."
	docker run -p 8080:8080 -v $(PWD)/data:/app/data -v $(PWD)/uploads:/app/uploads hashcat-server

docker-run-agent:
	@echo "Running agent in Docker..."
	docker run --network host hashcat-agent --server http://localhost:8080

# Database operations
db-migrate:
	@echo "Running database migrations..."
	mkdir -p data
	go run cmd/migrate/main.go

db-reset:
	@echo "Resetting database..."
	rm -f data/hashcat.db
	$(MAKE) db-migrate

# Initialize project
init:
	@echo "Initializing project..."
	mkdir -p bin data uploads configs web/static
	go mod download
	$(MAKE) install-tools

# Full setup including frontend
init-full: init frontend-install
	@echo "Full project initialization complete!"

# Development workflow
dev-setup: init deps install-tools frontend-setup
	@echo "Development setup complete!"

check: fmt vet lint test
	@echo "Code quality checks passed!"

# Agent setup
setup-agent:
	@echo "Setting up agent with local file structure..."
	bash scripts/setup-agent.sh

setup-agent-docker:
	@echo "Setting up agent directories for Docker..."
	mkdir -p agent-uploads/{wordlists,hash-files,temp}
	mkdir -p agent-uploads/wordlists/{common,leaked,custom}  
	mkdir -p agent-uploads/hash-files/{wifi,other}
	@echo "📁 Agent directories created in ./agent-uploads/"
	@echo "🐳 Mount with: -v $(PWD)/agent-uploads:/root/uploads"

# Download common wordlists
download-wordlists:
	@echo "Downloading popular wordlists..."
	mkdir -p wordlists/{common,leaked,custom}
	@echo "Downloading rockyou.txt..."
	@if [ ! -f "wordlists/common/rockyou.txt" ]; then \
		wget -q --show-progress -O wordlists/common/rockyou.txt \
		https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt; \
	fi
	@echo "Downloading common passwords..."
	@if [ ! -f "wordlists/common/10k-most-common.txt" ]; then \
		wget -q --show-progress -O wordlists/common/10k-most-common.txt \
		https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/10k-most-common.txt; \
	fi
	@echo "✅ Wordlists downloaded to ./wordlists/"

# Docker targets with file mounts
docker-run-server-with-files:
	@echo "Running server with file volumes..."
	docker run -p 1337:1337 \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/uploads:/app/uploads \
		-v $(PWD)/wordlists:/app/wordlists \
		hashcat-server

docker-run-agent-with-files:
	@echo "Running agent with local file access..."
	docker run --network host \
		-v $(PWD)/agent-uploads:/root/uploads \
		-v $(PWD)/wordlists:/root/uploads/wordlists \
		hashcat-agent --server http://localhost:1337

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build both server and agent"
	@echo "  build-prod    - Build for production (Linux)"
	@echo "  run-server    - Run the server"
	@echo "  run-agent     - Run the agent"
	@echo "  dev-server    - Run server in development mode"
	@echo "  dev-agent     - Run agent in development mode"
	@echo "  setup-agent   - Setup agent with local file structure"
	@echo "  download-wordlists - Download popular wordlists"
	@echo "  deps          - Download dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  test          - Run tests with race detection"
	@echo "  test-coverage - Generate coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-run-server-with-files - Run server with file volumes"
	@echo "  docker-run-agent-with-files  - Run agent with local files"
	@echo "  init          - Initialize project"
	@echo "  dev-setup     - Complete development setup"
	@echo "  check         - Run all code quality checks"
