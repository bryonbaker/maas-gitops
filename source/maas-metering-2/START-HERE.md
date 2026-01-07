# MaaS Metering v2 - Start Here

**Welcome to MaaS Metering v2!** This is a simplified credit-checking service for the Model-as-a-Service platform.

---

## Quick Navigation

### 📖 **New to this project?**
Start with the [README.md](README.md) to understand:
- What this service does
- How it works (architecture)
- Why v2 is simpler than v1
- Key design decisions

### 🚀 **Ready to deploy?**
Follow the [INSTALLATION.md](INSTALLATION.md) for step-by-step instructions:
- Building the container image
- Deploying to Kubernetes/OpenShift
- Configuring the AuthPolicy
- Testing credit checks

### 💡 **Key Differences from v1**

| v1 (maas-metering) | v2 (maas-metering-2) |
|-------------------|---------------------|
| ~180 lines of code | ~60 lines (67% less) |
| Path filtering in Go | No filtering needed |
| Complex logic | Single responsibility |
| Multiple test scenarios | Simple credit tests |

**Why the change?**  
Discovered that route-level AuthPolicies handle non-inference paths, so the gateway policy only sees inference requests. This eliminates the need for path filtering logic in the Go service!

---

## What's New in v2?

### ✨ Simplified Architecture

```
Route-Level Precedence Filters Requests:
  ├─ /v1/models → maas-api-route (not our concern)
  ├─ /maas-api/* → maas-api-route (not our concern)
  └─ /{namespace}/{model}/v1/* → Gateway policy (inference only)
      ↓
  maas-metering-2 service (pure credit check)
      ↓
  {"allow": true/false}
```

### 🎯 Single Responsibility

The Go service now does ONE thing: **check if the user has credits**.

- No path parsing
- No request type classification
- No routing decisions
- Pure credit check logic

### 📊 What's Kept

- ✅ All headers still passed (for logging)
- ✅ Same observability (logs show path, user, method)
- ✅ Same AuthPolicy structure
- ✅ Same error messages
- ✅ Same deployment process

### ❌ What's Removed

- Path filtering functions (`isInferencePath`, `isNonInferencePath`)
- Request type classification logic
- 120+ lines of complex code
- Path-based test scenarios

---

## Quick Start (5 Minutes)

```bash
# 1. Build and push image
cd source/maas-metering-2
make build
make push

# 2. Deploy service
kubectl apply -k yaml/base/

# 3. Verify deployment
kubectl get pods -n maas-metering-2

# 4. Apply gateway policy
kubectl apply -f yaml/gateway-auth-policy-updated.yaml

# 5. Test with allow=true (should succeed)
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=true -n maas-metering-2
curl -k https://maas.apps.your-domain.com/llm/model/v1/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"test","prompt":"hello","max_tokens":10}'

# 6. Test with allow=false (should deny)
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=false -n maas-metering-2
curl -k https://maas.apps.your-domain.com/llm/model/v1/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"test","prompt":"hello","max_tokens":10}'
```

---

## File Structure

```
maas-metering-2/
├── README.md                    ← Architecture & overview
├── INSTALLATION.md              ← Step-by-step deployment
├── START-HERE.md               ← This file
│
├── cmd/
│   └── server/
│       └── main.go             ← Server entrypoint
│
├── internal/
│   └── api/
│       ├── handlers.go         ← Credit check logic (~60 lines)
│       └── handlers_test.go    ← Unit tests
│
├── yaml/
│   ├── base/
│   │   ├── namespace.yaml      ← Namespace: maas-metering-2
│   │   ├── serviceaccount.yaml ← Service account
│   │   ├── deployment.yaml     ← Deployment config
│   │   ├── service.yaml        ← Service definition
│   │   └── kustomization.yaml  ← Kustomize config
│   │
│   └── gateway-auth-policy-updated.yaml  ← Gateway AuthPolicy
│
├── Containerfile               ← Container image definition
├── Makefile                    ← Build commands
├── go.mod                      ← Go dependencies
└── go.sum                      ← Go checksums
```

---

## Common Tasks

### View Logs

```bash
# Follow logs
kubectl logs -f -n maas-metering-2 deployment/maas-metering-2

# View last 20 lines
kubectl logs -n maas-metering-2 deployment/maas-metering-2 --tail=20
```

### Change Credit Decision

```bash
# Allow all requests
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=true -n maas-metering-2

# Deny all requests
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=false -n maas-metering-2

# Wait for rollout
kubectl rollout status deployment/maas-metering-2 -n maas-metering-2
```

### Check AuthPolicy Status

```bash
# Quick check
kubectl get authpolicy gateway-auth-policy -n openshift-ingress

# Detailed status
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml | grep -A 10 "status:"
```

### Run Tests

```bash
# Unit tests
make test

# Verbose output
make test-verbose

# Run locally
make run-local ALLOW_VALUE=true
```

---

## Understanding the Code

### Core Handler (handlers.go)

```go
func (h *CreditCheckHandler) HandleAllow(c *gin.Context) {
    // 1. Extract headers (for logging)
    username := c.GetHeader("X-Username")
    
    // 2. Perform credit check (Phase 1: use env var)
    allow := h.allowValue
    
    // 3. Log the decision
    log.Printf("[CREDIT-CHECK] user=%s allow=%v", username, allow)
    
    // 4. Return decision
    c.JSON(200, AllowResponse{Allow: allow})
}
```

That's it! No path parsing, no complex logic.

### AuthPolicy Integration

```yaml
# AuthPolicy calls our service
metadata:
  credit-allow:
    http:
      url: "http://maas-metering-2.maas-metering-2.svc.cluster.local:8080/allow"
      headers:
        X-Username:
          selector: auth.identity.user.username

# OPA checks the result
authorization:
  credit-access:
    opa:
      rego: |
        allow {
          input.auth.metadata["credit-allow"].allow == true
        }
```

---

## Phase Roadmap

### ✅ Phase 1 (Current - MVP)
- Pure credit check service
- Boolean `ALLOW_VALUE` env var
- Logging to stdout
- Simplified architecture

### 🔮 Phase 2 (Future)
- Query actual credit balance (database/ConfigMap/external API)
- Return credit balance in response
- Per-user credit tracking
- Credit consumption recording

### 🔮 Phase 3 (Future)
- Real-time credit updates
- Credit balance thresholds
- Alert on low credits
- Credit purchase integration

---

## Key Insights

### 💡 Discovery: Route Precedence

**Problem:** v1 had complex path filtering logic in Go.

**Discovery:** Route-level AuthPolicies take precedence over gateway-level policies.

**Result:** Gateway policy only sees inference requests, so filtering is unnecessary!

### 💡 Single Responsibility

**Before:** Go service did routing AND credit checking.

**After:** Go service ONLY does credit checking.

**Benefit:** 
- 67% less code
- Easier to test
- Easier to maintain
- Easier to extend

### 💡 Observability Without Logic

**Kept:** Headers still passed for logging (path, user, method)

**Removed:** Logic that uses these headers for decisions

**Result:** Full observability without complexity.

---

## Comparison with v1

See [README.md](README.md#comparison-with-v1) for detailed comparison table.

**TL;DR:** v2 is simpler, faster, and easier to maintain while providing the same functionality.

---

## Need Help?

1. **Architecture questions?** → Read [README.md](README.md)
2. **Deployment issues?** → Follow [INSTALLATION.md](INSTALLATION.md)
3. **Code questions?** → Check inline comments in `internal/api/handlers.go`
4. **Kuadrant questions?** → https://docs.kuadrant.io
5. **Authorino questions?** → https://docs.kuadrant.io/authorino/

---

## Quick Reference Card

```bash
# Build
make build

# Deploy
kubectl apply -k yaml/base/

# Allow all
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=true -n maas-metering-2

# Deny all
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=false -n maas-metering-2

# Logs
kubectl logs -f -n maas-metering-2 deployment/maas-metering-2

# Status
kubectl get pods -n maas-metering-2
kubectl get authpolicy gateway-auth-policy -n openshift-ingress

# Test
curl -k $HOST/llm/model/v1/completions -H "Authorization: Bearer $TOKEN" -d '{...}'
```

---

## License & Author

**License:** Apache 2.0  
**Author:** Bryon Baker  
**Version:** 2.0

