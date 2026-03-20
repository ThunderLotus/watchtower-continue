# Watchtower Test Runner Script for Windows
# Run all tests with coverage reporting

$ErrorActionPreference = "Stop"

Write-Host "========================================"  -ForegroundColor Cyan
Write-Host "  Watchtower Test Runner"  -ForegroundColor Cyan
Write-Host "========================================"  -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✓ Go found: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go not found. Please install Go from https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Run tests
Write-Host "Running tests..." -ForegroundColor Yellow
Write-Host ""

$testResults = go test ./... -v -coverprofile=coverage.out -covermode=count 2>&1
$exitCode = $LASTEXITCODE

Write-Host ""

# Check test results
if ($exitCode -eq 0) {
    Write-Host "✓ All tests passed!" -ForegroundColor Green
} else {
    Write-Host "✗ Some tests failed!" -ForegroundColor Red
    Write-Host $testResults
    exit 1
}

Write-Host ""

# Generate coverage report
Write-Host "Generating coverage report..." -ForegroundColor Yellow

if (Test-Path coverage.out) {
    go tool cover -func=coverage.out | Out-File coverage-report.txt
    Write-Host "✓ Coverage report saved to coverage-report.txt" -ForegroundColor Green
    
    # Display coverage summary
    Write-Host ""
    Write-Host "Coverage Summary:" -ForegroundColor Cyan
    go tool cover -func=coverage.out | Select-String "total:"
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    Write-Host "✓ HTML coverage report saved to coverage.html" -ForegroundColor Green
    Write-Host "  Open coverage.html in your browser to view detailed coverage"
} else {
    Write-Host "✗ No coverage data generated" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================"  -ForegroundColor Cyan
Write-Host "  Test run completed successfully!"  -ForegroundColor Green
Write-Host "========================================"  -ForegroundColor Cyan
