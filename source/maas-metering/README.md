# MaaS Metering Service

Credit check service for Open Data Hub Model as a Service (MaaS) platform. This service integrates with Kuadrant AuthPolicy to provide intelligent, path-aware authorization for model inference requests.

---

## Overview

The MaaS Metering service is called by Authorino (via AuthPolicy metadata fetch) for **every request** to the MaaS gateway. It intelligently determines whether to apply credit checking based on the request path:

- **Non-inference paths** (API calls, OAuth, health checks) → **Always allowed** (no credit check)
- **Inference paths** (model completions) → **Credit check applied**

This approach provides a **single enforcement point** at the gateway level while preventing reconciliation loops and maintaining fast response times.

### Architecture

```
User Request
    ↓
Gateway (openshift-ingress/maas-default-gateway)
    ↓
Authorino (AuthPolicy enforcement)
    ├─ Authentication: TokenReview
    ├─ Metadata: Call maas-metering service with X-Original-Path
    │             ↓
    │   maas-metering:8080/allow
    │       ├─ Receives path, username, method
    │       ├─ Detects: inference vs non-inference
    │       ├─ Non-inference → return {"allow": true}
    │       └─ Inference → return {"allow": true/false} (credit check)
    │             ↓
    │   Returns 200 OK for ALL paths (no failures)
    │             ↓
    ├─ Authorization: OPA checks if allow==true
    └─ Response: Inject headers
    ↓
Request allowed/denied
```

---

## Path Detection Strategy

### OpenAI API Version Detection

The service detects inference requests by checking for `/v1/` as the **3rd segment** of the URL path. The `v1` refers to the **OpenAI API version** that MaaS supports.

**Pattern**: `/{namespace}/{model}/v1/{endpoint}`

Where:
- `{namespace}` = Kubernetes namespace (e.g., `llm`, `acme-inc-models`, `serverless-models`)
- `{model}` = Model name (e.g., `facebook-opt-125m-simulated`)
- `/v1/` = **OpenAI API version** (this is what we detect)
- `{endpoint}` = OpenAI endpoint (e.g., `completions`, `chat/completions`)

### Examples

**Inference Paths** (credit check applies):
```
/llm/facebook-opt-125m-simulated/v1/completions
/acme-inc-models/acme-inc-model-1/v1/chat/completions
/ethan-group-models/ethan-group-model-1/v1/completions
/serverless-models/my-model-name/v1/completions
```
→ All have `/v1/` at position 2 (3rd segment)

**Non-Inference Paths** (always allowed, no credit check):
```
/v1/models                  (API: list models)
/maas-api/v1/tokens         (API: get tokens)
/maas-api/v1/tiers          (API: tier management)
/oauth2                     (OAuth: authentication)
/health                     (Health check)
/                           (Root)
```

### Why This Works

- **Namespace-agnostic**: Works for any namespace without hardcoding names
- **Simple**: Just check if `parts[2] == "v1"`
- **Future-proof**: If MaaS adds v2 API, we can extend the logic
- **Fast**: No regex, no complex parsing

---

## API Endpoints

### `GET /allow`

Credit check endpoint called by Authorino.

**Request Headers** (passed by AuthPolicy):
```
X-Original-Path: /llm/model/v1/completions
X-Username: admin
X-Method: POST
```

**Response**:
```json
{
  "allow": true
}
```

**Status Codes:**
- `200 OK` - Always returns 200 (prevents reconciliation loops)

**Behavior:**
1. Extracts `X-Original-Path` header
2. Determines if path is inference or non-inference
3. Non-inference → returns `{"allow": true}` immediately (< 1ms)
4. Inference → checks `ALLOW_VALUE` env var → returns `{"allow": true/false}`
5. Unknown paths → returns `{"allow": false}` (fail-closed)

### `GET /health`

Health check endpoint for Kubernetes probes.

**Response:**
```json
{
  "status": "healthy"
}
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ALLOW_VALUE` | `"true"` | Credit check result for inference requests: `"true"` (allow) or `"false"` (deny) |
| `PORT` | `"8080"` | Server port |

**Note**: In Phase 1 (MVP), `ALLOW_VALUE` is a simple toggle. In production, this will be replaced with actual credit checking logic (database lookup, ConfigMap, etc.).

### Deployment Configuration

- **Namespace:** `maas-metering`
- **Service Name:** `maas-metering`
- **Service DNS:** `maas-metering.maas-metering.svc.cluster.local:8080`
- **Replicas:** 1
- **Resources:**
  - Requests: 50m CPU, 64Mi memory
  - Limits: 200m CPU, 256Mi memory

---

## Logging

All requests are logged to stdout in structured format:

```
[CREDIT-CHECK] path=/v1/models user=admin method=GET type=non-inference allow=true duration_ms=0
[CREDIT-CHECK] path=/llm/facebook-opt-125m-simulated/v1/completions user=alice method=POST type=inference allow=true duration_ms=1
[CREDIT-CHECK] path=/acme-inc-models/model-1/v1/chat/completions user=bob method=POST type=inference allow=false duration_ms=1
```

**Log Fields:**
- `path` - Original request path
- `user` - Authenticated username
- `method` - HTTP method (GET, POST, etc.)
- `type` - Request type (`non-inference`, `inference`, `unknown`)
- `allow` - Authorization decision (`true` or `false`)
- `duration_ms` - Processing time in milliseconds

---

## Building and Testing

### Prerequisites

- Go 1.24+
- Podman (for container builds)
- Access to `quay.io/bryonbaker` registry
- Kubernetes/OpenShift cluster with Kuadrant installed

### Local Development

#### Run Tests

```bash
cd /home/bryon/git-repos/maas-gitops/source/maas-metering

# Run all tests
go test -v ./internal/api/

# Expected output:
# PASS: TestIsNonInferencePath
# PASS: TestIsInferencePath
# PASS: TestEvaluateRequest
# PASS: TestEvaluateRequestDifferentUsers
# PASS: TestEvaluateRequestAllowToggle
# PASS: TestInferencePathVariations
```

#### Run Locally

```bash
# Allow mode (default)
make run-local

# In another terminal, test it:
curl -v http://localhost:8080/allow \
  -H "X-Original-Path: /v1/models" \
  -H "X-Username: testuser" \
  -H "X-Method: GET"

# Expected: {"allow":true} with type=non-inference log

curl -v http://localhost:8080/allow \
  -H "X-Original-Path: /llm/model/v1/completions" \
  -H "X-Username: testuser" \
  -H "X-Method: POST"

# Expected: {"allow":true} with type=inference log
```

#### Test Deny Mode

```bash
# Terminal 1: Run in deny mode
ALLOW_VALUE=false make run-local

# Terminal 2: Test non-inference (should still be allowed)
curl http://localhost:8080/allow \
  -H "X-Original-Path: /v1/models" \
  -H "X-Username: testuser"

# Expected: {"allow":true} (non-inference always allowed)

# Terminal 2: Test inference (should be denied)
curl http://localhost:8080/allow \
  -H "X-Original-Path: /llm/model/v1/completions" \
  -H "X-Username: testuser"

# Expected: {"allow":false} (inference denied when ALLOW_VALUE=false)
```

### Build and Push Container

```bash
# Build container image
make build

# Push to registry
make push

# Build + push in one command
make push
```

---

## Deployment

For detailed deployment instructions, see **[INSTALLATION.md](INSTALLATION.md)**.

### Quick Deploy

```bash
# 1. Build and push image
cd /home/bryon/git-repos/maas-gitops/source/maas-metering
make push

# 2. Deploy service
kubectl apply -k yaml/base/

# 3. Verify deployment
kubectl get pods -n maas-metering
kubectl logs -n maas-metering deployment/maas-metering -f

# 4. Apply updated gateway AuthPolicy
kubectl apply -f yaml/gateway-auth-policy-updated-gpt.yaml

# 5. Verify AuthPolicy is enforced
kubectl get authpolicy gateway-auth-policy -n openshift-ingress \
  -o jsonpath='{.status.conditions[?(@.type=="Enforced")].status}'
# Expected: True
```

---

## Performance

### Response Times

- **Non-inference requests**: < 1ms (immediate bypass)
- **Inference requests**: 1-2ms (simple boolean check)
- **Unknown paths**: < 1ms (immediate deny)

### Caching

The AuthPolicy caches responses for **60 seconds per user**:
- Reduces load on maas-metering service
- Improves gateway response time
- Cache key: `auth.identity.user.username`

---

## Security

### Current Implementation (Phase 1)

- ✅ Runs in dedicated namespace (`maas-metering`)
- ✅ Uses ServiceAccount with minimal RBAC
- ✅ ClusterIP service (not exposed outside cluster)
- ✅ No sensitive data processed
- ✅ Fail-closed for unknown paths
- ❌ No authentication required to call `/allow` (cluster-internal only)

### Future Enhancements

- Add service-to-service authentication (bearer token or mTLS)
- Implement request signing
- Add audit logging for all credit decisions
- Add rate limiting protection

---

## Troubleshooting

### Service not responding

```bash
# Check pod status
kubectl get pods -n maas-metering

# Check logs
kubectl logs -n maas-metering deployment/maas-metering -f

# Test service directly
kubectl run test --rm -it --image=curlimages/curl -- \
  curl http://maas-metering.maas-metering.svc.cluster.local:8080/health
```

### AuthPolicy not enforced

```bash
# Check AuthPolicy status
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml | grep -A 10 status:

# Check for ENFORCED=True
kubectl get authpolicy gateway-auth-policy -n openshift-ingress \
  -o jsonpath='{.status.conditions[?(@.type=="Enforced")].status}'

# If False, check Authorino logs
kubectl logs -n kuadrant-system deployment/authorino -f
```

### Requests not being logged

Check if headers are being passed correctly:

```bash
# Check AuthPolicy configuration
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml | grep -A 20 "X-Original-Path"

# Should show:
# X-Original-Path:
#   selector: context.request.http.path
```

---

## Roadmap

### ✅ Phase 1: MVP (Current)
- [x] Path-aware authorization
- [x] Service-side filtering logic
- [x] Gateway-level AuthPolicy integration
- [x] Comprehensive unit tests
- [x] Structured logging

### 📋 Phase 2: Real Credit Checking
- [ ] Replace `ALLOW_VALUE` with actual credit lookup
- [ ] Add ConfigMap or database integration
- [ ] Implement per-user credit tracking
- [ ] Add credit consumption API

### 📋 Phase 3: Production Hardening
- [ ] Add Prometheus metrics
- [ ] Implement distributed caching (Redis)
- [ ] Add request tracing (OpenTelemetry)
- [ ] Add detailed error responses
- [ ] Document operational procedures

---

## Files Structure

```
source/maas-metering/
├── cmd/
│   └── server/
│       └── main.go                        # Entry point
├── internal/
│   └── api/
│       ├── handlers.go                    # HTTP handlers with path detection
│       ├── handlers_test.go               # Comprehensive unit tests
│       └── router.go                      # Route setup
├── yaml/
│   ├── base/
│   │   ├── namespace.yaml                 # maas-metering namespace
│   │   ├── serviceaccount.yaml            # Service account
│   │   ├── deployment.yaml                # Deployment config
│   │   ├── service.yaml                   # ClusterIP service
│   │   └── kustomization.yaml             # Kustomize config
│   ├── gateway-auth-policy-updated-gpt.yaml   # Updated gateway AuthPolicy
│   └── gateway-auth-policy-backup-*.yaml      # Original policy backup
├── Containerfile                          # Multi-stage container build
├── Makefile                               # Build, push, test, run-local targets
├── go.mod                                 # Go module definition
├── go.sum                                 # Go module checksums
├── .gitignore                             # Git ignore rules
├── README.md                              # This file
├── INSTALLATION.md                        # Deployment guide
└── IMPLEMENTATION-PLAN.md                 # Technical implementation plan
```

---

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
