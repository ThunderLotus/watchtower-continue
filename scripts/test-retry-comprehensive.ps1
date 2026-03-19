# Comprehensive Retry Logic Test Script
# This script runs comprehensive tests for the retry logic implementation

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Watchtower Retry Logic Comprehensive Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$ErrorActionPreference = "Continue"
$testResults = @()
$failedTests = 0
$passedTests = 0

# Function to run test and record result
function Run-Test {
    param (
        [string]$TestName,
        [string]$Command
    )
    
    Write-Host "Running: $TestName..." -ForegroundColor Yellow
    $startTime = Get-Date
    
    try {
        $output = Invoke-Expression $Command 2>&1
        $exitCode = $LASTEXITCODE
        
        $duration = ((Get-Date) - $startTime).TotalSeconds
        
        if ($exitCode -eq 0) {
            Write-Host "✓ PASSED" -ForegroundColor Green
            $passedTests++
            $testResults += [PSCustomObject]@{
                Test = $TestName
                Status = "PASSED"
                Duration = "$($duration.ToString('0.00'))s"
                ExitCode = $exitCode
            }
        } else {
            Write-Host "✗ FAILED (Exit Code: $exitCode)" -ForegroundColor Red
            $failedTests++
            $testResults += [PSCustomObject]@{
                Test = $TestName
                Status = "FAILED"
                Duration = "$($duration.ToString('0.00'))s"
                ExitCode = $exitCode
            }
        }
    } catch {
        Write-Host "✗ ERROR: $($_.Exception.Message)" -ForegroundColor Red
        $failedTests++
        $testResults += [PSCustomObject]@{
            Test = $TestName
            Status = "ERROR"
            Duration = "0.00s"
            ExitCode = -1
        }
    }
    
    Write-Host ""
}

# Test 1: Retry module tests with race detection
Run-Test "Retry Module Tests (Race Detection)" "go test ./pkg/retry -v -race"

# Test 2: Container module tests with race detection
Run-Test "Container Module Tests (Race Detection)" "go test ./pkg/container -v -race"

# Test 3: Registry module tests with race detection
Run-Test "Registry Module Tests (Race Detection)" "go test ./pkg/registry -v -race"

# Test 4: API module tests with race detection
Run-Test "API Module Tests (Race Detection)" "go test ./pkg/api -v -race"

# Test 5: Notifications module tests with race detection
Run-Test "Notifications Module Tests (Race Detection)" "go test ./pkg/notifications -v -race"

# Test 6: Internal modules tests with race detection
Run-Test "Internal Modules Tests (Race Detection)" "go test ./internal/... -v -race"

# Test 7: Retry configuration validation
Run-Test "Retry Configuration Tests" "go test ./pkg/retry -run TestDefaultConfig -v"

# Test 8: Retry error classification
Run-Test "Retry Error Classification Tests" "go test ./pkg/retry -run TestIsRetryableError -v"

# Test 9: Retry statistics
Run-Test "Retry Statistics Tests" "go test ./pkg/retry -run TestStats -v"

# Test 10: Retry thread safety
Run-Test "Retry Thread Safety Tests" "go test ./pkg/retry -run TestStatsThreadSafety -v"

# Test 11: Retry concurrency
Run-Test "Retry Concurrency Tests" "go test ./pkg/retry -run TestWithRetryConcurrency -v"

# Test 12: Retry race condition detector
Run-Test "Retry Race Condition Detector Tests" "go test ./pkg/retry -run TestStatsRaceConditionDetector -v"

# Test 13: Benchmark tests
Write-Host "Running: Benchmark Tests..." -ForegroundColor Yellow
$startTime = Get-Date
try {
    $benchOutput = go test ./pkg/retry -bench=. -benchmem 2>&1
    $exitCode = $LASTEXITCODE
    $duration = ((Get-Date) - $startTime).TotalSeconds
    
    if ($exitCode -eq 0) {
        Write-Host "✓ PASSED" -ForegroundColor Green
        $passedTests++
        $testResults += [PSCustomObject]@{
            Test = "Benchmark Tests"
            Status = "PASSED"
            Duration = "$($duration.ToString('0.00'))s"
            ExitCode = $exitCode
        }
        
        # Save benchmark results to file
        $benchOutput | Out-File -FilePath "benchmark_results.txt" -Encoding UTF8
        Write-Host "  Benchmark results saved to benchmark_results.txt" -ForegroundColor Gray
    } else {
        Write-Host "✗ FAILED (Exit Code: $exitCode)" -ForegroundColor Red
        $failedTests++
        $testResults += [PSCustomObject]@{
            Test = "Benchmark Tests"
            Status = "FAILED"
            Duration = "$($duration.ToString('0.00'))s"
            ExitCode = $exitCode
        }
    }
} catch {
    Write-Host "✗ ERROR: $($_.Exception.Message)" -ForegroundColor Red
    $failedTests++
}

Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Total Tests: $($testResults.Count)" -ForegroundColor White
Write-Host "Passed: $passedTests" -ForegroundColor Green
Write-Host "Failed: $failedTests" -ForegroundColor Red
$successRate = if ($testResults.Count -gt 0) { [math]::Round(($passedTests / $testResults.Count) * 100, 2) } else { 0 }
Write-Host "Success Rate: $successRate%" -ForegroundColor $(if ($successRate -ge 80) { "Green" } elseif ($successRate -ge 50) { "Yellow" } else { "Red" })
Write-Host ""

# Detailed results
Write-Host "Detailed Results:" -ForegroundColor Cyan
$testResults | Format-Table -AutoSize

# Save results to file
$testResults | Export-Csv -Path "test_results.csv" -NoTypeInformation -Encoding UTF8
Write-Host "Test results saved to test_results.csv" -ForegroundColor Gray

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Complete" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

exit $failedTests