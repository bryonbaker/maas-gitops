# Sync Waves Quick Reference

## Wave Timeline Visualization

```
┌──────────────────────────────────────────────────────────────┐
│                    DEPLOYMENT TIMELINE                        │
└──────────────────────────────────────────────────────────────┘

Wave 3:  [Cluster Configuration]
         └─ cluster-config-app (monitoring, cluster basics)
         
Wave 10: [Foundation Operators]
         └─ cert-manager-operator
         
Wave 12: [Independent Operators]
         ├─ nfd-operator ⭐ (needed for GPU discovery)
         └─ kueue-operator
         
Wave 14: [GPU Operators]
         └─ nvidia-operator (depends on NFD)
         
Wave 16: [Distributed Workload Operators]
         └─ lws-operator
         
Wave 18: [AI/ML Platform - LAST OPERATOR]
         └─ ocpai-operator (depends on kueue, nvidia, lws)
         
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Wave 20: [🚦 HEALTH CHECK GATE] ⚠️  CRITICAL
         └─ operators-health-check-job
              ├─ Validates all CSVs are Succeeded
              ├─ Checks all CRDs are established  
              ├─ Lists operator pods
              └─ BLOCKS all instances until success

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Wave 30: [First Instance]
         └─ nfd-instance
         
Wave 35: [NVIDIA Configuration]
         └─ nvidia-cluster-policy (depends on NFD instance)
         
Wave 36: [Distributed Workload Instances]
         └─ lws-instance

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Wave 38: [🚦 INSTANCES HEALTH CHECK GATE] ⚠️  SECOND GATE
         └─ instances-health-check-job
              ├─ Validates NFD instance discovering nodes
              ├─ Checks NVIDIA GPU drivers deployed
              ├─ Ensures LWS controller running
              └─ BLOCKS OpenShift AI DSC until success

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Wave 42: [OpenShift AI Platform Instance]
         └─ ocpai-dsc-instance (after instances health check)
         
Wave 50: [Infrastructure]
         └─ machine-set-app
         
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Wave 80: [Users & RBAC - Applications Start]
         └─ users-app.yaml
         
Wave 82: [MAAS Configuration]
         └─ maas-tier-config-app
         
Wave 85: [Model Deployment Apps]
         └─ model-deployments-app

Wave 87: [GPU Models]
         └─ gpu-models-app
         
Wave 90: [Authentication - Deploy Last]
         └─ keycloak-app (commented out by default)
```

---

## File → Wave Mapping

| Wave | File | Purpose | Dependencies |
|------|------|---------|--------------|
| 3 | cluster-config-app.yaml | Monitoring/cluster config | None |
| 10 | cert-manager-operator-application.yaml | Certificates | None |
| 12 | nfd-operator-application.yaml | Node features | None |
| 12 | kueue-operator-application.yaml | Job queuing | None |
| 14 | nvidia-operator-application.yaml | GPU operator | NFD |
| 16 | lws-operator-application.yaml | Leader-Worker Sets | None |
| 18 | ocpai-operator.yaml | OpenShift AI | Kueue, NVIDIA, LWS |
| **20** | **operators-health-check-job.yaml** | **Health gate** | **All operators** |
| 30 | nfd-instance.yaml | NFD instance | NFD operator healthy |
| 35 | nvidia-cluster-policy-application.yaml | GPU policy | NFD instance |
| 36 | lws-instance.yaml | LWS instance | LWS operator healthy |
| **38** | **instances-health-check-job.yaml** | **Instances gate** | **NFD, NVIDIA, LWS instances** |
| 42 | ocpai-dcs-instance.yaml | AI platform | All instances healthy |
| 50 | machine-set-app.yaml | Compute nodes | Operators ready |
| 80 | users-app.yaml | Users & RBAC | Platform ready |
| 82 | maas-tier-config-app.yaml | Multi-tenancy | Platform ready |
| 85 | model-deployments-app.yaml | Model serving apps | All platform ready |
| 87 | gpu-models-app.yaml | Model configs | MAAS config ready |
| 90 | keycloak-app.yaml | Auth/SSO service | All apps ready |

---

## Operator → CRD Mapping

| Operator | Namespace | CRD(s) Checked by Health Job |
|----------|-----------|------------------------------|
| NFD | `openshift-nfd` | `nodefeaturediscoveries.nfd.openshift.io` |
| NVIDIA | `nvidia-gpu-operator` | `clusterpolicies.nvidia.com` |
| OCPAI | `redhat-ods-operator` | `datascienceclusters.datasciencecluster.opendatahub.io` |
| Kueue | `openshift-kueue-operator` | `clusterqueues.kueue.x-k8s.io`<br>`localqueues.kueue.x-k8s.io` |
| LWS | `openshift-lws-operator` | `leaderworkersets.leaderworkerset.x-k8s.io` |
| Cert Manager | `cert-manager-operator` | (checked via CSV only) |

---

## Common Commands

### Monitor Deployment Progress
```bash
# Watch all applications
watch kubectl get applications -n openshift-gitops

# Check specific app
kubectl get application <app-name> -n openshift-gitops -o yaml

# View health check logs
kubectl logs -n openshift-gitops job/operators-health-check
```

### Check Operator Status
```bash
# All CSVs (should be "Succeeded")
kubectl get csv -A | grep -v Succeeded

# Operator pods
kubectl get pods -n openshift-nfd
kubectl get pods -n nvidia-gpu-operator
kubectl get pods -n redhat-ods-operator
kubectl get pods -n openshift-kueue-operator
kubectl get pods -n openshift-lws-operator
```

### Check CRDs
```bash
# List all operator CRDs
kubectl get crd | grep -E "(nfd|nvidia|datasciencecluster|kueue|leaderworkerset)"

# Check specific CRD status
kubectl get crd nodefeaturediscoveries.nfd.openshift.io -o yaml
```

### Manual Sync (if needed)
```bash
# Sync specific app
argocd app sync <app-name> --prune

# Sync all apps (careful!)
argocd app sync -l app.kubernetes.io/part-of=platform
```

### Force Re-run Health Checks
```bash
# Delete operators health check job (wave 20)
kubectl delete job operators-health-check -n openshift-gitops
argocd app sync operators-health-check

# Delete instances health check job (wave 38)
kubectl delete job instances-health-check -n openshift-gitops
argocd app sync instances-health-check
```

---

## Retry Configuration (Instances)

All instances are configured with automatic retry:

```yaml
retry:
  limit: 10              # Try up to 10 times
  backoff:
    duration: 30s        # First retry after 30s
    factor: 2            # Double each time
    maxDuration: 10m     # Cap at 10 minutes
```

**Retry schedule:** 30s → 1m → 2m → 4m → 8m → 10m → 10m → 10m → 10m → 10m

---

## Troubleshooting Matrix

| Symptom | Likely Cause | Solution |
|---------|--------------|----------|
| Apps stuck at sync wave 20 | Operators health check failing | Check `kubectl logs -n openshift-gitops job/operators-health-check` |
| Apps stuck at sync wave 38 | Instances health check failing | Check `kubectl logs -n openshift-gitops job/instances-health-check` |
| OCPAI DSC stuck "OutOfSync" | Waiting for instances gate | Check instances health check job status |
| Instance "OutOfSync" | CRD not ready yet | Wait for retry; check operator CSV status |
| Operator CSV "Installing" | Operator still deploying | Give it more time; check operator pods |
| "SkipDryRunOnMissingResource" | CRD not established | Normal; retry will handle it automatically |
| Health check job "Failed" | One or more operators unhealthy | Check which operator CSV is not "Succeeded" |
| Apps deploying out of order | Sync wave annotation missing/wrong | Check application YAML annotations |

---

## Best Practices

1. **Don't skip the health check**
   - It's the only gate ensuring operators are truly ready
   - Prevents cryptic errors from instances deploying too early

2. **Monitor the first deployment carefully**
   - Watch for operators that take longer than expected
   - Note any that need wave adjustment

3. **Respect wave gaps**
   - Don't reduce gaps below recommended values
   - Cluster stability > deployment speed

4. **Use ArgoCD UI for visibility**
   - Application dependencies are clearly shown
   - Easier to spot issues than CLI

5. **Trust the retry logic**
   - Don't manually intervene too quickly
   - Let the backoff algorithm work

---

## When to Use Bash Script Instead

Consider a bash script for:
- ✅ Initial cluster bootstrap (one-time)
- ✅ Testing operator installation order
- ✅ Custom health checks beyond CSV/CRD
- ✅ Integration with external systems
- ✅ Non-GitOps environments

Use sync waves for:
- ✅ Production GitOps workflows
- ✅ Self-healing infrastructure
- ✅ Continuous deployment from Git
- ✅ Multiple clusters with same config
- ✅ Day-2 operations

---

## Dependencies Graph

```
cert-manager ──────────┐
                       │
nfd ────────────┬──────┼──► [Health Check Wave 20]
                │      │            │
nvidia ◄────────┘      │            ├──► nfd-instance
                       │            │         │
kueue ─────────────────┤            │         ├──► nvidia-cluster-policy
                       │            │         │
lws ───────────────────┤            ├──► lws-instance
                       │            │
ocpai ─────────────────┘            ├──► ocpai-dsc-instance
  (depends on all)                  │         │
                                    │         │
                                    │         ├──► machine-sets
                                    │         │
                                    │         ├──► gpu-models
                                    │         │         │
                                    │         │         ├──► model-deployments
                                    │         │         │         │
                                    │         │         │         ├──► maas-config
                                    │         │         │         │
                                    │         │         │         └──► monitoring
```

---

## Quick Deployment Test

After applying root application:

```bash
# 1. Check operators deploy in order
watch kubectl get csv -A

# 2. Wait for health check to complete (~5-10 minutes)
kubectl wait --for=condition=complete --timeout=15m \
  job/operators-health-check -n openshift-gitops

# 3. Check instances start deploying
watch kubectl get applications -n openshift-gitops

# 4. Verify final state (all should be "Synced" and "Healthy")
kubectl get applications -n openshift-gitops
```

Expected timeline:
- **0-10 min:** Operators deploy (waves 10-18)
- **10-15 min:** Operators health check validates (wave 20)
- **15-22 min:** First instances deploy (waves 30-36: NFD, NVIDIA, LWS)
- **22-25 min:** Instances health check validates (wave 38)
- **25-30 min:** OpenShift AI DSC deploys (wave 42)
- **30-40 min:** Infrastructure deploys (wave 50)
- **40-45 min:** Applications deploy (waves 80-90: users, MAAS, models, GPU models, keycloak)
- **Total:** ~45-50 minutes for full deployment

---

## Emergency: Skip Health Check

**⚠️ Only if absolutely necessary:**

```bash
# Option 1: Change health check to wave 99
kubectl edit application operators-health-check -n openshift-gitops
# Change sync-wave from "20" to "99"

# Option 2: Remove from kustomization
# Edit clusters/ocpai3-aws/kustomization.yaml
# Comment out: # - operators-health-check-job.yaml

# Re-sync
argocd app sync maas-experiments-root-application
```

**Warning:** This removes the safety gate. Instances may fail if operators aren't ready.

