# MaaS Metering Service - Installation Guide

Complete step-by-step guide to deploy the MaaS Metering service to your OpenShift/Kubernetes cluster.

---

## Prerequisites

Before you begin, ensure you have:

- ✅ OpenShift/Kubernetes cluster access with `kubectl` or `oc` CLI
- ✅ Kuadrant installed and configured
- ✅ Existing MaaS gateway (`maas-default-gateway`) deployed
- ✅ Podman installed for building container images
- ✅ Access to push images to `quay.io/bryonbaker` (or update to your registry)
- ✅ Cluster admin permissions (for AuthPolicy changes)

---

## Installation Steps

### Step 1: Build Container Image

Build and push the container image to your registry.

```bash
cd /home/bryon/git-repos/maas-gitops/source/maas-metering

# Build and push in one command
make push
```

**Expected output:**
```
Emulate Docker CLI using podman...
[1/2] STEP 1/8: FROM registry.access.redhat.com/ubi9/go-toolset:1.24 AS builder
...
Successfully tagged quay.io/bryonbaker/maas-metering:latest
Pushing to quay.io/bryonbaker/maas-metering:latest
...
✓ Image pushed successfully
```

**Troubleshooting:**
- If push fails with auth error: `podman login quay.io`
- To use a different registry, edit `IMAGE_NAME` in `Makefile`

---

### Step 2: Deploy Service to Cluster

Deploy the maas-metering service using Kustomize.

```bash
# Apply all resources
kubectl apply -k yaml/base/

# Expected output:
# namespace/maas-metering created
# serviceaccount/maas-metering created
# service/maas-metering created
# deployment.apps/maas-metering created
```

**Verify deployment:**

```bash
# Check namespace
kubectl get namespace maas-metering

# Check all resources
kubectl get all -n maas-metering

# Expected output:
# NAME                                READY   STATUS    RESTARTS   AGE
# pod/maas-metering-xxxxxxxxx-yyyyy   1/1     Running   0          30s
#
# NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
# service/maas-metering   ClusterIP   10.217.x.x      <none>        8080/TCP   30s
#
# NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
# deployment.apps/maas-metering   1/1     1            1           30s
```

**Wait for pod to be ready:**

```bash
kubectl rollout status deployment/maas-metering -n maas-metering

# Expected output:
# deployment "maas-metering" successfully rolled out
```

---

### Step 3: Test Service Directly

Verify the service is responding before integrating with AuthPolicy.

#### Test Health Endpoint

```bash
kubectl run test-health --rm -it --restart=Never --image=curlimages/curl -- \
  curl -v http://maas-metering.maas-metering.svc.cluster.local:8080/health

# Expected output:
# {"status":"healthy"}
```

#### Test Allow Endpoint (Non-Inference Path)

```bash
kubectl run test-allow --rm -it --restart=Never --image=curlimages/curl -- \
  curl -v http://maas-metering.maas-metering.svc.cluster.local:8080/allow \
  -H "X-Original-Path: /v1/models" \
  -H "X-Username: testuser" \
  -H "X-Method: GET"

# Expected output:
# {"allow":true}
```

#### Test Allow Endpoint (Inference Path)

```bash
kubectl run test-infer --rm -it --restart=Never --image=curlimages/curl -- \
  curl -v http://maas-metering.maas-metering.svc.cluster.local:8080/allow \
  -H "X-Original-Path: /llm/facebook-opt-125m-simulated/v1/completions" \
  -H "X-Username: testuser" \
  -H "X-Method: POST"

# Expected output:
# {"allow":true}  (because ALLOW_VALUE=true by default)
```

#### Check Logs

```bash
kubectl logs -n maas-metering deployment/maas-metering

# Expected output:
# [CREDIT-CHECK] path=/v1/models user=testuser method=GET type=non-inference allow=true duration_ms=0
# [CREDIT-CHECK] path=/llm/facebook-opt-125m-simulated/v1/completions user=testuser method=POST type=inference allow=true duration_ms=1
```

**If tests fail**, troubleshoot before proceeding:
- Check pod logs: `kubectl logs -n maas-metering deployment/maas-metering`
- Check pod status: `kubectl describe pod -n maas-metering`
- Check service endpoints: `kubectl get endpoints -n maas-metering`

---

### Step 4: Backup Existing Gateway AuthPolicy

**IMPORTANT:** Always back up before making changes to production policies.

```bash
# Create backup
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml > \
  ~/gateway-auth-policy-backup-$(date +%Y%m%d-%H%M%S).yaml

# Verify backup
ls -lh ~/gateway-auth-policy-backup-*.yaml
```

---

### Step 5: Apply Updated Gateway AuthPolicy

Apply the updated AuthPolicy that integrates credit checking.

```bash
cd /home/bryon/git-repos/maas-gitops/source/maas-metering

# Apply the updated policy
kubectl apply -f yaml/gateway-auth-policy-updated-gpt.yaml

# Expected output:
# authpolicy.kuadrant.io/gateway-auth-policy configured
```

**What this does:**
- Adds `metadata.credit-allow` fetch to call maas-metering service
- Passes `X-Original-Path`, `X-Username`, `X-Method` headers
- Adds `authorization.credit-access` rule to check the allow response
- Keeps all existing authentication, tier-access, and response rules

---

### Step 6: Verify AuthPolicy Status

Wait for the AuthPolicy to be enforced (may take 30-60 seconds).

```bash
# Watch status (Ctrl+C to exit)
watch -n 5 'kubectl get authpolicy gateway-auth-policy -n openshift-ingress \
  -o jsonpath="{.status.conditions[?(@.type==\"Enforced\")].status}" && echo'

# Should show: True
```

**Check full status:**

```bash
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml | grep -A 20 status:

# Expected output:
# status:
#   conditions:
#   - type: Accepted
#     status: "True"
#     message: AuthPolicy has been accepted
#   - type: Enforced
#     status: "True"
#     message: AuthPolicy has been successfully enforced
```

**If Enforced=False:**
- Check message for errors
- View Authorino logs: `kubectl logs -n kuadrant-system deployment/authorino -f`
- View maas-metering logs: `kubectl logs -n maas-metering deployment/maas-metering -f`
- See [Troubleshooting](#troubleshooting) section below

---

### Step 7: Test End-to-End (Allow Mode)

Test through the gateway with actual inference requests.

```bash
# Get cluster domain
CLUSTER_DOMAIN=$(kubectl get ingresses.config.openshift.io cluster -o jsonpath='{.spec.domain}')
HOST="https://maas.${CLUSTER_DOMAIN}"

# Get authentication token
TOKEN=$(oc whoami -t)

# Test 1: Get API token (non-inference, should always work)
TOKEN_RESPONSE=$(curl -sSk \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{"expiration": "10m"}' \
  "${HOST}/maas-api/v1/tokens")

echo $TOKEN_RESPONSE | jq

# Extract the MaaS token
MAAS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r .token)

# Test 2: List models (non-inference, should always work)
curl -sSk ${HOST}/v1/models \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MAAS_TOKEN" | jq

# Expected: List of available models

# Test 3: Run inference (should work with ALLOW_VALUE=true)
curl -sSk ${HOST}/llm/facebook-opt-125m-simulated/v1/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MAAS_TOKEN" \
  -d '{
    "model": "facebook/opt-125m",
    "prompt": "Once upon a time",
    "max_tokens": 50
  }' | jq

# Expected: Successful completion response
```

**Check service logs:**

```bash
kubectl logs -n maas-metering deployment/maas-metering -f

# Should see:
# [CREDIT-CHECK] path=/maas-api/v1/tokens user=... method=POST type=non-inference allow=true duration_ms=0
# [CREDIT-CHECK] path=/v1/models user=... method=GET type=non-inference allow=true duration_ms=0
# [CREDIT-CHECK] path=/llm/facebook-opt-125m-simulated/v1/completions user=... method=POST type=inference allow=true duration_ms=1
```

---

### Step 8: Test Deny Mode (Optional)

Test that credit checking works by switching to deny mode.

```bash
# Set service to deny inference requests
kubectl set env deployment/maas-metering -n maas-metering ALLOW_VALUE=false

# Wait for pod to restart
kubectl rollout status deployment/maas-metering -n maas-metering

# Wait for cache to expire (60 seconds)
echo "Waiting 65 seconds for cache to expire..."
sleep 65

# Try inference again (should be denied)
curl -sSk ${HOST}/llm/facebook-opt-125m-simulated/v1/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MAAS_TOKEN" \
  -d '{
    "model": "facebook/opt-125m",
    "prompt": "Once upon a time",
    "max_tokens": 50
  }'

# Expected: 403 Forbidden or similar error

# Try non-inference (should still work)
curl -sSk ${HOST}/v1/models \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MAAS_TOKEN" | jq

# Expected: ✓ List of models (non-inference bypasses credit check)
```

**Restore allow mode:**

```bash
kubectl set env deployment/maas-metering -n maas-metering ALLOW_VALUE=true
kubectl rollout status deployment/maas-metering -n maas-metering
```

---

## Rollback Procedure

If you need to rollback to the original state:

### Step 1: Restore Original AuthPolicy

```bash
# Apply your backup
kubectl apply -f ~/gateway-auth-policy-backup-YYYYMMDD-HHMMSS.yaml

# Wait for enforcement
watch -n 5 'kubectl get authpolicy gateway-auth-policy -n openshift-ingress \
  -o jsonpath="{.status.conditions[?(@.type==\"Enforced\")].status}" && echo'
```

### Step 2: Delete maas-metering Service (Optional)

```bash
# Delete all resources
kubectl delete -k yaml/base/

# Or just scale down
kubectl scale deployment/maas-metering -n maas-metering --replicas=0
```

### Step 3: Verify Gateway Works

```bash
# Test inference without credit check
curl -sSk ${HOST}/llm/model/v1/completions \
  -H "Authorization: Bearer $MAAS_TOKEN" \
  -d '{"model": "facebook/opt-125m", "prompt": "test", "max_tokens": 10}'

# Should work (back to original behavior)
```

---

## Troubleshooting

### AuthPolicy Not Enforced (`ENFORCED=False`)

**Check status message:**

```bash
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml | grep -A 10 "type: Enforced"
```

**Common causes:**

1. **AuthConfigs stuck in reconciliation**
   ```bash
   # List AuthConfigs
   kubectl get authconfig -n kuadrant-system
   
   # Delete stuck ones (they'll recreate)
   kubectl delete authconfig <hash> -n kuadrant-system
   ```

2. **Authorino can't reach maas-metering service**
   ```bash
   # Test from Authorino pod
   POD=$(kubectl get pod -n kuadrant-system -l app=authorino -o jsonpath='{.items[0].metadata.name}')
   kubectl exec -n kuadrant-system $POD -- curl -v http://maas-metering.maas-metering.svc.cluster.local:8080/health
   ```

3. **Restart Kuadrant operator**
   ```bash
   kubectl rollout restart deployment/kuadrant-operator-controller-manager -n kuadrant-system
   ```

### Service Not Receiving Requests

**Check if headers are configured:**

```bash
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml | grep -A 10 "X-Original-Path"

# Should show:
# X-Original-Path:
#   selector: context.request.http.path
```

**Check Authorino logs:**

```bash
kubectl logs -n kuadrant-system deployment/authorino -f | grep credit
```

### Inference Still Denied with ALLOW_VALUE=true

1. **Check pod has new environment variable:**
   ```bash
   kubectl get pod -n maas-metering -o yaml | grep ALLOW_VALUE
   ```

2. **Clear cache:**
   ```bash
   # Wait 60+ seconds or restart Authorino
   kubectl rollout restart deployment/authorino -n kuadrant-system
   ```

3. **Check service logs:**
   ```bash
   kubectl logs -n maas-metering deployment/maas-metering | grep inference
   ```

### DNS Resolution Issues

```bash
# Test DNS from Authorino namespace
kubectl run test-dns -n kuadrant-system --rm -it --image=curlimages/curl -- \
  nslookup maas-metering.maas-metering.svc.cluster.local
```

---

## Post-Installation

### Monitor Service

```bash
# Watch logs
kubectl logs -n maas-metering deployment/maas-metering -f

# Check resource usage
kubectl top pod -n maas-metering

# Check AuthPolicy status
kubectl get authpolicy -A
```

### Verify Metrics (Future)

Once metrics are implemented:

```bash
# Prometheus metrics endpoint
kubectl port-forward -n maas-metering deployment/maas-metering 8080:8080
curl http://localhost:8080/metrics
```

---

## Next Steps

### Phase 2: Real Credit Checking

Replace `ALLOW_VALUE` environment variable with actual credit logic:

1. Add ConfigMap or database for credit storage
2. Update `evaluateRequest()` in `handlers.go` to query real credits
3. Implement credit consumption tracking
4. Add credit management API

### Phase 3: Production Hardening

1. Add Prometheus metrics
2. Implement distributed caching (Redis)
3. Add OpenTelemetry tracing
4. Set up alerts for failures
5. Document operational runbooks

---

## Summary

✅ **You have successfully deployed MaaS Metering!**

The system now:
- ✅ Intercepts all requests to the MaaS gateway
- ✅ Allows non-inference requests instantly (< 1ms)
- ✅ Applies credit checking to inference requests
- ✅ Prevents reconciliation loops
- ✅ Logs all authorization decisions

**Key Files:**
- Service code: `internal/api/handlers.go`
- Tests: `internal/api/handlers_test.go`
- Gateway policy: `yaml/gateway-auth-policy-updated-gpt.yaml`
- Deployment: `yaml/base/`

**Key Commands:**
```bash
# View logs
kubectl logs -n maas-metering deployment/maas-metering -f

# Toggle allow/deny
kubectl set env deployment/maas-metering -n maas-metering ALLOW_VALUE=true|false

# Check policy status
kubectl get authpolicy gateway-auth-policy -n openshift-ingress
```

---

For more information, see:
- **[README.md](README.md)** - Architecture and overview
- **[IMPLEMENTATION-PLAN.md](IMPLEMENTATION-PLAN.md)** - Technical implementation details

