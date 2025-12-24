# MaaS Deployment Summary

This document provides a comprehensive overview of the Model-as-a-Service (MaaS) artifacts deployed via ArgoCD in this cluster environment.

## Table of Contents

- [Users](#users)
- [Groups](#groups)
- [Tier-to-Group Mapping](#tier-to-group-mapping)
- [Deployed Models](#deployed-models)
- [Rate Limiting Policies](#rate-limiting-policies)

---

## Users

The system defines both cluster admin users and standard MaaS users using htpasswd authentication.

### Cluster Admin Users

The following users have cluster-admin privileges:
- **bryonbaker**
- **bartoszmajsak**
- **nagarajan-ethan**
- **noyitz**
- **butler54**

### Standard Users

| Username | Identity Provider | Purpose |
|----------|------------------|---------|
| acme-user1 | local-htpasswd | Acme Inc test user |
| acme-user2 | local-htpasswd | Acme Inc test user |

---

## Groups

Groups are used to manage permissions and tier access for MaaS functionality.

### Group Membership Matrix

| Group Name | Members | Purpose |
|------------|---------|---------|
| **cluster-admins** | bryonbaker, bartoszmajsak, nagarajan-ethan, noyitz, butler54 | Cluster administration access |
| **maas-users** | bryonbaker, noyitz, bartoszmajsak, butler54, nagarajan-ethan, acme-user1, acme-user2 | Base MaaS access (view permissions + ingress config viewer) |
| **serverless-users** | bryonbaker, noyitz, bartoszmajsak, butler54, nagarajan-ethan, acme-user1, acme-user2 | Access to serverless models |
| **redhat-users** | bryonbaker, noyitz, bartoszmajsak, butler54 | Red Hat company users |
| **acme-inc-users** | acme-user1, acme-user2 | Acme Inc company users |
| **ethan-group-users** | nagarajan-ethan | Ethan Group company users |

Note: If you add a new user to the cluster, you will need to add them to the `maas-users` group so they can access the API.

### Group Permissions

#### maas-users
- **ClusterRole**: `view` - Read-only access to most resources in all namespaces
- **ClusterRole**: `ingress-config-viewer` - Custom role to get/list ingress configurations

---

## Tier-to-Group Mapping

Tiers define access levels and resource allocation for different user groups. The tier configuration is stored in a ConfigMap (`tier-to-group-mapping`) in the `maas-api` namespace.

| Tier Name | Level | Description | Associated Groups |
|-----------|-------|-------------|-------------------|
| **serverless** | 1 | Tier for provisioning all serverless models | serverless-users |
| **acme-inc-dedicated** | 50 | Tier for Acme Inc's dedicated models | acme-inc-users |
| **red-hat-dedicated** | 50 | Tier for Red Hat's dedicated models | red-hat-users |
| **ethan-group-dedicated** | 50 | Tier for Ethan Group's dedicated models | ethan-group-users |

Note: There is currently a bug that limits a user to only accessing models in one tier. (The highest priority tier.) This only impacts users that have more than one tier (obviously). I.e. It is okay to assign multiple tiers to users - just that they can only reach the models in highest level tier.

### Tier Level Explanation

- **Level 1**: Entry-level tier (serverless, shared resources)
- **Level 50**: Dedicated tier with higher priority and dedicated resources

---

## Deployed Models

The following LLM inference services are deployed across different namespaces:

### Model Deployment Details

| Model Name | Namespace | Model Path | Tier Association | Replicas | Image |
|------------|-----------|------------|------------------|----------|-------|
| **acme-inc-model-1** | acme-inc-models | facebook/opt-125m | acme-inc-dedicated | 1 | ghcr.io/llm-d/llm-d-inference-sim:v0.5.1 |
| **ethan-group-model-1** | ethan-group-models | facebook/opt-125m | ethan-group-dedicated | 1 | ghcr.io/llm-d/llm-d-inference-sim:v0.5.1 |
| **serverless-model-1** | serverless-models | facebook/opt-125m | serverless | 1 | ghcr.io/llm-d/llm-d-inference-sim:v0.5.1 |
| **serverless-model-2** | serverless-models | facebook/opt-125m | serverless | 1 | ghcr.io/llm-d/llm-d-inference-sim:v0.5.1 |

### Model Access by Group

#### serverless-users Group
Can access:
- serverless-model-1
- serverless-model-2

#### acme-inc-users Group
Can access:
- acme-inc-model-1

#### ethan-group-users Group
Can access:
- ethan-group-model-1

#### red-hat-users Group
Currently no dedicated models deployed (infrastructure ready)

### Model Configuration

All models share the following configuration:
- **Gateway**: maas-default-gateway (namespace: openshift-ingress)
- **Port**: 8000 (HTTPS)
- **Health Check**: `/health` endpoint
- **Readiness Check**: `/ready` endpoint
- **TLS**: Enabled with certificates mounted at `/var/run/kserve/tls/`
- **Mode**: Random inference simulation

---

## Rate Limiting Policies

Two types of rate limiting policies are configured at the gateway level to control resource usage.

### Request Rate Limiting

**Policy Name**: `gateway-rate-limits` (namespace: openshift-ingress)

| Tier | Limit | Window | Counter |
|------|-------|--------|---------|
| free | 5 requests | 2 minutes | per user ID |

- **Target**: maas-default-gateway
- **Tracking**: Automatic tracking via `auth.identity.tier` and `auth.identity.userid`

### Token Rate Limiting

**Policy Name**: `gateway-token-rate-limits` (namespace: openshift-ingress)

| Tier | Limit | Window | Counter |
|------|-------|--------|---------|
| free-user-tokens | 100 tokens | 1 minute | per user ID |

- **Target**: maas-default-gateway
- **Tracking**: Automatic tracking from response bodies (`usage.total_tokens`)
- **Predicate**: `auth.identity.tier == "free"`

### Important Notes

- Rate limit policies require resyncing after changes
- Token tracking automatically extracts usage from response bodies
- User groups align with authentication policy configuration

---

## Configuration Management

### MaaS Toolbox API

The deployment includes a REST API service (`maas-toolbox`) for simplified MaaS configuration management:

- **Purpose**: CRUD operations for tier and subscription policy management
- **Storage**: Kubernetes ConfigMaps
- **Namespace**: maas-api
- **ConfigMap**: tier-to-group-mapping

**Key Features**:
- Create, read, update, and delete tiers
- Manage group-to-tier associations
- Query tiers by group
- MaaS Tookbox base url: `https://maas-toolbox-maas-toolbox.apps.ethan-sno.sandbox3529.opentlc.com`
- Swagger documentation at `https://maas-toolbox-maas-toolbox.apps.ethan-sno.sandbox3529.opentlc.com/swagger/index.html`

### Rollout Requirements

After modifying the tier-to-group-mapping ConfigMap, restart the MaaS API deployment:

```bash
kubectl rollout restart deployment/maas-api -n maas-api
```

---

## Deployment Architecture

### Namespace Organization

| Namespace | Purpose | Resources |
|-----------|---------|-----------|
| maas-api | MaaS API and configuration | ConfigMaps, deployments, services |
| openshift-ingress | Gateway and rate limiting | Gateway, RateLimitPolicy, TokenRateLimitPolicy |
| acme-inc-models | Acme Inc model deployments | LLMInferenceService resources |
| ethan-group-models | Ethan Group model deployments | LLMInferenceService resources |
| serverless-models | Shared serverless model deployments | LLMInferenceService resources |

### ArgoCD Application Pattern

The deployment uses an App-of-Apps pattern with the following structure:
- **Root Application**: Manages cluster-level applications
- **Component Applications**: 
  - users-app: User and group definitions
  - maas-tier-config-app: Tier-to-group mappings
  - model-deployments-app: LLM inference services
  - machine-set-app: Machine set configurations for GPU nodes

---

## Summary Statistics

- **Total Users**: 7 (5 admin + 2 standard)
- **Total Groups**: 6
- **Total Tiers**: 4
- **Deployed Models**: 4
- **Model Namespaces**: 3
- **Active Rate Limit Policies**: 2

---

*Document generated: December 24, 2025*  
*Repository: maas-experiments*

