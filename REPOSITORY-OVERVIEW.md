# MaaS Experiments Repository - High-Level Overview

## Purpose

This repository provides a complete reference implementation for deploying and managing Model-as-a-Service (MaaS) infrastructure on OpenShift/Kubernetes clusters. It demonstrates production-grade patterns for multi-tenant LLM deployments with tier-based access control, rate limiting, and automated GitOps delivery via ArgoCD.

## Target Audience

- **Platform Engineers**: Building private/sovereign cloud AI platforms
- **Cloud Providers**: Implementing MaaS for customer consumption
- **DevOps Teams**: Deploying and managing LLM inference infrastructure
- **Enterprise Architects**: Designing multi-tenant AI service platforms

---

## Repository Structure

### 1. **clusters/** - Cluster-Specific Configurations

Contains ArgoCD application definitions and cluster overlays using the App-of-Apps pattern.

- **ocpai3-aws/**: Example cluster configuration for AWS deployment
  - Operator installations (NFD, NVIDIA GPU, Kueue, LeaderWorkerSet, cert-manager, OCPAI)
  - Instance configurations (GPU cluster policies, node feature discovery)
  - Application deployments (Keycloak, model deployments, machine sets)
  - User and tier configurations
- **root-application.yaml**: Top-level ArgoCD application that bootstraps all other apps
- **app-of-apps.yaml**: ArgoCD App-of-Apps pattern implementation
- **appofapps-repository-config.yaml**: Repository configuration for ArgoCD

### 2. **components/** - Reusable Kustomize Bases

Modular, reusable Kubernetes manifests organized by function.

#### **components/operators/** - Operator Deployments

OpenShift operator subscriptions with proper namespace and operator group configuration:
- **cert-manager**: Certificate management for TLS
- **kueue**: Job queuing and resource management
- **leader-worker-set**: Distributed workload management
- **nfd** (Node Feature Discovery): Hardware capability detection
- **nvidia**: GPU operator for NVIDIA hardware support
- **ocpai_3_0**: OpenShift AI 3.0 operator

#### **components/apps/** - Application Deployments

- **keycloak**: Identity and access management (base + overlays)
  - Admin secrets
  - Deployment configurations

#### **components/platform/** - Platform Infrastructure

Core MaaS platform components:

- **gpu-model/**: GPU model definitions for LLM inference
- **lws-instance/**: LeaderWorkerSet instance configurations
- **maas/**: Core MaaS configuration
  - Tier-to-group mappings
  - Request rate limit policies
  - Token rate limit policies
- **machine-set/**: MachineSet definitions for dynamic node provisioning
  - 4xlarge instances
  - g6.8xlarge GPU instances
- **model-deployments/**: LLM inference service definitions
  - 4 model deployments across 3 namespaces
  - Tier-based access control
- **monitoring/**: Monitoring and observability configuration
- **nfd-instance/**: Node Feature Discovery instance
- **nvidia-instance/**: NVIDIA GPU cluster policy
- **oauth/**: OAuth and user management
  - Cluster role bindings
  - User and group definitions
  - htpasswd authentication
- **ocpai/**: OpenShift AI configuration
- **users/**: User authentication and authorization
  - htpasswd secrets
  - User and group definitions

### 3. **source/** - Supporting Applications

#### **source/client-example/**

Example client implementations for interacting with deployed models:

- **curl-cheatsheet/**: Shell script examples for API testing
- **golang-client/**: Go-based client application
  - OpenShift resource listing
  - Model interaction examples
  - Internal packages for API communication

#### **source/maas-toolbox/**

A production-ready REST API service for MaaS configuration management:

**Features**:
- CRUD operations for tier management
- Group-to-tier association management
- ConfigMap-based storage (Kubernetes-native)
- Swagger/OpenAPI documentation
- Health check endpoints

**Architecture**:
- Clean separation: models, storage, service, API layers
- Kubernetes ConfigMap persistence
- Gin web framework
- Comprehensive test suite

**API Capabilities**:
- Create/list/update/delete tiers
- Add/remove groups from tiers
- Query tiers by group
- Validates business rules (tier uniqueness, immutable names)

### 4. **tools/** - Automation Scripts

Helper scripts for cluster bootstrapping and testing:
- **bootstrap-cluster.sh**: Automated ArgoCD deployment and configuration
- **test-inference.sh**: Model inference testing script

### 5. **images/** - Documentation Assets

Visual documentation and diagrams:
- **maas-domain-model.png**: High-level architecture diagram showing MaaS domain model

---

## Key Concepts

### Multi-Tenant Architecture

The repository implements a complete multi-tenant LLM serving platform:

1. **Tier-Based Access Control**: Users grouped into tiers (serverless, dedicated)
2. **Resource Isolation**: Separate namespaces per organization/tier
3. **Rate Limiting**: Request and token-based limits per tier
4. **Gateway Pattern**: Centralized ingress with unified policy enforcement

### GitOps Workflow

- **Declarative Configuration**: All infrastructure as code
- **ArgoCD App-of-Apps**: Hierarchical application management
- **Environment Promotion**: Easy replication across clusters
- **Audit Trail**: Git history provides complete change tracking

### Operator-Based Deployment

Leverages Kubernetes operators for lifecycle management:
- GPU drivers and device plugins (NVIDIA operator)
- Node hardware discovery (NFD)
- AI/ML workload orchestration (OpenShift AI)
- Job queuing and prioritization (Kueue)

---

## Technology Stack

### Core Infrastructure
- **OpenShift/Kubernetes**: Container orchestration platform
- **ArgoCD**: GitOps continuous delivery
- **Kustomize**: Configuration management and templating

### AI/ML Stack
- **OpenShift AI 3.0**: Enterprise AI/ML platform
- **KServe**: Model serving and inference
- **LeaderWorkerSet**: Distributed training/inference
- **NVIDIA GPU Operator**: GPU resource management

### Observability & Management
- **Node Feature Discovery**: Hardware capability detection
- **Kueue**: Job queuing and resource quotas
- **Kuadrant**: Gateway API and policy enforcement

### Security & Identity
- **Keycloak**: Identity provider (optional)
- **htpasswd**: Local authentication
- **OAuth**: OpenShift integrated authentication
- **cert-manager**: Certificate lifecycle management

---

## Deployment Model

### Tier Structure

| Tier Level | Type | Resource Model | Use Case |
|------------|------|----------------|----------|
| Level 1 | Serverless | Shared infrastructure | Basic users, low-volume workloads |
| Level 50 | Dedicated | Isolated resources | Enterprise customers, high-volume |

### User Journey

1. **Authentication**: Users authenticate via OAuth/htpasswd
2. **Authorization**: Group membership determines tier access
3. **Model Discovery**: Users query available models via API
4. **Inference**: Requests routed through centralized gateway
5. **Rate Limiting**: Enforced at gateway based on tier
6. **Billing/Metering**: Token usage tracked for chargeback

---

## Use Cases

### 1. Private Cloud LLM Platform
Deploy a secure, on-premises LLM serving platform with full data sovereignty.

### 2. Multi-Customer AI SaaS
Build a multi-tenant AI platform with tier-based pricing and resource isolation.

### 3. Enterprise AI Infrastructure
Provide self-service AI capabilities to internal teams with governance and cost control.

### 4. Development/Testing Environment
Rapidly spin up complete MaaS environments for testing and experimentation.

---

## Key Features

### ✅ Production-Ready Patterns
- Operator-based lifecycle management
- GitOps continuous delivery
- Multi-tenant isolation
- Comprehensive RBAC

### ✅ Scalability
- Dynamic node provisioning with MachineSets
- GPU resource management
- Horizontal pod autoscaling ready
- Queue-based workload management

### ✅ Security & Governance
- Tier-based access control
- Request and token rate limiting
- TLS encryption
- Audit logging via Git

### ✅ Developer Experience
- REST API for configuration management
- Comprehensive examples (curl, Go client)
- Interactive Swagger documentation
- Automated testing scripts

### ✅ Operational Excellence
- Infrastructure as Code
- Automated bootstrapping
- ConfigMap-based configuration
- Health checks and monitoring

---

## Getting Started

### Prerequisites
- OpenShift 4.x or Kubernetes 1.24+ cluster
- Cluster-admin access
- GPU nodes (for model serving)
- `kubectl` or `oc` CLI
- ArgoCD (or install via bootstrap script)

### Quick Start

1. **Bootstrap the cluster**:
   ```bash
   ./tools/bootstrap-cluster.sh
   ```

2. **Deploy base configuration**:
   ```bash
   kubectl apply -k components/platform/maas/base
   kubectl apply -k components/platform/users/base
   ```

3. **Deploy models**:
   ```bash
   kubectl apply -k components/apps/model-deployments/base
   ```

4. **Test inference**:
   ```bash
   ./tools/test-inference.sh
   ```

### ArgoCD Deployment

1. Configure ArgoCD repository connection
2. Apply root application:
   ```bash
   kubectl apply -f clusters/root-application.yaml
   ```
3. ArgoCD will automatically deploy all child applications

---

## Configuration Management

### Tier Management

**Declarative** (GitOps):
- Edit `components/platform/maas/base/tier-to-group-mapping.yaml`
- Commit and push
- ArgoCD syncs automatically

**Imperative** (API):
- Use MaaS Toolbox REST API
- Interactive changes via Swagger UI
- ConfigMap updated automatically

### User Management

1. Add users to `components/platform/users/files/users.htpasswd`
2. Update group membership in `components/platform/users/base/users-and-groups.yaml`
3. Commit and sync

### Model Deployment

1. Create new LLMInferenceService YAML in `components/apps/model-deployments/base/`
2. Add tier annotation: `alpha.maas.opendatahub.io/tiers`
3. Update kustomization.yaml
4. Commit and sync

---

## Repository Highlights

### 🎯 Well-Organized
Clear separation of concerns: operators, apps, platform, cluster overlays

### 📚 Documented
Comprehensive README files, inline comments, example scripts

### 🔧 Modular
Reusable kustomize bases, easy to adapt for different environments

### 🚀 Production-Ready
Follows OpenShift/Kubernetes best practices, operator-based lifecycle

### 🧪 Testable
Includes test scripts, example clients, validation tools

### 🔄 GitOps-Native
Designed for ArgoCD App-of-Apps pattern from the ground up

---

## Future Enhancements

- **Enhanced Monitoring**: Prometheus metrics, Grafana dashboards
- **Multi-Cluster**: Federation across regions/clouds
- **Advanced Scheduling**: GPU time-slicing, multi-instance GPU (MIG)
- **Cost Management**: Chargeback/showback integration
- **Model Registry**: Centralized model catalog and versioning
- **Autoscaling**: Cluster autoscaler integration, KEDA for event-driven scaling

---

## Summary

This repository provides a **complete, production-ready reference architecture** for deploying Model-as-a-Service on OpenShift/Kubernetes. It combines:

- **Enterprise-grade security** with multi-tenant isolation and RBAC
- **Operational excellence** through GitOps and operator-based management
- **Developer productivity** with REST APIs and comprehensive tooling
- **Scalability** via dynamic provisioning and GPU resource management
- **Flexibility** through modular, kustomize-based configuration

Whether you're building a private AI cloud, a multi-customer SaaS platform, or enterprise AI infrastructure, this repository provides the foundational patterns and components to get started quickly.

---

## License

Apache License 2.0 (for maas-toolbox component)

---

*Document generated: December 24, 2025*  
*Repository: https://github.com/[organization]/maas-experiments*

