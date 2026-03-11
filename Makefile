BUILD_DIR := $(CURDIR)/build
PLUGIN_DIR := .
PLUGIN_SO  := ./loglinter.so

PLUGIN_FLAGS := -buildmode=plugin -tags plugin

.PHONY: build test check clean pre-commit all

# Build golangci-lint plugin
build:
	@echo "Building plugin..."
	CGO_ENABLED=1 go build $(PLUGIN_FLAGS) -o $(PLUGIN_SO) ./plugin/main.go

# Run unit tests
test:
	go test -v ./pkg/analyzer/...

# Build plugin and run golangci-lint on test data
check: build
	$(shell go env GOPATH)/bin/golangci-lint run --config .golangci.yml ./pkg/analyzer/testdata/src/a/... || echo "Expected exit code 1 because issues are found"

# Run pre-commit hooks
pre-commit:
	@pre-commit run --all-files

# Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f loglinter.so

# Full cycle: test + check
all: test check
