# Tests Directory

This directory contains all tests for the Go Distributed Hashcat API project, organized by test type and layer.

## Directory Structure

```
tests/
├── unit/              # Unit tests with mocks
│   ├── handler/       # HTTP handler tests
│   ├── usecase/       # Business logic tests
│   └── repository/    # Data layer tests
├── integration/       # Integration tests with real dependencies
└── benchmarks/        # Performance benchmark tests
```

## Test Organization

### Unit Tests (`tests/unit/`)

Unit tests use mocks to isolate components and test individual functionality:

- **Handler Tests** (`tests/unit/handler/`): Test HTTP handlers with mocked use cases
  - `agent_handler_test.go` - Agent management endpoints
  - `job_handler_test.go` - Job management endpoints

- **Usecase Tests** (`tests/unit/usecase/`): Test business logic with mocked repositories
  - Test business rules and workflows
  - Validate input/output transformations

- **Repository Tests** (`tests/unit/repository/`): Test data layer with mocked database
  - Test database operations
  - Validate SQL queries and transactions

### Integration Tests (`tests/integration/`)

Integration tests use real dependencies and test complete workflows:

- `integration_test.go` - Full API test suite with real database
- Tests complete user workflows (agent registration, job creation, etc.)
- Uses temporary SQLite database for isolation

### Benchmark Tests (`tests/benchmarks/`)

Performance tests to measure API performance:

- `benchmark_test.go` - Performance benchmarks
- Agent creation performance
- Job creation performance
- Concurrent request handling
- API endpoint response times

## Running Tests

### Using the Test Runner Script

```bash
# Run all tests
./scripts/run_tests.sh --all

# Run only unit tests
./scripts/run_tests.sh --unit

# Run only integration tests
./scripts/run_tests.sh --integration

# Run only benchmarks
./scripts/run_tests.sh --benchmark

# Run with coverage report
./scripts/run_tests.sh --all --coverage

# Verbose output
./scripts/run_tests.sh --unit --verbose
```

### Using Make Commands

```bash
# Run all Go tests
make test-go

# Run unit tests only
make test-go-unit

# Run integration tests only
make test-go-integration

# Run benchmarks only
make test-go-benchmarks

# Run with coverage
make test-go-coverage
```

### Using Go Commands Directly

```bash
# Unit tests
go test ./tests/unit/...

# Integration tests
go test ./tests/integration/...

# Benchmarks
go test -bench=. ./tests/benchmarks/...

# With coverage
go test -cover -coverprofile=coverage.out ./tests/...
```

## Test Conventions

### Package Naming

- Unit tests use `package <layer>_test` (e.g., `package handler_test`)
- Integration tests use `package integration`
- Benchmark tests use `package benchmarks`

### Test File Naming

- Unit test files: `<component>_test.go`
- Integration test files: `integration_test.go`
- Benchmark test files: `benchmark_test.go`

### Mock Naming

- Mock structs: `Mock<Interface>` (e.g., `MockAgentUsecase`)
- Mock methods follow the same signature as the interface

### Test Function Naming

- Unit tests: `Test<Component>_<Method>` (e.g., `TestAgentHandler_RegisterAgent`)
- Integration tests: `Test<Workflow>` (e.g., `TestAgentWorkflow`)
- Benchmarks: `Benchmark<Operation>` (e.g., `BenchmarkAgentCreation`)

## Dependencies

The test suite uses the following Go testing libraries:

- `testing` - Standard Go testing package
- `github.com/stretchr/testify` - Assertions and mocking
  - `testify/assert` - Test assertions
  - `testify/mock` - Mock objects
  - `testify/suite` - Test suites for integration tests
- `net/http/httptest` - HTTP testing utilities
- `github.com/gin-gonic/gin` - Web framework (test mode)

## Coverage Reports

Coverage reports are generated in the `test-results/` directory:

- `coverage.out` - Raw coverage data
- `coverage.html` - HTML coverage report

View the HTML report:
```bash
open test-results/coverage.html
```

## Best Practices

1. **Test Isolation**: Each test should be independent and not rely on other tests
2. **Mock External Dependencies**: Use mocks for external services and databases in unit tests
3. **Real Dependencies in Integration**: Use real implementations in integration tests
4. **Clear Test Names**: Test names should clearly describe what is being tested
5. **Arrange-Act-Assert**: Structure tests with clear setup, execution, and assertion phases
6. **Test Data**: Use realistic but minimal test data
7. **Error Cases**: Test both success and error scenarios
8. **Performance**: Use benchmarks to monitor performance regressions

## Troubleshooting

### Common Issues

1. **Import Cycles**: Make sure test packages don't create import cycles
2. **Database Locks**: Clean up test databases properly in integration tests
3. **Mock Interface Mismatch**: Ensure mock implementations match actual interfaces
4. **Test Isolation**: Tests should not depend on external state or other tests

### Debug Tests

```bash
# Run specific test with verbose output
go test -v ./tests/unit/handler -run TestAgentHandler_RegisterAgent

# Run with race detection
go test -race ./tests/...

# Show test coverage for specific package
go test -cover ./tests/unit/handler
``` 
