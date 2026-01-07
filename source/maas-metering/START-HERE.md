# MaaS Metering Service - Start Here 📍

Welcome! This guide will help you navigate the documentation and get started quickly.

---

## 📖 Documentation Guide

### For Quick Start

1. **[README.md](README.md)** - Start here
   - Overview and architecture
   - Path detection strategy (OpenAI API v1)
   - API endpoints
   - Local development and testing
   - Performance characteristics

2. **[INSTALLATION.md](INSTALLATION.md)** - Deployment guide
   - Step-by-step installation instructions
   - Testing procedures
   - Troubleshooting
   - Rollback procedures

### For Developers

3. **[IMPLEMENTATION-PLAN.md](IMPLEMENTATION-PLAN.md)** - Technical details
   - Complete implementation plan
   - Code examples
   - Test cases
   - Phase-by-phase breakdown
   - Future roadmap

---

## 🚀 Quick Start (3 Steps)

### 1. Run Tests

```bash
cd /home/bryon/git-repos/maas-gitops/source/maas-metering
go test -v ./internal/api/
```

### 2. Test Locally

```bash
# Terminal 1: Start service
make run-local

# Terminal 2: Test it
curl http://localhost:8080/allow \
  -H "X-Original-Path: /v1/models" \
  -H "X-Username: testuser"

# Expected: {"allow":true}
```

### 3. Deploy to Cluster

See **[INSTALLATION.md](INSTALLATION.md)** for complete instructions.

```bash
# Quick version:
make push
kubectl apply -k yaml/base/
kubectl apply -f yaml/gateway-auth-policy-updated-gpt.yaml
```

---

## 📂 Key Files

### Source Code

| File | Purpose |
|------|---------|
| `internal/api/handlers.go` | Credit check logic with path detection |
| `internal/api/handlers_test.go` | Comprehensive unit tests (100+ test cases) |
| `internal/api/router.go` | HTTP route setup |
| `cmd/server/main.go` | Application entry point |

### Configuration

| File | Purpose |
|------|---------|
| `yaml/base/deployment.yaml` | Kubernetes deployment config |
| `yaml/base/service.yaml` | Kubernetes service (ClusterIP) |
| `yaml/gateway-auth-policy-updated-gpt.yaml` | Gateway AuthPolicy with credit check |
| `Makefile` | Build, test, push targets |
| `Containerfile` | Multi-stage container build |

### Documentation

| File | Purpose |
|------|---------|
| `README.md` | Overview and getting started |
| `INSTALLATION.md` | Deployment guide |
| `IMPLEMENTATION-PLAN.md` | Technical implementation details |
| `START-HERE.md` | This file |

---

## 🔍 How It Works

### Path Detection (OpenAI API Version)

The service detects inference requests by looking for `/v1/` as the **3rd segment** of the path.

**Pattern**: `/{namespace}/{model}/v1/{endpoint}`

- The `/v1/` indicates **OpenAI API version 1**
- Works for any namespace: `llm`, `acme-inc-models`, `serverless-models`, etc.

**Examples**:

✅ **Inference** (credit check applies):
```
/llm/facebook-opt-125m-simulated/v1/completions
/acme-inc-models/acme-inc-model-1/v1/chat/completions
/serverless-models/my-model/v1/completions
```

✅ **Non-Inference** (always allowed):
```
/v1/models                  (OpenAI: list models)
/maas-api/v1/tokens         (MaaS API: get tokens)
/oauth2                     (OAuth)
```

### Authorization Flow

```
1. Request arrives at Gateway
2. AuthPolicy calls maas-metering with X-Original-Path
3. Service analyzes path:
   ├─ Non-inference? → return {"allow": true} (< 1ms)
   └─ Inference? → check credits → return {"allow": true/false}
4. OPA checks if allow==true
5. Gateway allows/denies request
```

---

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test -v ./internal/api/

# Tests cover:
# - Path detection (30+ test cases)
# - Request evaluation (15+ test cases)
# - Edge cases and variations
```

### Local Testing

```bash
# Allow mode
make run-local

# Deny mode
ALLOW_VALUE=false make run-local

# Test with curl (see examples in README.md)
```

### Integration Testing

See **[INSTALLATION.md](INSTALLATION.md)** Section 7 for end-to-end testing procedures.

---

## 🎯 Current Status

### ✅ Phase 1 Complete (MVP)
- [x] Path-aware authorization
- [x] Service-side filtering logic
- [x] Gateway-level AuthPolicy integration
- [x] Comprehensive unit tests (7 test suites)
- [x] Structured logging
- [x] Documentation

### 📋 Next: Phase 2 (Real Credit Checking)
- [ ] Replace `ALLOW_VALUE` with actual credit lookup
- [ ] Add ConfigMap or database integration
- [ ] Implement per-user credit tracking
- [ ] Add credit consumption API

---

## 💡 Key Concepts

### Single Enforcement Point

- **One gateway-level AuthPolicy** applies to all routes
- **Service-side filtering** handles all path logic
- **No per-model policies** - scales to unlimited models

### Always Returns 200 OK

- **Prevents reconciliation loops** that plagued earlier attempts
- Non-inference → `{"allow": true}` (no credit check)
- Inference → `{"allow": true/false}` (credit check result)
- Unknown → `{"allow": false}` (fail-closed)

### Fast & Efficient

- Non-inference: < 1ms (immediate bypass)
- Inference: 1-2ms (simple check)
- Caching: 60 seconds per user

---

## 📞 Need Help?

1. **Common issues?** → See [INSTALLATION.md - Troubleshooting](INSTALLATION.md#troubleshooting)
2. **How does it work?** → See [README.md - Architecture](README.md#architecture)
3. **Want to contribute?** → See [IMPLEMENTATION-PLAN.md](IMPLEMENTATION-PLAN.md)

---

## 🎓 Learn More

- **OpenAI API v1 Spec**: https://platform.openai.com/docs/api-reference
- **Kuadrant AuthPolicy**: https://docs.kuadrant.io/kuadrant-operator/doc/auth/
- **Authorino**: https://github.com/Kuadrant/authorino

---

**Ready to deploy?** → Start with **[INSTALLATION.md](INSTALLATION.md)**

**Want to understand the code?** → Start with **[README.md](README.md)**

**Need technical details?** → See **[IMPLEMENTATION-PLAN.md](IMPLEMENTATION-PLAN.md)**

