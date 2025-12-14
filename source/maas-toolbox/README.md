# MaaS Toolbox API

A REST API service for simplifying configuration of Open Data Hub MaaS (ODH MaaS) via an API. The target use case for this toolbox is for Neo/private/sovereign cloud providers that are building their own Service request Portal in front of ODH MaaS and need a simplified management API. This service provides CRUD operations for managing subscription policies and tiers.

## Features

- **Create Tiers**: Add new tiers with name, description, level, and groups
- **List Tiers**: Retrieve all tiers or a specific tier by name
- **Update Tiers**: Modify tier description, level, and groups (name is immutable)
- **Delete Tiers**: Remove tiers from the configuration
- **Kubernetes ConfigMap Storage**: Stores tier configuration in Kubernetes ConfigMaps
- **OpenShift/Kubernetes Native**: Designed to run on OpenShift/Kubernetes clusters

## Backlog

1. Policy cofiguration

## Architecture

The service is built with a clean architecture:

- **Models**: Data structures for Tier and TierConfig
- **Storage**: Kubernetes ConfigMap-based persistence
- **Service Layer**: Business logic for tier management
- **API Layer**: REST API handlers using Gin framework

## Installation

### Prerequisites

- OpenShift/Kubernetes cluster access
- `oc` or `kubectl` CLI tool
- Container registry access (e.g., quay.io)
- Podman or Docker for building images

### Building the Container Image

```bash
make build
```

This builds the container image using podman.

### Pushing to Registry

```bash
make push
```

This tags and pushes the image to `quay.io/bryonbaker/maas-toolbox:latest`.

### Deployment

See the [yaml/README.md](yaml/README.md) for detailed deployment instructions.

```bash
oc apply -k yaml/
```

## API Endpoints

All endpoints are under `/api/v1/tiers`

### Get the API URL

After deployment, get the route URL:

```bash
ROUTE_URL=$(oc get route maas-toolbox -n maas-dev -o jsonpath='{.spec.host}')
echo "API URL: https://$ROUTE_URL"
```

### Create a Tier

```bash
curl -X POST https://$ROUTE_URL/api/v1/tiers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "free",
    "description": "Free tier for basic users",
    "level": 1,
    "groups": ["system:authenticated"]
  }'
```

### List All Tiers

```bash
curl https://$ROUTE_URL/api/v1/tiers
```

### Get a Specific Tier

```bash
curl https://$ROUTE_URL/api/v1/tiers/free
```

### Update a Tier

```bash
curl -X PUT https://$ROUTE_URL/api/v1/tiers/free \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated free tier description",
    "level": 2,
    "groups": ["system:authenticated", "free-users"]
  }'
```

Note: The `name` field cannot be changed. Only `description`, `level`, and `groups` can be updated.

### Delete a Tier

```bash
curl -X DELETE https://$ROUTE_URL/api/v1/tiers/free
```

### Add a Group to a Tier

```bash
curl -X POST https://$ROUTE_URL/api/v1/tiers/free/groups \
  -H "Content-Type: application/json" \
  -d '{"group": "new-group"}'
```

### Remove a Group from a Tier

```bash
curl -X DELETE https://$ROUTE_URL/api/v1/tiers/free/groups/system:authenticated
```

### Health Check

```bash
curl https://$ROUTE_URL/health
```

### Swagger Documentation

The API includes interactive Swagger documentation. Access it at:

```
https://$ROUTE_URL/swagger/index.html
```

The Swagger UI provides:
- Interactive API documentation
- Try-it-out functionality for all endpoints
- Request/response examples
- Schema definitions

To regenerate Swagger documentation after making changes to API annotations:

```bash
swag init -g cmd/server/main.go -o docs
```

## Configuration

The service stores tier configuration in a Kubernetes ConfigMap. The ConfigMap is automatically created in the `maas-api` namespace (configurable via `NAMESPACE` environment variable for non-disruptive testing) with the name `tier-to-group-mapping` (configurable via `CONFIGMAP_NAME` environment variable).

### Environment Variables

- `NAMESPACE`: Kubernetes namespace for the ConfigMap (default: `maas-api`)
- `CONFIGMAP_NAME`: Name of the ConfigMap (default: `tier-to-group-mapping`)
- `PORT`: Server port (default: `8080`)

### ConfigMap Format

Official documentation for Tier Configuration can be found [here](https://opendatahub-io.github.io/models-as-a-service/latest/configuration-and-management/tier-overview/).

The ConfigMap stores tiers in the following format:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tier-to-group-mapping
  namespace: maas-api
data:
  tiers: |
    - name: free
      description: Free tier for basic users
      level: 1
      groups:
      - system:authenticated
    - name: premium
      description: Premium tier
      level: 10
      groups:
      - premium-users
```

## Business Rules

1. **Tier Name**: Set at creation time and cannot be changed
2. **Tier Uniqueness**: Tier names must be unique
3. **Required Fields**: Name and description are required
4. **Level**: Must be a non-negative integer
5. **Groups**: Array of Kubernetes group names

## Error Responses

All errors follow this format:

```json
{
  "error": "error message"
}
```

HTTP Status Codes:
- `200 OK`: Success
- `201 Created`: Tier created successfully
- `204 No Content`: Tier deleted successfully
- `400 Bad Request`: Validation error or invalid request
- `404 Not Found`: Tier not found
- `409 Conflict`: Tier already exists
- `500 Internal Server Error`: Server error

## Future Enhancements

- Authentication and authorization
- Rate limiting
- Enhanced logging and metrics
- Multi-namespace support
- Subscription policy configuration. [Documentation](https://opendatahub-io.github.io/models-as-a-service/latest/configuration-and-management/tier-configuration/)

## Development

### Project Structure

```
maas-toolbox/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go      # HTTP request handlers
│   │   └── router.go        # Route configuration
│   ├── models/
│   │   ├── tier.go          # Data models
│   │   └── errors.go        # Error definitions
│   ├── storage/
│   │   ├── k8s.go           # Kubernetes ConfigMap storage
│   │   └── k8s_client.go    # Kubernetes client initialization
│   └── service/
│       └── tier.go          # Business logic
├── yaml/                    # Kubernetes/OpenShift deployment files
├── tests/
│   └── test-api.sh          # API integration test script
├── go.mod
├── go.sum
├── README.md
└── Makefile
```

### Building

Build the container image:

```bash
make build
```

### Running Tests

#### Unit Tests

```bash
make test
```

Or with coverage:

```bash
make test-coverage
```

#### API Integration Tests

A comprehensive test script is provided to test all API endpoints against a deployed cluster:

```bash
# Test against deployed cluster
./tests/test-api.sh https://maas-toolbox-maas-dev.apps.$BASE_DOMAIN
```

The test script will:
- Test all CRUD operations (Create, Read, Update, Delete)
- Test group management (Add/Remove groups)
- Test error cases (duplicate tiers, invalid data, not found, etc.)
- Test edge cases (empty groups, validation, etc.)
- Display colored output with pass/fail status
- Provide a summary at the end

The script uses tier names `acme-inc-1`, `acme-inc-2`, and `acme-inc-3` for testing.

## License

Copyright 2025 Bryon Baker

Licensed under the Apache License, Version 2.0 (the "License");  
you may not use this file except in compliance with the License.  
You may obtain a copy of the License at  

    http://www.apache.org/licenses/LICENSE-2.0  

Unless required by applicable law or agreed to in writing, software  
distributed under the License is distributed on an "AS IS" BASIS,  
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  
See the License for the specific language governing permissions and  
limitations under the License.  

