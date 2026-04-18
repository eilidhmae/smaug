.PHONY: build test test-v test-race clean run vet lint fmt check install-hooks

BINARY  := smaug-go
PORT    := 4000
DATADIR := ../db

# Build the server binary
build:
	go build -o $(BINARY) ./cmd/smaug/

# Run all tests
test:
	go test -count=1 -timeout 30s ./...

# Run all tests with verbose output
test-v:
	go test -v -count=1 -timeout 30s ./...

# Run tests with race detector
test-race:
	go test -race -count=1 -timeout 60s ./...

# Run a specific package's tests (usage: make test-pkg PKG=util)
test-pkg:
	go test -v -count=1 -timeout 30s ./internal/$(PKG)/

# Run a specific test by name (usage: make test-run RUN=TestBitVector PKG=types)
test-run:
	go test -v -count=1 -timeout 30s ./internal/$(PKG)/ -run $(RUN)

# Run the server
run: build
	./$(BINARY) -port $(PORT) -data $(DATADIR)

# Go vet (static analysis)
vet:
	go vet ./...

# Format all Go files
fmt:
	gofmt -w .

# Check formatting without writing (CI-friendly)
fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

# All checks: vet + test + format check
check: vet fmt-check test

# Install the git pre-commit hook for this clone
install-hooks:
	./hooks/install.sh

# Remove build artifacts
clean:
	rm -f $(BINARY)
	go clean -testcache
