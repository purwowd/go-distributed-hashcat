#!/bin/bash

# Go Distributed Hashcat - Test Runner for Reorganized Tests
# This script runs all tests from the tests/ folder with proper coverage reporting

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

# Function to run all tests and generate coverage for the main packages
run_tests_with_coverage() {
    print_header "Running All Tests with Coverage"
    
    # Remove old coverage files
    rm -f coverage.out coverage.html *.coverage 2>/dev/null || true
    
    print_info "Running tests and generating coverage for main packages..."
    
    # Run tests from tests folder but generate coverage for the actual source packages
    if go test $RACE_FLAG -timeout $TIMEOUT -v \
        -coverprofile="$COVER_PROFILE" \
        -coverpkg=./internal/usecase/...,./internal/delivery/http/handler/...,./internal/infrastructure/repository/... \
        ./tests/...; then
        print_success "All tests passed"
    else
        print_error "Some tests failed"
        return 1
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    print_header "Generating Coverage Report"
    
    if [ -f "$COVER_PROFILE" ]; then
        # Generate coverage statistics
        print_info "Calculating coverage statistics..."
        go tool cover -func="$COVER_PROFILE" | tail -20
        
        local coverage=$(go tool cover -func="$COVER_PROFILE" | tail -n 1 | awk '{print $3}')
        print_success "Total Coverage: $coverage"
        
        # Generate HTML coverage report
        go tool cover -html="$COVER_PROFILE" -o "$COVER_HTML"
        print_success "HTML coverage report generated: $COVER_HTML"
        
        # Coverage thresholds
        local threshold=70
        local coverage_num=$(echo "$coverage" | sed 's/%//')
        
        if (( $(echo "$coverage_num >= $threshold" | bc -l) )); then
            print_success "Coverage meets threshold: $coverage_num% >= $threshold%"
        else
            print_warning "Coverage below threshold: $coverage_num% < $threshold%"
        fi
        
        # Show package-level coverage breakdown
        print_info "Package-level coverage:"
        go tool cover -func="$COVER_PROFILE" | grep -E "(usecase|handler|repository)" | head -10
        
    else
        print_warning "No coverage data generated"
    fi
}

# Function to run specific test suites
run_specific_tests() {
    local test_type="$1"
    
    case $test_type in
        "unit")
            print_header "Running Unit Tests Only"
            go test $RACE_FLAG -timeout $TIMEOUT -v ./tests/unit/...
            ;;
        "integration")
            print_header "Running Integration Tests Only"
            go test $RACE_FLAG -timeout $TIMEOUT -v ./tests/integration/...
            ;;
        "usecase")
            print_header "Running Usecase Tests Only"
            go test $RACE_FLAG -timeout $TIMEOUT -v ./tests/unit/usecase/...
            ;;
        "handler")
            print_header "Running Handler Tests Only"
            go test $RACE_FLAG -timeout $TIMEOUT -v ./tests/unit/handler/...
            ;;
        "repository")
            print_header "Running Repository Tests Only"
            go test $RACE_FLAG -timeout $TIMEOUT -v ./tests/unit/repository/...
            ;;
        *)
            print_error "Unknown test type: $test_type"
            print_info "Available types: unit, integration, usecase, handler, repository"
            exit 1
            ;;
    esac
}

# Function to show test structure
show_test_structure() {
    print_header "Test Structure"
    print_info "Current test organization:"
    find tests/ -name "*.go" | sort
    echo
    print_info "Test counts by type:"
    echo "Unit tests: $(find tests/unit -name "*_test.go" | wc -l)"
    echo "Integration tests: $(find tests/integration -name "*_test.go" | wc -l)"
    echo "Benchmark tests: $(find tests/benchmarks -name "*_test.go" | wc -l)"
}

# Function to run quick smoke test
run_quick_test() {
    print_header "Running Quick Smoke Test"
    go test -timeout 30s ./tests/unit/usecase/... ./tests/unit/handler/... -v -short
}

# Main function
main() {
    case "${1:-all}" in
        "all")
            run_tests_with_coverage
            generate_coverage_report
            ;;
        "quick")
            run_quick_test
            ;;
        "coverage")
            run_tests_with_coverage
            generate_coverage_report
            ;;
        "structure")
            show_test_structure
            ;;
        "unit"|"integration"|"usecase"|"handler"|"repository")
            run_specific_tests "$1"
            ;;
        "help")
            echo "Usage: $0 [command]"
            echo "Commands:"
            echo "  all         - Run all tests with coverage (default)"
            echo "  quick       - Run quick smoke test"
            echo "  coverage    - Run tests and generate detailed coverage report"
            echo "  structure   - Show current test structure"
            echo "  unit        - Run only unit tests"
            echo "  integration - Run only integration tests"
            echo "  usecase     - Run only usecase tests"
            echo "  handler     - Run only handler tests"
            echo "  repository  - Run only repository tests"
            echo "  help        - Show this help message"
            ;;
        *)
            print_error "Unknown command: $1"
            $0 help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 
