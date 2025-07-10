# Testing Guide

This document describes how to run tests for the GasDB project.

## Overview

The project includes several types of tests:

- **Unit Tests**: Test individual functions and components in isolation
- **Integration Tests**: Test the API against the real external service
- **Benchmark Tests**: Measure performance of API operations

## Running Tests

### All Tests

Run all tests in the project:

```bash
go test ./...
```

### Unit Tests Only

Run tests excluding integration tests:

```bash
go test -short ./...
```

### Integration Tests

Run API integration tests that make real HTTP requests:

```bash
cd pkg/api
go test -v -run TestFuelPriceAPI
```

**Note**: Integration tests make real API calls and may take longer to complete. They require an internet connection.

### Benchmark Tests

Run performance benchmarks:

```bash
cd pkg/api
go test -bench=. -benchmem
```

### Coverage

Generate test coverage reports:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Structure

### API Tests (`pkg/api/api_test.go`)

- `TestFuelPriceAPI_FetchPrices`: Tests fetching current fuel prices
- `TestFuelPriceAPI_FetchPricesForDate`: Tests fetching historical prices
- `TestFuelPriceAPI_NearbyPrices`: Tests location-based filtering
- `TestFuelPriceAPI_InvalidCoordinates`: Tests edge cases
- `TestParseLatLong`: Tests coordinate parsing utility

### Benchmark Tests

- `BenchmarkFuelPriceAPI_FetchPrices`: Measures API fetch performance
- `BenchmarkFuelPriceAPI_NearbyPrices`: Measures filtering performance

## CI/CD

The project uses GitHub Actions for continuous integration:

- **Test Workflow** (`.github/workflows/test.yml`): Runs basic tests
- **CI Workflow** (`.github/workflows/ci.yml`): Comprehensive pipeline including:
  - Linting with golangci-lint
  - Security scanning with gosec
  - Unit and integration tests
  - Build verification
  - Benchmark tracking

## Test Configuration

### Environment Variables

No special environment variables are required for testing.

### Test Data

Integration tests use:
- Real API endpoints (external dependency)
- Madrid coordinates (40.4168, -3.7038) for location tests
- Various distance parameters for nearby search tests

## Writing New Tests

### Test Naming

Follow Go testing conventions:
- Test functions: `TestFunctionName`
- Benchmark functions: `BenchmarkFunctionName`
- Example functions: `ExampleFunctionName`

### Integration Test Guidelines

When adding new integration tests:

1. Use descriptive test names
2. Include proper error handling
3. Test both success and failure cases
4. Consider API rate limits
5. Use reasonable timeouts
6. Clean up any resources if needed

### Example Test Structure

```go
func TestNewFeature(t *testing.T) {
    // Arrange
    api := NewFuelPriceAPI()
    
    // Act
    result, err := api.NewFeature()
    
    // Assert
    if err != nil {
        t.Fatalf("NewFeature() failed: %v", err)
    }
    
    if result == nil {
        t.Fatal("NewFeature() returned nil")
    }
    
    // Additional assertions...
}
```

## Troubleshooting

### Common Issues

1. **Network timeouts**: Integration tests may fail due to network issues
   - Solution: Retry the test or check internet connection

2. **API rate limiting**: Too many requests to external API
   - Solution: Add delays between tests or run tests less frequently

3. **Build failures**: Missing dependencies or module issues
   - Solution: Run `go mod download` and `go mod tidy`

### Debug Mode

Run tests with verbose output:

```bash
go test -v ./...
```

Run specific test:

```bash
go test -v -run TestSpecificFunction ./pkg/api
```

## Performance Guidelines

- Integration tests should complete within 60 seconds
- Benchmark tests should run for at least 1 second
- API calls should be cached when possible to reduce external dependencies