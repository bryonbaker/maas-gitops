# MaaS Toolbox Test Harness - Quick Start Guide

## Running the Tests

### Basic Usage

```bash
# Run against deployed cluster
./test-api.sh https://maas-toolbox-maas-dev.apps.example.com

# Or set via environment variable
export BASE_URL=https://maas-toolbox-maas-dev.apps.example.com
./test-api.sh

# Run against localhost (LLM tests will be skipped)
./test-api.sh http://localhost:8080
```

---

## Prerequisites

### Minimum Requirements (for Tier & Group tests only)
- OpenShift/Kubernetes cluster access
- `oc` CLI tool installed and logged in
- MaaS Toolbox API deployed and accessible

### Full Test Suite Requirements (includes LLM tests)
- All minimum requirements above
- KServe operator installed (for LLMInferenceService CRD)
- Namespace `llm` exists
- LLMInferenceService `facebook-opt-125m-simulated` exists in `llm` namespace

---

## Pre-Check Validation

The test harness automatically checks for dependencies before running:

### Example: All Dependencies Present ✅

```
========================================
PRE-CHECK: LLMInferenceService Dependencies
========================================

✓ Namespace 'llm' found
✓ LLMInferenceService 'facebook-opt-125m-simulated' found in namespace 'llm'
All LLMInferenceService dependencies satisfied
```

### Example: Missing LLM Dependencies ⚠️

```
========================================
PRE-CHECK: LLMInferenceService Dependencies
========================================

FAILED: Required LLMInferenceService not found
Expected: facebook-opt-125m-simulated in namespace llm
LLMInferenceService tests will be marked as FAILED

To enable these tests, ensure the following resource exists:
  Resource: llminferenceservices.kserve.io/facebook-opt-125m-simulated
  Namespace: llm

LLMInferenceService tests will be skipped or marked as failed
```

---

## Test Sections

### TEST 1-11: Tier & Group Operations (Always Run)
- ✅ TEST 1: Create Tiers
- ✅ TEST 2: List All Tiers
- ✅ TEST 3: Get Specific Tier
- ✅ TEST 4: Update Tiers
- ✅ TEST 5: Add Groups to Tiers
- ✅ TEST 6: Remove Groups from Tiers
- ✅ TEST 7: Get Tiers by Group
- ✅ TEST 8: Verify Updates
- ✅ TEST 9: Delete Tiers
- ✅ TEST 10: Verify All Tiers Deleted
- ✅ TEST 11: Edge Cases

### TEST 12-15: LLM Operations (Conditional - Requires Dependencies)
- 🔄 TEST 12: Annotate LLMInferenceService with Tiers
- 🔄 TEST 13: Get LLMInferenceServices by Tier
- 🔄 TEST 14: Get LLMInferenceServices by Group
- 🔄 TEST 15: Remove Tier from LLMInferenceService

---

## Understanding Test Output

### Test Result Formats

```bash
# PASSING TEST
Testing: Create acme-inc-1 tier ... PASS (HTTP 201)
  Response: {"name":"acme-inc-1","description":"Acme Inc Tier 1","level":1,"groups":["system:authenticated"]}

# FAILING TEST
Testing: Try to create duplicate tier ... FAIL (Expected HTTP 409, got HTTP 200)
  Response: {"error":"tier already exists"}

# SKIPPED TEST (Missing LLM dependency)
Testing: Annotate service with tier ... SKIPPED (Missing dependency: facebook-opt-125m-simulated)
```

### Color Coding
- 🟢 **GREEN (PASS):** Test passed successfully
- 🔴 **RED (FAIL):** Test failed
- 🟡 **YELLOW (SKIPPED):** Test skipped due to missing dependencies or warnings

---

## Test Summary Examples

### Example 1: All Tests Pass ✅

```
========================================
TEST SUMMARY
========================================

Total Tests: 95
Passed: 95
Failed: 0

All tests passed!
```

**Exit Code:** 0

---

### Example 2: Missing LLM Dependencies ⚠️

```
========================================
TEST SUMMARY
========================================

Total Tests: 95
Passed: 65
Failed: 30

⚠ LLMInferenceService Dependency Missing
Required: facebook-opt-125m-simulated in namespace llm
LLMInferenceService tests were skipped/failed

Some tests failed!
```

**Exit Code:** 1  
**Note:** The 30 "failed" tests are actually SKIPPED due to missing dependencies

---

### Example 3: Actual Test Failures ❌

```
========================================
TEST SUMMARY
========================================

Total Tests: 95
Passed: 90
Failed: 5

Some tests failed!
```

**Exit Code:** 1  
**Action:** Review test output to identify which tests failed and why

---

## Troubleshooting

### Error: "Server is not accessible"

```
Error: Server is not accessible at https://example.com
Please check:
  - The server is running
  - The URL is correct
  - Network connectivity
```

**Solution:**
- Verify MaaS Toolbox is deployed: `oc get pods -n maas-dev`
- Check route exists: `oc get route maas-toolbox -n maas-dev`
- Test connectivity: `curl -k https://your-route-url/health`

---

### Error: "Not logged into OpenShift cluster"

```
Error: Not logged into OpenShift cluster. Please run 'oc login'.
```

**Solution:**
```bash
oc login https://api.your-cluster.com:6443
```

---

### Error: "Cluster domain mismatch"

```
Error: Cluster domain mismatch!
  URL domain: example1.com
  Cluster domain: example2.com
  Please ensure you are logged into the correct cluster.
```

**Solution:**
- Login to the correct cluster that matches your BASE_URL
- Or update BASE_URL to match your current cluster

---

### Warning: "LLMInferenceService CRD not found"

```
WARNING: LLMInferenceService CRD not found in cluster
LLMInferenceService tests will be skipped
```

**Solution:** Install KServe operator:
```bash
# Install via OperatorHub or:
oc apply -f kserve-operator-subscription.yaml
```

---

### Error: "Required namespace 'llm' not found"

```
FAILED: Required namespace 'llm' not found
LLMInferenceService tests require namespace: llm
```

**Solution:**
```bash
oc create namespace llm
```

---

### Error: "Required LLMInferenceService not found"

```
FAILED: Required LLMInferenceService not found
Expected: facebook-opt-125m-simulated in namespace llm
```

**Solution:** Deploy the required LLMInferenceService:
```bash
# Check what services exist
oc get llminferenceservices -n llm

# Deploy the required service (if you have the manifest)
oc apply -f facebook-opt-125m-simulated.yaml -n llm

# Or use any existing service (you'll need to modify the test script)
```

---

## Expected Test Duration

| Scenario | Tests Run | Duration |
|----------|-----------|----------|
| Full suite (with LLM) | ~95 tests | 4-5 minutes |
| Without LLM dependencies | ~65 tests | 2-3 minutes |
| Localhost only | ~65 tests | 1-2 minutes |

---

## Common Use Cases

### Use Case 1: CI/CD Pipeline Testing

```bash
#!/bin/bash
set -e

# Get the route URL
ROUTE_URL=$(oc get route maas-toolbox -n maas-dev -o jsonpath='{.spec.host}')

# Run tests
./test-api.sh "https://${ROUTE_URL}"

# Exit code 0 = success, non-zero = failure
```

---

### Use Case 2: Local Development Testing

```bash
# Terminal 1: Run MaaS Toolbox locally
cd source/maas-toolbox
go run cmd/server/main.go

# Terminal 2: Run tests
cd tests
./test-api.sh http://localhost:8080
```

---

### Use Case 3: Testing Specific Tier/Group Configuration

The test suite automatically creates and cleans up test groups:
- `maas-toolbox-premium-users`
- `maas-toolbox-enterprise-users`
- `maas-toolbox-vip-users`
- `maas-toolbox-trial-users`
- `maas-toolbox-free-users`
- `maas-toolbox-beta-users`
- `maas-toolbox-alpha-users`
- `maas-toolbox-test-group`
- `maas-toolbox-shared-group`

These groups are automatically deleted at the end of the test run.

---

## Test Isolation

### What Gets Created During Tests
- **Test Groups:** Multiple `maas-toolbox-*` groups in OpenShift
- **Test Tiers:** Various tiers like `acme-inc-1`, `llm-test-tier-1`, etc.
- **Annotations:** Temporary tier annotations on pre-existing LLMInferenceService

### What Gets Cleaned Up
- ✅ All test tiers are deleted
- ✅ All test groups are deleted
- ✅ All tier annotations are removed from LLMInferenceService
- ✅ Pre-existing LLMInferenceService is left unchanged

### What Is NOT Modified
- ❌ Pre-existing tiers (if any)
- ❌ Pre-existing groups (not matching `maas-toolbox-*` pattern)
- ❌ LLMInferenceService itself (only annotations added/removed)
- ❌ Any other namespaces or resources

---

## Exit Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | All tests passed successfully |
| 1 | One or more tests failed OR LLM dependencies missing |

**Note:** Exit code 1 is returned even when tests are only SKIPPED due to missing dependencies. This is intentional to flag that the full test suite did not run.

---

## Dry Run (Syntax Check Only)

To validate the script without running tests:

```bash
bash -n test-api.sh
# No output = no syntax errors
```

---

## Tips & Best Practices

1. **Always login to cluster first:** `oc login ...` before running tests
2. **Check cluster domain matches URL:** Avoid authentication issues
3. **Run in isolated namespace:** Use test/dev namespace, not production
4. **Review output carefully:** Even passing tests may have warnings
5. **Monitor resource usage:** Tests create/delete many resources
6. **Use unique tier names:** Avoid conflicts with existing tiers

---

## Quick Reference Commands

```bash
# Check if server is accessible
curl -k https://your-route-url/health

# Check if LLM dependencies exist
oc get namespace llm
oc get llminferenceservice facebook-opt-125m-simulated -n llm

# Check if KServe CRD is installed
oc get crd llminferenceservices.kserve.io

# View test groups that will be created
grep "TEST_GROUPS=" test-api.sh -A 10

# Check current cluster domain
oc get ingresses.config.openshift.io cluster -o jsonpath='{.spec.domain}'

# Run a quick syntax check
bash -n test-api.sh
```

---

## Getting Help

If you encounter issues not covered in this guide:

1. Check the main **TEST-HARNESS-UPDATE-SUMMARY.md** for detailed implementation notes
2. Review the **MAAS-DEPLOYMENT-SUMMARY.md** for deployment guidance
3. Check MaaS Toolbox logs: `oc logs deployment/maas-toolbox -n maas-dev`
4. Verify API health: `curl -k https://your-route-url/health`
5. Check Swagger docs: `https://your-route-url/swagger/index.html`

---

**Happy Testing! 🚀**

