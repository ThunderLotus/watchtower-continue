# Testing Guide

This document provides comprehensive guidance for running and writing tests in the Watchtower project.

## Table of Contents

- [Quick Start](#quick-start)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Writing Tests](#writing-tests)
- [Test Conventions](#test-conventions)
- [CI/CD Integration](#cicd-integration)

## Quick Start

### Windows

```powershell
# Run all tests with coverage
.\scripts\run-tests.ps1

# Run specific package tests
go test ./pkg/container -v

# Run specific test
go test ./pkg/container -run TestContainerStateManagement -v
```

### Linux/Mac

```bash
# Run all tests with coverage
./scripts/run-tests.sh

# Run specific package tests
go test ./pkg/container -v

# Run specific test
go test ./pkg/container -run TestContainerStateManagement -v
```

## Running Tests

### Full Test Suite

Run all tests in the project:

```bash
go test ./... -v
```

### With Coverage

Run tests with coverage reporting:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Specific Package

Run tests for a specific package:

```bash
go test ./pkg/container -v
go test ./internal/actions -v
```

### Specific Test

Run a specific test function:

```bash
go test ./pkg/container -run TestContainerStateManagement -v
```

### Verbose Output

Show detailed test output:

```bash
go test ./... -v
```

### Race Detection

Enable race detector:

```bash
go test ./... -race
```

## Test Coverage

### Current Coverage

| Package | Coverage |
|---------|----------|
| internal/actions | 72.2% |
| internal/flags | 62.3% |
| internal/util | 83.7% |
| pkg/container | 56.8% |
| pkg/filters | 95.2% |
| pkg/notifications | 56.2% |
| pkg/registry | 81.2% |
| pkg/registry/auth | 31.5% |
| pkg/registry/digest | 58.0% |
| pkg/registry/helpers | 100.0% |
| pkg/registry/manifest | 84.6% |
| pkg/retry | 94.2% |

### Coverage Goals

- **Core modules** (container, registry): > 80%
- **Utility modules** (filters, retry): > 90%
- **Overall average**: > 75%

### Generating Coverage Reports

```bash
# Text coverage report
go tool cover -func=coverage.out > coverage-report.txt

# HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Percentage by package
go test ./... -cover
```

## Writing Tests

### Test File Naming

Test files should be named `*_test.go` and placed in the same directory as the code being tested:

```
pkg/
  container/
    container.go
    container_test.go
```

### Test Structure

Watchtower uses two testing frameworks:

1. **Ginkgo + Gomega** (BDD style, used in most packages)
2. **Standard testing package** (used in simple unit tests)

#### Ginkgo + Gomega Example

```go
package container

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

func TestContainer(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Container Suite")
}

var _ = Describe("Container", func() {
    var c *Container

    BeforeEach(func() {
        c = MockContainer()
    })

    Context("When initialized", func() {
        It("should have valid configuration", func() {
            Expect(c).NotTo(BeNil())
            Expect(c.ID()).NotTo(BeEmpty())
        })
    })

    Context("When updating", func() {
        It("should successfully pull new image", func() {
            err := c.PullImage()
            Expect(err).NotTo(HaveOccurred())
        })
    })
})
```

#### Standard Testing Example

```go
package container

import (
    "testing"
)

func TestContainerStaleState(t *testing.T) {
    c := &Container{}

    if c.IsStale() {
        t.Error("Initial state should be false")
    }

    c.SetStale(true)
    if !c.IsStale() {
        t.Error("After setting, state should be true")
    }
}
```

### Mock Objects

Use the mock objects in `pkg/container/mocks/`:

```go
import "github.com/containrrr/watchtower/pkg/container/mocks"

// Create a mock container
container := mocks.CreateMockContainer(
    "test-id",
    "/test",
    "nginx:latest",
    time.Now(),
)
```

### Table-Driven Tests

```go
func TestCalculateTimeout(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected time.Duration
    }{
        {"valid minutes", "5", 5 * time.Minute},
        {"valid hours", "1h", 1 * time.Hour},
        {"default", "", 1 * time.Minute},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateTimeout(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Test Conventions

### Naming Conventions

- Test files: `*_test.go`
- Test functions: `Test<FunctionName>` or `Test<FeatureName>`
- Ginkgo contexts: `When`, `Context`, `It`
- Variable names: Clear and descriptive

### Organization

1. **Setup**: Use `BeforeEach` to prepare test data
2. **Execution**: Test the actual functionality
3. **Assertion**: Verify the expected behavior
4. **Cleanup**: Use `AfterEach` if needed

### Error Handling

Always test error conditions:

```go
It("should return error for invalid input", func() {
    _, err := SomeFunction(invalidInput)
    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(Contain("invalid"))
})
```

### Concurrency Testing

For concurrent code, use race detector:

```bash
go test ./... -race
```

And write explicit concurrent tests:

```go
func TestConcurrentStateAccess(t *testing.T) {
    c := &Container{}
    var wg sync.WaitGroup

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c.SetStale(i%2 == 0)
        }()
    }

    wg.Wait()
    // Verify no race condition occurred
}
```

## CI/CD Integration

### GitHub Actions

Tests run automatically on:
- Pull requests
- Push to main branch
- Pull request merges

### Coverage Reporting

Coverage reports are automatically generated and can be viewed in the PR comments.

### Failing Tests

If tests fail:
1. Check the test logs in the CI/CD output
2. Run tests locally to reproduce
3. Fix the issue
4. Push the fix

## Best Practices

1. **Test One Thing**: Each test should verify a single behavior
2. **Use Descriptive Names**: Test names should clearly describe what is being tested
3. **Test Edge Cases**: Don't just test the happy path
4. **Keep Tests Independent**: Tests should not depend on each other
5. **Use Mocks**: Isolate the code under test from external dependencies
6. **Keep Tests Fast**: Avoid slow operations in tests
7. **Maintain Tests**: Update tests when code changes

## Troubleshooting

### Tests Fail Locally but Pass in CI

- Check Go version mismatch
- Ensure all dependencies are up to date: `go mod tidy`
- Check for platform-specific issues

### Coverage Not Updating

- Delete old coverage files: `rm coverage.out`
- Run tests with coverage: `go test ./... -coverprofile=coverage.out`

### Slow Tests

- Use `-short` flag to skip slow tests: `go test ./... -short`
- Profile tests to find bottlenecks: `go test -cpuprofile=cpu.prof ./...`

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Documentation](https://onsi.github.io/gomega/)
- [Go Profiling](https://golang.org/pkg/runtime/pprof/)
