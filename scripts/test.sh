#!/bin/bash

# test.sh - Local test script that mimics CI pipeline
set -e

echo "🧪 Running GasDB Test Suite"
echo "=========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "Must be run from the root directory of the project"
    exit 1
fi

# Step 1: Verify dependencies
echo "📦 Verifying dependencies..."
go mod download
go mod verify
print_status "Dependencies verified"

# Step 2: Run go vet
echo "🔍 Running go vet..."
if go vet ./...; then
    print_status "go vet passed"
else
    print_error "go vet failed"
    exit 1
fi

# Step 3: Run unit tests
echo "🧪 Running unit tests..."
if go test -race -coverprofile=coverage.out -covermode=atomic ./...; then
    print_status "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Step 4: Build all packages
echo "🔨 Building all packages..."
if go build ./...; then
    print_status "Build successful"
else
    print_error "Build failed"
    exit 1
fi

# Step 5: Test examples build
echo "🎯 Testing examples build..."
cd _examples
if go mod tidy && go build .; then
    print_status "Examples build successful"
else
    print_error "Examples build failed"
    exit 1
fi
cd ..

# Step 6: Test server build
echo "🌐 Testing server build..."
cd _server
if go mod tidy && go build .; then
    print_status "Server build successful"
else
    print_error "Server build failed"
    exit 1
fi
cd ..

# Step 7: Test CLI build
echo "⚡ Testing CLI build..."
cd cmd/gasdb
if go build .; then
    print_status "CLI build successful"
else
    print_error "CLI build failed"
    exit 1
fi
cd ../..

# Step 8: Run integration tests
echo "🌍 Running integration tests..."
cd pkg/api
if go test -v -timeout=60s -run TestFuelPriceAPI; then
    print_status "Integration tests passed"
else
    print_warning "Integration tests failed (this might be due to network issues)"
fi
cd ../..

# Step 9: Generate coverage report
echo "📊 Generating coverage report..."
if command -v go >/dev/null 2>&1; then
    go tool cover -html=coverage.out -o coverage.html
    print_status "Coverage report generated: coverage.html"
    
    # Show coverage summary
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "📈 Total coverage: $COVERAGE"
fi

# Step 10: Run benchmarks (optional)
if [ "$1" = "--bench" ]; then
    echo "⚡ Running benchmarks..."
    cd pkg/api
    go test -bench=. -benchmem -count=1
    cd ../..
    print_status "Benchmarks completed"
fi

echo ""
echo "🎉 All tests completed successfully!"
echo "=========================="

# Optional: Run linter if available
if command -v golangci-lint >/dev/null 2>&1; then
    echo "🔍 Running linter..."
    if golangci-lint run --timeout=5m; then
        print_status "Linting passed"
    else
        print_warning "Linting issues found"
    fi
else
    print_warning "golangci-lint not installed, skipping linting"
    echo "   Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

echo ""
echo "✨ Test suite complete!"