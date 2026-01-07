# MaaS Metering v2 - Installation Guide

**Version:** 2.0  
**Target:** OpenShift 4.x / Kubernetes 1.24+  
**Prerequisites:** Kuadrant, Podman/Docker, kubectl/oc

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Step 1: Build Container Image](#step-1-build-container-image)
4. [Step 2: Push to Registry](#step-2-push-to-registry)
5. [Step 3: Deploy Service](#step-3-deploy-service)
6. [Step 4: Verify Deployment](#step-4-verify-deployment)
7. [Step 5: Apply Gateway AuthPolicy](#step-5-apply-gateway-authpolicy)
8. [Step 6: Test Credit Check](#step-6-test-credit-check)
9. [Configuration](#configuration)
10. [Troubleshooting](#troubleshooting)
11. [Uninstallation](#uninstallation)

---

## Overview

This guide walks through deploying the **MaaS Metering v2** service, which provides credit checking for model inference requests in the MaaS platform.

**Key Features:**
- Pure credit check service (simplified from v1)
- No path filtering logic (handled by route precedence)
- Phase 1: Simple allow/deny based on `ALLOW_VALUE` env var
- Logs all inference requests to stdout

---

## Prerequisites

### Required Software

- **OpenShift CLI (`oc`)** or **kubectl** 1.24+
- **Podman** 4.0+ or **Docker** 20.10+
- **Go** 1.23+ (for building from source)
- **make** utility

### Cluster Requirements

- **Kuadrant** installed and configured
- **Gateway API** resources (`Gateway`, `HTTPRoute`)
- **Existing MaaS deployment** with gateway-level AuthPolicy
- Cluster admin access (for creating namespace and deploying resources)

### Verify Prerequisites

```bash
# Check cluster access
kubectl cluster-info

# Check Kuadrant installation
kubectl get crds | grep kuadrant

# Check Gateway API
kubectl get gateway maas-default-gateway -n openshift-ingress

# Check podman/docker
podman --version
# or
docker --version
```

---

## Step 1: Build Container Image

Navigate to the `maas-metering-2` directory and build the container image:

```bash
cd source/maas-metering-2

# Build with podman (recommended)
make build

# Or build manually
podman build -t maas-metering-2:latest .
```

**Expected output:**
```
Building container image with podman...
Container image built: maas-metering-2:latest
```

**Verify the image:**
```bash
podman images | grep maas-metering-2
```

---

## Step 2: Push to Registry

Push the image to your container registry:

```bash
# Push to quay.io (default)
make push

# Or push manually to a different registry
podman tag maas-metering-2:latest your-registry.com/your-org/maas-metering-2:latest
podman push your-registry.com/your-org/maas-metering-2:latest
```

**Update deployment YAML if using different registry:**

Edit `yaml/base/deployment.yaml`:
```yaml
spec:
  template:
    spec:
      containers:
      - name: maas-metering-2
        image: your-registry.com/your-org/maas-metering-2:latest  # Update this
```

---

## Step 3: Deploy Service

Deploy the service to your cluster:

```bash
# Deploy using kustomize
kubectl apply -k yaml/base/

# Or deploy individual files
kubectl apply -f yaml/base/namespace.yaml
kubectl apply -f yaml/base/serviceaccount.yaml
kubectl apply -f yaml/base/deployment.yaml
kubectl apply -f yaml/base/service.yaml
```

**Expected output:**
```
namespace/maas-metering-2 created
serviceaccount/maas-metering-2-sa created
deployment.apps/maas-metering-2 created
service/maas-metering-2 created
```

---

## Step 4: Verify Deployment

### Check Pod Status

```bash
# List pods
kubectl get pods -n maas-metering-2

# Expected output:
# NAME                                READY   STATUS    RESTARTS   AGE
# maas-metering-2-xxxxxxxxxx-xxxxx    1/1     Running   0          30s
```

### Check Service

```bash
# Verify service
kubectl get svc -n maas-metering-2

# Expected output:
# NAME              TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
# maas-metering-2   ClusterIP   172.30.xxx.xxx   <none>        8080/TCP   1m
```

### Check Logs

```bash
# View logs
kubectl logs -n maas-metering-2 deployment/maas-metering-2

# Expected output:
# 2026/01/07 15:00:00 Starting MaaS Metering Service v2...
# 2026/01/07 15:00:00 Environment: ALLOW_VALUE=true
# 2026/01/07 15:00:00 Server listening on :8080
```

### Test Health Endpoint

```bash
# Port-forward to local machine
kubectl port-forward -n maas-metering-2 svc/maas-metering-2 8080:8080 &

# Test health endpoint
curl -s http://localhost:8080/health | jq

# Expected output:
# {
#   "status": "healthy"
# }

# Stop port-forward
killall kubectl
```

---

## Step 5: Apply Gateway AuthPolicy

### Review the Policy

The gateway AuthPolicy needs to be updated to call the `maas-metering-2` service:

```bash
cat yaml/gateway-auth-policy-updated.yaml
```

Key section to verify:
```yaml
metadata:
  credit-allow:
    http:
      url: "http://maas-metering-2.maas-metering-2.svc.cluster.local:8080/allow"
```

### Apply the Policy

```bash
# Apply the updated gateway policy
kubectl apply -f yaml/gateway-auth-policy-updated.yaml

# Verify policy is accepted
kubectl get authpolicy gateway-auth-policy -n openshift-ingress

# Check detailed status
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml
```

**Expected conditions:**
```yaml
status:
  conditions:
  - type: Accepted
    status: "True"
  - type: Enforced
    status: "True"
```

**Note:** It may take 10-30 seconds for the policy to become `Enforced: True`.

---

## Step 6: Test Credit Check

### Get Authentication Token

First, obtain a valid MaaS token:

```bash
# Example for OpenShift
export MAAS_TOKEN=$(oc whoami -t)

# Or request a token via API
export HOST="https://maas.apps.your-domain.com"
```

### Test 1: Non-Inference Request (Should NOT Hit Service)

```bash
# List models endpoint (handled by route-level policy)
curl -sSk ${HOST}/v1/models \
  -H "Authorization: Bearer $MAAS_TOKEN" | jq

# Check logs - should be EMPTY (service not called)
kubectl logs -n maas-metering-2 deployment/maas-metering-2 --tail=10
```

**Expected:** No new log entries (route-level policy handles this).

### Test 2: Inference Request with Credits (Allow)

```bash
# Ensure ALLOW_VALUE=true
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=true -n maas-metering-2

# Wait for rollout
kubectl rollout status deployment/maas-metering-2 -n maas-metering-2

# Make inference request
curl -sSk ${HOST}/llm/facebook-opt-125m-simulated/v1/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MAAS_TOKEN" \
  -d '{
    "model": "facebook/opt-125m",
    "prompt": "Once upon a time",
    "max_tokens": 50
  }' | jq

# Check logs - should show credit check with allow=true
kubectl logs -n maas-metering-2 deployment/maas-metering-2 --tail=5
```

**Expected log:**
```
[CREDIT-CHECK] path=/llm/facebook-opt-125m-simulated/v1/completions user=system:serviceaccount:... method=POST allow=true duration_ms=0
```

**Expected response:** Model completion JSON.

### Test 3: Inference Request without Credits (Deny)

```bash
# Set ALLOW_VALUE=false
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=false -n maas-metering-2

# Wait for rollout
kubectl rollout status deployment/maas-metering-2 -n maas-metering-2

# Make inference request
curl -sSk ${HOST}/llm/facebook-opt-125m-simulated/v1/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MAAS_TOKEN" \
  -d '{
    "model": "facebook/opt-125m",
    "prompt": "Once upon a time",
    "max_tokens": 50
  }'

# Check logs - should show credit check with allow=false
kubectl logs -n maas-metering-2 deployment/maas-metering-2 --tail=5
```

**Expected log:**
```
[CREDIT-CHECK] path=/llm/facebook-opt-125m-simulated/v1/completions user=system:serviceaccount:... method=POST allow=false duration_ms=0
```

**Expected response:**
```json
{
  "error": "access_denied",
  "message": "Access denied: You do not have sufficient credits or permissions to perform this inference request.",
  "code": 403
}
```

---

## Configuration

### Environment Variables

The service behavior is controlled by the `ALLOW_VALUE` environment variable:

**Allow all inference requests:**
```bash
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=true -n maas-metering-2
```

**Deny all inference requests:**
```bash
kubectl set env deployment/maas-metering-2 ALLOW_VALUE=false -n maas-metering-2
```

**Change server port (optional):**
```bash
kubectl set env deployment/maas-metering-2 PORT=9090 -n maas-metering-2
# Also update service port in yaml/base/service.yaml
```

### Scaling

Scale the deployment for high availability:

```bash
# Scale to 3 replicas
kubectl scale deployment maas-metering-2 --replicas=3 -n maas-metering-2

# Verify scaling
kubectl get pods -n maas-metering-2
```

### Resource Limits

Edit `yaml/base/deployment.yaml` to adjust resources:

```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "128Mi"
    cpu: "200m"
```

Apply changes:
```bash
kubectl apply -f yaml/base/deployment.yaml
```

---

## Troubleshooting

### Issue: Pod not starting

**Check pod events:**
```bash
kubectl describe pod -n maas-metering-2 -l app=maas-metering-2
```

**Common causes:**
- Image pull failure (check registry credentials)
- Resource constraints (check node capacity)
- Invalid environment variables

### Issue: Service not called for inference requests

**Check gateway policy:**
```bash
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml
```

Verify:
1. `targetRef` points to correct Gateway
2. `metadata.credit-allow.http.url` points to `maas-metering-2.maas-metering-2.svc.cluster.local:8080`
3. Policy status shows `Enforced: True`

**Check Authorino logs:**
```bash
kubectl logs -n kuadrant-system deployment/authorino | grep credit-allow
```

### Issue: AuthPolicy not enforced

**Check policy status:**
```bash
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o jsonpath='{.status.conditions}' | jq
```

**If Enforced=False:**
- Wait 30 seconds (reconciliation takes time)
- Check for conflicting policies
- Restart Kuadrant operator:
  ```bash
  kubectl rollout restart deployment -n kuadrant-system
  ```

### Issue: Wrong credit decision

**Check ALLOW_VALUE:**
```bash
kubectl get deployment maas-metering-2 -n maas-metering-2 -o jsonpath='{.spec.template.spec.containers[0].env}' | jq
```

**Force pod restart:**
```bash
kubectl rollout restart deployment/maas-metering-2 -n maas-metering-2
```

### Issue: Service called for non-inference requests

This should NOT happen if route-level policies are configured correctly.

**Check route-level policies:**
```bash
# Check maas-api-route policy
kubectl get authpolicy maas-api-auth-policy -n maas-api -o yaml

# Verify it targets the correct route
# targetRef:
#   kind: HTTPRoute
#   name: maas-api-route
```

---

## Uninstallation

### Remove Gateway Policy Updates

**Option 1: Revert to original policy**
```bash
# Restore backup (if you have one)
kubectl apply -f yaml/gateway-auth-policy-backup.yaml
```

**Option 2: Remove credit-check sections manually**

Edit the gateway policy and remove:
- `metadata.credit-allow` section
- `authorization.credit-access` section

### Remove Service Deployment

```bash
# Delete all resources
kubectl delete -k yaml/base/

# Or delete individually
kubectl delete -f yaml/base/service.yaml
kubectl delete -f yaml/base/deployment.yaml
kubectl delete -f yaml/base/serviceaccount.yaml
kubectl delete -f yaml/base/namespace.yaml
```

### Verify Cleanup

```bash
# Check namespace is gone
kubectl get ns | grep maas-metering-2

# Check gateway policy is restored
kubectl get authpolicy gateway-auth-policy -n openshift-ingress -o yaml | grep credit
```

---

## Next Steps

1. **Monitor logs** to understand credit check patterns
2. **Plan Phase 2** implementation (actual credit balance lookup)
3. **Test with different user tiers** to verify authorization flow
4. **Consider HA setup** (multiple replicas, pod disruption budgets)
5. **Implement observability** (Prometheus metrics, distributed tracing)

---

## Support

For issues or questions:
- Check the [README.md](README.md) for architecture details
- Review Kuadrant documentation: https://docs.kuadrant.io
- Check Authorino documentation: https://docs.kuadrant.io/authorino/

---

## License

Apache License 2.0

## Author

Bryon Baker

