#!/bin/bash

# Go Distributed Hashcat - Comprehensive Test Runner
# This script runs all tests with coverage reporting and various test modes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TIMEOUT="300s"
RACE_FLAG="-race"
COVER_PROFILE="coverage.out"
COVER_HTML="coverage.html"
VERBOSE=""

# Function to print colored output
print_header() {
    echo -e "${BLUE}===========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}===========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Function to check if go is installed
check_dependencies() {
    print_header "Checking Dependencies"
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    print_success "Go version: $(go version)"
    
    # Check if required test packages are available
    if ! go list github.com/stretchr/testify &> /dev/null; then
        print_warning "Installing testify..."
        go mod tidy
    fi
    
    print_success "Dependencies verified"
}

# Function to run unit tests
run_unit_tests() {
    print_header "Running Unit Tests"
    
    local test_dirs=(
        "./internal/usecase/..."
        "./internal/delivery/http/handler/..."
        "./internal/infrastructure/repository/..."
    )
    
    for dir in "${test_dirs[@]}"; do
        print_info "Testing: $dir"
        if go test $RACE_FLAG -timeout $TIMEOUT $VERBOSE -coverprofile="${dir//\//_}.coverage" $dir; then
            print_success "Unit tests passed: $dir"
        else
            print_error "Unit tests failed: $dir"
            return 1
        fi
    done
}

# Function to run integration tests
run_integration_tests() {
    print_header "Running Integration Tests"
    
    # Integration tests typically test the full stack
    local integration_dirs=(
        "./tests/integration/..."
    )
    
    for dir in "${integration_dirs[@]}"; do
        if [ -d "${dir%/...}" ]; then
            print_info "Testing: $dir"
            if go test $RACE_FLAG -timeout $TIMEOUT $VERBOSE -tags=integration $dir; then
                print_success "Integration tests passed: $dir"
            else
                print_error "Integration tests failed: $dir"
                return 1
            fi
        else
            print_warning "Integration test directory not found: $dir"
        fi
    done
}

# Function to run benchmark tests
run_benchmark_tests() {
    print_header "Running Benchmark Tests"
    
    local benchmark_pattern="./..."
    
    print_info "Running benchmarks: $benchmark_pattern"
    if go test -bench=. -benchmem -run=^$ $benchmark_pattern; then
        print_success "Benchmarks completed"
    else
        print_warning "Some benchmarks failed or no benchmarks found"
    fi
}

# Function to generate coverage report
generate_coverage() {
    print_header "Generating Coverage Report"
    
    # Combine all coverage files
    echo "mode: set" > "$COVER_PROFILE"
    find . -name "*.coverage" -exec tail -n +2 {} \; >> "$COVER_PROFILE"
    
    # Clean up individual coverage files
    find . -name "*.coverage" -delete
    
    if [ -f "$COVER_PROFILE" ]; then
        # Generate coverage statistics
        local coverage=$(go tool cover -func="$COVER_PROFILE" | tail -n 1 | awk '{print $3}')
        print_success "Total Coverage: $coverage"
        
        # Generate HTML coverage report
        go tool cover -html="$COVER_PROFILE" -o "$COVER_HTML"
        print_success "HTML coverage report generated: $COVER_HTML"
        
        # Coverage thresholds
        local threshold=80
        local coverage_num=$(echo "$coverage" | sed 's/%//')
        
        if (( $(echo "$coverage_num >= $threshold" | bc -l) )); then
            print_success "Coverage meets threshold: $coverage_num% >= $threshold%"
        else
            print_warning "Coverage below threshold: $coverage_num% < $threshold%"
        fi
    else
        print_warning "No coverage data generated"
    fi
}

# Function to run linting
run_linting() {
    print_header "Running Code Quality Checks"
    
    # golangci-lint is the standard for Go linting
    if command -v golangci-lint &> /dev/null; then
        print_info "Running golangci-lint..."
        if golangci-lint run ./...; then
            print_success "Linting passed"
        else
            print_error "Linting failed"
            return 1
        fi
    else
        print_warning "golangci-lint not installed, skipping lint checks"
        print_info "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.54.2"
    fi
    
    # Go vet
    print_info "Running go vet..."
    if go vet ./...; then
        print_success "go vet passed"
    else
        print_error "go vet failed"
        return 1
    fi
    
    # Go fmt
    print_info "Checking code formatting..."
    if [ -n "$(go fmt ./...)" ]; then
        print_error "Code is not properly formatted. Run 'go fmt ./...'"
        return 1
    else
        print_success "Code formatting verified"
    fi
}

# Function to run security checks
run_security_checks() {
    print_header "Running Security Checks"
    
    # gosec for security vulnerabilities
    if command -v gosec &> /dev/null; then
        print_info "Running gosec security scanner..."
        if gosec -quiet ./...; then
            print_success "Security scan passed"
        else
            print_warning "Security issues found (check output above)"
        fi
    else
        print_warning "gosec not installed, skipping security checks"
        print_info "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
    fi
    
    # Check for known vulnerabilities in dependencies
    print_info "Checking for vulnerable dependencies..."
    if go list -json -deps ./... | nancy sleuth; then
        print_success "No known vulnerabilities found"
    else
        print_warning "Potential vulnerabilities found in dependencies"
    fi 2>/dev/null || print_info "Install nancy for vulnerability scanning: go install github.com/sonatypecommunity/nancy@latest"
}

# Function to run all tests
run_all_tests() {
    print_header "Running Complete Test Suite"
    
    local failed=0
    
    # Run different test types
    run_unit_tests || failed=1
    run_integration_tests || failed=1
    
    if [ $failed -eq 1 ]; then
        print_error "Some tests failed"
        return 1
    else
        print_success "All tests passed"
        return 0
    fi
}

# Function to clean test artifacts
clean_artifacts() {
    print_header "Cleaning Test Artifacts"
    
    rm -f "$COVER_PROFILE" "$COVER_HTML"
    find . -name "*.coverage" -delete
    find . -name "*.test" -delete
    
    print_success "Test artifacts cleaned"
}

# Function to show help
show_help() {
    echo "Go Distributed Hashcat Test Runner"
    echo ""
    echo "Usage: $0 [OPTIONS] [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  unit          Run unit tests only"
    echo "  integration   Run integration tests only"
    echo "  benchmark     Run benchmark tests"
    echo "  coverage      Generate coverage report"
    echo "  lint          Run code quality checks"
    echo "  security      Run security checks"
    echo "  all           Run complete test suite (default)"
    echo "  clean         Clean test artifacts"
    echo ""
    echo "Options:"
    echo "  -v, --verbose     Enable verbose output"
    echo "  -r, --no-race     Disable race detection"
    echo "  -t, --timeout     Set test timeout (default: 300s)"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                # Run all tests"
    echo "  $0 unit           # Run only unit tests"
    echo "  $0 -v coverage    # Generate coverage with verbose output"
    echo "  $0 --no-race lint # Run linting without race detection"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -r|--no-race)
            RACE_FLAG=""
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        unit)
            COMMAND="unit"
            shift
            ;;
        integration)
            COMMAND="integration"
            shift
            ;;
        benchmark)
            COMMAND="benchmark"
            shift
            ;;
        coverage)
            COMMAND="coverage"
            shift
            ;;
        lint)
            COMMAND="lint"
            shift
            ;;
        security)
            COMMAND="security"
            shift
            ;;
        all)
            COMMAND="all"
            shift
            ;;
        clean)
            COMMAND="clean"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Default command
COMMAND=${COMMAND:-"all"}

# Main execution
main() {
    local start_time=$(date +%s)
    
    print_header "Go Distributed Hashcat Test Suite"
    print_info "Command: $COMMAND"
    print_info "Timeout: $TIMEOUT"
    print_info "Race Detection: $([ -n "$RACE_FLAG" ] && echo "enabled" || echo "disabled")"
    print_info "Verbose: $([ -n "$VERBOSE" ] && echo "enabled" || echo "disabled")"
    echo ""
    
    # Execute based on command
    case $COMMAND in
        unit)
            check_dependencies
            run_unit_tests
            ;;
        integration)
            check_dependencies
            run_integration_tests
            ;;
        benchmark)
            check_dependencies
            run_benchmark_tests
            ;;
        coverage)
            check_dependencies
            run_unit_tests
            generate_coverage
            ;;
        lint)
            check_dependencies
            run_linting
            ;;
        security)
            check_dependencies
            run_security_checks
            ;;
        all)
            check_dependencies
            run_linting
            run_all_tests
            generate_coverage
            run_security_checks
            ;;
        clean)
            clean_artifacts
            ;;
        *)
            print_error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    print_header "Test Summary"
    print_success "Total execution time: ${duration}s"
    
    if [ $? -eq 0 ]; then
        print_success "Test suite completed successfully!"
        exit 0
    else
        print_error "Test suite failed!"
        exit 1
    fi
}

# Run main function
main "$@"
