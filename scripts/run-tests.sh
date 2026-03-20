#!/bin/bash
# Watchtower Test Runner Script for Linux/Mac
# Run all tests with coverage reporting

set -e

echo "========================================"
echo "  Watchtower Test Runner"
echo "========================================"
echo ""

# Check if Go is installed
echo "Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo "✗ Go not found. Please install Go from https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version)
echo "✓ Go found: $GO_VERSION"
echo ""

# Run tests
echo "Running tests..."
echo ""

go test ./... -v -coverprofile=coverage.out -covermode=count

echo ""
echo "✓ All tests passed!"
echo ""

# Generate coverage report
echo "Generating coverage report..."

if [ -f coverage.out ]; then
    go tool cover -func=coverage.out > coverage-report.txt
    echo "✓ Coverage report saved to coverage-report.txt"
    
    # Display coverage summary
    echo ""
    echo "Coverage Summary:"
    go tool cover -func=coverage.out | grep total
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    echo "✓ HTML coverage report saved to coverage.html"
    echo "  Open coverage.html in your browser to view detailed coverage"
else
    echo "✗ No coverage data generated"
fi

echo ""
echo "========================================"
echo "  Test run completed successfully!"
echo "========================================"
