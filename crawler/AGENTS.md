# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## ! Important !

**Never** mention agent's name, model and vendor in commit messages or generated code or any other materials

## Project Structure
- Uses standard Go project layout with `internal/` for private packages
- `tests/e2e/` directory is for end-to-end testing
- `migrations/` directory is for database migration files

## Commands

### Build
- `go build ./...` - Build all packages
- `go build -o bin/app ./cmd/app` - Build specific application
- `go install ./cmd/app` - Install application to $GOPATH/bin
- `go mod tidy` - Clean up module dependencies
- `go mod download` - Download module dependencies

### Test
- `go test ./...` - Run all tests
- `go test -v ./...` - Run all tests with verbose output
- `go test -race ./...` - Run tests with race detector
- `go test -cover ./...` - Run tests with coverage
- `go test -coverprofile=coverage.out ./...` - Generate coverage profile
- `go test -run TestName ./internal/pkg` - Run specific test
- `go test -run "^TestName$" ./path/to/package` - Run exact test match
- `go test -run "TestPrefix*" ./...` - Run tests matching pattern
- `go test ./internal/pkg -run TestSpecific` - Run specific test in package
- `go test -bench=. ./...` - Run benchmarks
- `go test -timeout 30s ./...` - Set test timeout

### Lint/Format
- `go fmt ./...` - Format all Go code
- `go vet ./...` - Run go vet for static analysis
- `golangci-lint run ./...` - Run golangci-lint (if installed)
- `go mod verify` - Verify module dependencies

## Code Style Guidelines

### Common Practices
- Write modern, clean and readable code
- Follow style and conventions already used in the project
- Avoid redundant comments, but document complex logic and APIs

### Imports
- Group imports: standard library, third-party, local packages
- Use blank lines between import groups
- Alphabetical order within groups
- Remove unused imports

 ```go
import (
	"errors"
	"fmt"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"{{project-name}}/internal/config"
	"{{project-name}}/internal/shared"
)
```

### Formatting
- Use `go fmt` standards
- 4-space indentation (tabs converted to spaces)
- 100-120 character line length limit
- Empty lines between logical code blocks

### Naming Conventions
- **Packages**: lowercase, single word, descriptive
- **Variables**: camelCase, descriptive names
- **Constants**: PascalCase or UPPER_CASE
- **Functions**: PascalCase for exported, camelCase for internal
- **Interfaces**: Use -er suffix (Reader, Writer, Handler)
- **Test files**: _test.go suffix
- **Test functions**: TestXxx with PascalCase

### Error Handling
- Use `fmt.Errorf()` with `%w` verb for wrapping errors
- Use `errors.Is()` to check for specific errors (including wrapped ones)
- Use `errors.As()` to check and extract error types
- Use `errors.New()` for simple constant errors
- Always check errors, don't ignore them
- Return errors from functions rather than logging them
- Use meaningful error messages
- Define sentinel errors with `var ErrName = errors.New("...")` for comparison

```go
var ErrNotFound = errors.New("resource not found")

func processFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()
	
	data, err := readFileData(file)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	
	return nil
}
```

### Types and Interfaces
- Use type-safe interfaces
- Prefer small, focused interfaces
- Use struct tags for marshaling/unmarshaling
- Document exported types and methods

### Logging
- Use structured logging
- Log at appropriate levels (debug, info, warn, error)
- Include relevant context in log messages
- Avoid logging sensitive information

### Testing Conventions
- Table-driven tests for multiple test cases
- Use Given-When-Then formula:
    - Given: Setup/preconditions
    - When: Action being tested
    - Then: Expected outcome
- Use the following naming conventions for tests:
    - func TestFff (t *testing.T) for functions
    - func TestTtt_Mmm (t *testing.T) for type methods
    - add `_WhenXxx_ThenYxx` suffix to describe action and expected outcome
- Use test helpers for common setup/teardown
- Test both success and error cases
- Use github.com/stretchr/testify/assert for assertions
- Use github.com/stretchr/testify/require for assertions that should stop test execution
- Mock external dependencies with testify/mock

```go
func TestCalculateTotal_WhenValidNumbersThenExpectedTotal(t *testing.T) {
	// Given
	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{"empty slice", []int{}, 0},
		{"single item", []int{5}, 5},
		{"multiple items", []int{1, 2, 3}, 6},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := CalculateTotal(tt.input)
			
			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

### Go Common Practices
- Write idiomatic, modern code, according to the project's Go version
- Follow `Effective Go` principles
- Follow `Uber Go Style Guilde` recommendations
- Use context.Context for request-scoped values and cancellation
- Use sync package for concurrency control
- Implement graceful shutdown patterns
- Use environment variables for configuration
- Write meaningful Godoc comments

## Package Organization
- `cmd/` - Application entry points (main packages)
- `internal/` - Private application code
- `internal/config/` - Configuration management
- `internal/sample_feature/` - Example feature implementation
- `internal/shared/` - Shared utilities and common code
- `tests/e2e/` - End-to-end tests
- `migrations/` - Database migration files

## Development Workflow
1. Run `go mod tidy` after adding dependencies
2. Run `go fmt` before committing
3. Run `go vet` and `go test` before pushing
4. Use meaningful commit messages
5. Follow semantic versioning for releases

## Verification
After making changes, always run lint and typecheck commands (e.g., `go vet`) to ensure code correctness.
Do not run integration/e2e tests until user explicitly ask it.

## Security Best Practices
- Never commit secrets or credentials
- Use environment variables for sensitive configuration
- Validate all input data
- Use prepared statements for database queries
- Implement proper authentication and authorization

This file will be updated as the project evolves and specific patterns are established.
