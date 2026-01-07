# MaaS Metering Service v2 (Simplified Architecture)

**Version:** 2.0  
**Status:** Phase 1 - MVP Credit Check

## Overview

The MaaS Metering Service v2 is a **pure credit-checking service** for the Model-as-a-Service (MaaS) platform. This version simplifies the architecture by removing path filtering logic from the Go application and relying on Kuadrant's route-level policy precedence.

### Key Simplification

**v1 Architecture (maas-metering):**
- Go service performed path filtering (inference vs non-inference)
- Complex logic with `isInferencePath()`, `isNonInferencePath()`, etc.
- ~180 lines of code

**v2 Architecture (maas-metering-2):**
- Go service is a **pure credit check** (single responsibility)
- Route-level AuthPolicies filter non-inference paths
- ~60 lines of code (67% reduction)

---

## Architecture

### Request Flow

```
HTTP Request
    ↓
Route-Level Policy Precedence
    ├─ /v1/models → maas-api-route → maas-api-auth-policy (Route-level)
    ├─ /maas-api/* → maas-api-route → maas-api-auth-policy (Route-level)
    ├─ /oauth/* → oauth-route → oauth-auth-policy (Route-level)
    └─ /{namespace}/{model}/v1/* → Gateway-level policy (Inference)
        ↓
1. Authentication (TokenReview)
    ↓
2. Metadata Fetch: Call maas-metering-2 service
    │   GET /allow
    │   Headers: X-Username (for logging)
    ↓
3. Go Service: Pure Credit Check
    └─ return {"allow": ALLOW_VALUE}
    ↓
4. OPA Authorization: Check result
    │   allow { metadata["credit-allow"].allow == true }
    ↓
5. tier-access (RBAC check)
    ↓
6. Allow/Deny
```

### Why This Works

**Route-Level Policies Take Precedence:**
- Non-inference paths (`/v1/models`, `/maas-api/*`, `/oauth/*`) are handled by route-level policies
- These policies execute **before** the gateway-level policy
- The gateway-level policy **only sees inference requests**
- Therefore, the Go service **only processes inference requests**

**No Path Filtering Needed:**
- Every request reaching the gateway policy is already an inference request
- No need for `isInferencePath()` logic in Go
- No need for `when` clauses in AuthPolicy
- Pure credit check responsibility

---

## Components

### 1. Go Service (`maas-metering-2`)

**Endpoints:**
- `GET /allow` - Credit check endpoint (called by Authorino)
- `GET /health` - Health check endpoint

**Environment Variables:**
- `ALLOW_VALUE` - Phase 1 credit decision (`true`/`false`)
- `PORT` - HTTP server port (default: `8080`)

**Response Format:**
```json
{
  "allow": true
}
```

**Logging:**
```
[CREDIT-CHECK] path=/llm/model/v1/completions user=system:serviceaccount:... method=POST allow=true duration_ms=0
```

### 2. AuthPolicy (`gateway-auth-policy-updated.yaml`)

**Metadata Fetch:**
```yaml
metadata:
  credit-allow:
    http:
      url: "http://maas-metering-2.maas-metering-2.svc.cluster.local:8080/allow"
      headers:
        X-Username:
          selector: auth.identity.user.username
```

**Authorization:**
```yaml
authorization:
  credit-access:
    opa:
      rego: |
        allow {
          input.auth.metadata["credit-allow"].allow == true
        }
```

---

## Deployment

### Prerequisites

- OpenShift 4.x or Kubernetes 1.24+
- Kuadrant installed
- Podman or Docker
- Go 1.23+

### Quick Start

```bash
# 1. Build container image
cd source/maas-metering-2
make build

# 2. Push to registry
make push

# 3. Deploy to cluster
kubectl apply -k yaml/base/

# 4. Verify deployment
kubectl get pods -n maas-metering-2
kubectl logs -n maas-metering-2 deployment/maas-metering-2

# 5. Apply gateway AuthPolicy
kubectl apply -f yaml/gateway-auth-policy-updated.yaml
```

### Configuration

**Allow all inference requests:**
```bash
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=true -n maas-metering-2
```

**Deny all inference requests:**
```bash
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=false -n maas-metering-2
```

---

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose
```

### Integration Tests

**Test 1: Non-inference paths (should NOT call service)**
```bash
# These are handled by route-level policies
curl -k https://maas.apps.your-domain.com/v1/models \
  -H "Authorization: Bearer $TOKEN"

# Check logs - should be empty (service not called)
kubectl logs -n maas-metering-2 deployment/maas-metering-2
```

**Test 2: Inference paths (should call service)**
```bash
# This is handled by gateway policy
curl -k https://maas.apps.your-domain.com/llm/model-name/v1/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"facebook/opt-125m","prompt":"test","max_tokens":10}'

# Check logs - should show credit check
kubectl logs -n maas-metering-2 deployment/maas-metering-2
```

**Test 3: Credit check deny**
```bash
# Set ALLOW_VALUE=false
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=false -n maas-metering-2

# Wait for rollout
kubectl rollout status deployment/maas-metering-2 -n maas-metering-2

# Test inference (should be denied)
curl -k https://maas.apps.your-domain.com/llm/model-name/v1/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"facebook/opt-125m","prompt":"test","max_tokens":10}'

# Expected response:
# {
#   "error": "access_denied",
#   "message": "Access denied: You do not have sufficient credits or permissions...",
#   "code": 403
# }
```

---

## Phase 1 vs Future Phases

### Phase 1 (Current - MVP)
- ✅ Pure credit check service
- ✅ `ALLOW_VALUE` environment variable (boolean)
- ✅ Logging to stdout
- ✅ Simple allow/deny decision
- ✅ No path filtering logic

### Phase 2+ (Future)
- 🔮 Actual credit balance lookup (database/ConfigMap/API)
- 🔮 Per-user credit tracking
- 🔮 Credit consumption recording
- 🔮 Credit balance response in metadata
- 🔮 Graceful degradation strategies

---

## Troubleshooting

### Service not called for inference requests

**Check gateway policy target:**
```bash
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml
```

Ensure `targetRef` points to the correct gateway:
```yaml
targetRef:
  kind: Gateway
  name: maas-default-gateway
```

### Service returning wrong decision

**Check ALLOW_VALUE:**
```bash
kubectl get deployment maas-metering-2 -n maas-metering-2 -o jsonpath='{.spec.template.spec.containers[0].env}'
```

### AuthPolicy not enforced

**Check status:**
```bash
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o jsonpath='{.status.conditions}'
```

Look for `Enforced: True`.



