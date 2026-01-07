.PHONY: clean build push test test-verbose run-local

# Build container image
build:
	@echo "Building container image with podman..."
	@podman build -t maas-metering-2:latest .
	@echo "Container image built: maas-metering-2:latest"

# Push container image to registry
push: build
	@echo "Tagging container image as quay.io/bryonbaker/maas-metering-2:latest"
	@podman tag maas-metering-2:latest quay.io/bryonbaker/maas-metering-2:latest
	@echo "Pushing container image to quay.io/bryonbaker/maas-metering-2:latest"
	@podman push quay.io/bryonbaker/maas-metering-2:latest
	@echo "Container image pushed to quay.io/bryonbaker/maas-metering-2:latest"

# Clean up container images
clean:
	@echo "Cleaning..."
	@podman rmi -f localhost/maas-metering-2:latest 2>/dev/null || true
	@podman rmi -f quay.io/bryonbaker/maas-metering-2:latest 2>/dev/null || true
	@echo "Clean complete"

# Run unit tests
test:
	@echo "Running unit tests..."
	@go test ./... -v
	@echo ""
	@echo "✓ All tests passed!"

test-verbose:
	@echo "Running unit tests with verbose output..."
	@go test ./... -v

# Run server locally for testing
# Default allows requests (ALLOW_VALUE=true)
# To test denying requests: make run-local ALLOW_VALUE=false
run-local:
	@echo "Running server locally..."
	@echo "Server will return allow=$(ALLOW_VALUE:-true)"
	@ALLOW_VALUE=$${ALLOW_VALUE:-true} \
	 PORT=$${PORT:-8080} \
	 go run cmd/server/main.go --port=$${PORT:-8080}

