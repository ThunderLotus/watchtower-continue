# Test Conventions

This document defines the testing conventions and best practices for the Watchtower project.

## Table of Contents

- [Testing Frameworks](#testing-frameworks)
- [File Organization](#file-organization)
- [Test Structure](#test-structure)
- [Naming Conventions](#naming-conventions)
- [Assertion Guidelines](#assertion-guidelines)
- [Mock Usage](#mock-usage)
- [Error Testing](#error-testing)
- [Concurrency Testing](#concurrency-testing)
- [Documentation](#documentation)

## Testing Frameworks

### Ginkgo + Gomega (Preferred)

Use for complex test scenarios and BDD-style tests:

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
    // Test cases here
})
```

**When to use**:
- Integration tests
- Tests with multiple scenarios
- Tests requiring setup/teardown
- Tests that benefit from BDD style

### Standard Testing Package

Use for simple unit tests:

```go
package container

import (
    "testing"
)

func TestContainerStaleState(t *testing.T) {
    // Test code here
}
```

**When to use**:
- Simple unit tests
- Table-driven tests
- Performance benchmarks
- Tests with minimal setup

## File Organization

### Test File Location

Place test files in the same directory as the code being tested:

```
pkg/
  container/
    container.go
    container_test.go
    client.go
    client_test.go
    mocks/
      ApiServer.go
      container_ref.go
```

### Test File Naming

- Test files: `*_test.go`
- Mock files: `mocks/*.go`
- Test data: `mocks/data/*.json`

### Package Structure

Keep related tests together:

```go
// container_test.go
func TestContainerStateManagement(t *testing.T) { ... }
func TestContainerImageOperations(t *testing.T) { ... }
func TestContainerLifecycle(t *testing.T) { ... }
```

## Test Structure

### Ginkgo Test Structure

```go
var _ = Describe("Feature Name", func() {
    var (
        subject *Container
        err     error
    )

    BeforeEach(func() {
        // Setup code
        subject = NewContainer()
    })

    AfterEach(func() {
        // Cleanup code
        subject = nil
    })

    Context("When condition X is met", func() {
        It("should do Y", func() {
            // Test code
        })
    })

    When("condition Z is met", func() {
        It("should do W", func() {
            // Test code
        })
    })
})
```

### Standard Test Structure

```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"case 1", "input1", 1},
        {"case 2", "input2", 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Process(tt.input)
            if result != tt.expected {
                t.Errorf("got %d, want %d", result, tt.expected)
            }
        })
    }
}
```

## Naming Conventions

### Test Function Names

**Standard tests**: `Test<FeatureName><Scenario>`

```go
func TestContainerImageName(t *testing.T) { ... }
func TestContainerImageNameWithHash(t *testing.T) { ... }
func TestContainerStateTransitions(t *testing.T) { ... }
```

**Ginkgo contexts**: Descriptive phrases

```go
var _ = Describe("Container", func() {
    Context("When initialized", func() {
        It("should have valid ID", func() { ... })
    })

    Context("When using sha256 hash", func() {
        It("should resolve to tag", func() { ... })
    })
})
```

### Variable Names

Use clear, descriptive names:

```go
var (
    container    *Container
    imageName    string
    pullError    error
    mockServer   *ghttp.Server
)
```

### Test Data

Use descriptive names for test data:

```go
const (
    TestContainerID = "test-container-123"
    TestImageName   = "nginx:latest"
)
```

## Assertion Guidelines

### Gomega Assertions

Use appropriate matchers:

```go
// Equality
Expect(result).To(Equal(expected))

// Nil checks
Expect(result).To(BeNil())
Expect(result).NotTo(BeNil())

// Error checks
Expect(err).NotTo(HaveOccurred())
Expect(err).To(HaveOccurred())
Expect(err).To(MatchError("expected message"))

// Boolean
Expect(result).To(BeTrue())
Expect(result).To(BeFalse())

// Length
Expect(result).To(HaveLen(5))

// Containment
Expect(result).To(Contain("expected"))
Expect(result).To(HaveKey("key"))

// Numeric
Expect(result).To(BeNumerically(">", 0))
```

### Standard Testing Assertions

```go
if result != expected {
    t.Errorf("got %v, want %v", result, expected)
}

if err != nil {
    t.Errorf("unexpected error: %v", err)
}

if !reflect.DeepEqual(result, expected) {
    t.Errorf("got %v, want %v", result, expected)
}
```

### Assertion Best Practices

1. **One assertion per test**: Keep tests focused
2. **Use descriptive messages**: Explain what went wrong
3. **Test the contract**: Verify behavior, not implementation
4. **Use appropriate matchers**: Choose the right matcher for the assertion

## Mock Usage

### Creating Mocks

Use existing mock objects:

```go
import "github.com/containrrr/watchtower/pkg/container/mocks"

// Create a basic mock container
container := mocks.CreateMockContainer(
    "test-id",
    "/test",
    "nginx:latest",
    time.Now(),
)
```

### Mock Servers

Use ghttp for HTTP mocks:

```go
import "github.com/onsi/gomega/ghttp"

var server *ghttp.Server

BeforeEach(func() {
    server = ghttp.NewServer()
})

AfterEach(func() {
    server.Close()
})

It("should handle HTTP request", func() {
    server.AppendHandlers(
        ghttp.CombineHandlers(
            ghttp.VerifyRequest("GET", "/api"),
            ghttp.RespondWith(http.StatusOK, "response"),
        ),
    )

    // Test code
})
```

### Mock Configuration

Configure mocks with options:

```go
// With default options
container := mocks.MockContainer()

// With custom options
container := mocks.MockContainer(
    WithPortBindings(),
    WithImageName("nginx:1.25"),
    WithEnvironmentVariables([]string{"ENV=value"}),
)
```

## Error Testing

### Test Error Conditions

Always test error paths:

```go
It("should return error for invalid input", func() {
    _, err := SomeFunction(invalidInput)
    Expect(err).To(HaveOccurred())
    Expect(err).To(MatchError("invalid"))
})

It("should handle nil input gracefully", func() {
    err := SomeFunction(nil)
    Expect(err).To(HaveOccurred())
})
```

### Test Error Messages

Verify error messages are helpful:

```go
It("should return descriptive error message", func() {
    err := SomeFunction(invalidInput)
    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(Contain("expected"))
    Expect(err.Error()).NotTo(BeEmpty())
})
```

### Test Error Types

Check error types when appropriate:

```go
It("should return correct error type", func() {
    _, err := SomeFunction(invalidInput)
    Expect(err).To(HaveOccurred())
    var expectedErr *SomeError
    Expect(errors.As(err, &expectedErr)).To(BeTrue())
})
```

## Concurrency Testing

### Race Detection

Enable race detector in CI:

```bash
go test ./... -race
```

### Concurrent Test Cases

Test concurrent access:

```go
func TestConcurrentStateAccess(t *testing.T) {
    c := &Container{}
    var wg sync.WaitGroup

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            c.SetStale(idx%2 == 0)
        }(i)
    }

    wg.Wait()
    // Test passes if no race condition occurred
}
```

### Thread Safety Tests

Verify thread safety:

```go
It("should be safe for concurrent access", func() {
    var wg sync.WaitGroup
    errors := make(chan error, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            err := concurrentOperation()
            if err != nil {
                errors <- err
            }
        }()
    }

    wg.Wait()
    close(errors)

    for err := range errors {
        Expect(err).NotTo(HaveOccurred())
    }
})
```

## Documentation

### Test Documentation

Document complex tests:

```go
// TestContainerStateManagement verifies the three independent state fields
// (Stale, MarkedForUpdate, LinkedToRestarting) work correctly and
// independently. This test covers:
//   1. Initial state verification
//   2. State transitions
//   3. State independence
//   4. Concurrent access safety
func TestContainerStateManagement(t *testing.T) {
    // Test code
}
```

### Comment Complex Logic

Explain test logic:

```go
// We use a 100ms timeout here because the actual operation
// should complete within 10ms under normal conditions.
// This ensures the test fails quickly if there's a deadlock.
ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()
```

### Document Mock Behavior

Explain mock setup:

```go
// Set up mock server to simulate Docker API rate limiting
server.AppendHandlers(
    ghttp.CombineHandlers(
        ghttp.VerifyRequest("GET", "/images/json"),
        ghttp.RespondWith(http.StatusTooManyRequests, ""),
    ),
)
```

## Best Practices Summary

1. **Keep tests simple**: One test should verify one thing
2. **Use descriptive names**: Test names should be self-documenting
3. **Test edge cases**: Don't just test the happy path
4. **Use mocks appropriately**: Mock external dependencies
5. **Maintain tests**: Update tests when code changes
6. **Run tests frequently**: Catch issues early
7. **Keep tests fast**: Avoid slow operations in tests
8. **Test in isolation**: Tests should not depend on each other

## Common Mistakes to Avoid

1. **Testing implementation details**: Test behavior, not how it's implemented
2. **Skipping error cases**: Always test error paths
3. **Hard-coded values**: Use constants or test data helpers
4. **Ignoring cleanup**: Always clean up resources
5. **Too many assertions**: Keep tests focused
6. **Slow tests**: Optimize or skip slow tests with `-short`
7. **Flaky tests**: Ensure tests are deterministic
8. **Unnecessary mocks**: Don't mock when not needed

## Additional Resources

- [Effective Go: Testing](https://go.dev/doc/effective_go#testing)
- [Ginkgo Best Practices](https://onsi.github.io/ginkgo/#the-ginkgo-philosophy)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Subtests](https://go.dev/wiki/TableDrivenTests#subtests)
