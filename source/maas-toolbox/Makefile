.PHONY: clean build push test test-verbose test-coverage run-local swagger

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@$$(go env GOPATH)/bin/swag init -g cmd/server/main.go -o docs 2>&1 | grep -v "warning: failed to get package name" || true
	@echo "Swagger documentation generated in docs/"

# Build container image
build: swagger
	@echo "Building container image with podman..."
	@podman build -t maas-toolbox:latest .
	@echo "Container image built: maas-toolbox:latest"

# Push container image to registry
push: build
	@echo "Tagging container image as quay.io/bryonbaker/maas-toolbox:latest"
	@podman tag maas-toolbox:latest quay.io/bryonbaker/maas-toolbox:latest
	@echo "Pushing container image to quay.io/bryonbaker/maas-toolbox:latest"
	@podman push quay.io/bryonbaker/maas-toolbox:latest
	@echo "Container image pushed to quay.io/bryonbaker/maas-toolbox:latest"

# Clean up container images
clean:
	@echo "Cleaning..."
	@podman rmi -f localhost/maas-toolbox:latest 2>/dev/null || true
	@podman rmi -f quay.io/bryonbaker/maas-toolbox:latest 2>/dev/null || true
	@echo "Clean complete"

# Run unit tests (uses fake Kubernetes client, no cluster needed)
test:
	@echo "Running unit tests..."
	@go test ./... -v
	@echo ""
	@echo "✓ All tests passed!"

test-verbose:
	@echo "Running unit tests with verbose output..."
	@go test ./... -v

test-coverage:
	@echo "Running unit tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -1

# Run server locally for integration testing (requires kubeconfig)
# Uses your local kubeconfig to connect to a cluster
run-local:
	@echo "Running server locally (requires kubeconfig access to cluster)..."
	@echo "Server will use Kubernetes ConfigMap storage from your kubeconfig context"
	@NAMESPACE=$${NAMESPACE:-maas-api} \
	 CONFIGMAP_NAME=$${CONFIGMAP_NAME:-tier-to-group-mapping} \
	 PORT=$${PORT:-8080} \
	 go run cmd/server/main.go --port=$${PORT:-8080}



