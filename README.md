<p align="center">
  <h1 align="center">🛡️ AegisAI</h1>
  <p align="center">
    <strong>Kubernetes-Native Security Control Plane for AI Workloads, Autonomous Agents & Tool-Enabled LLMs</strong>
  </p>
  <p align="center">
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
    <a href="https://kubernetes.io"><img src="https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes&logoColor=white" alt="Kubernetes"></a>
    <a href="https://book.kubebuilder.io"><img src="https://img.shields.io/badge/Built%20with-Kubebuilder-000000?style=flat" alt="Kubebuilder"></a>
    <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
    <img src="https://img.shields.io/badge/Status-Alpha%20%2F%20In%20Active%20Development-orange" alt="Status">
    <a href="#contributing"><img src="https://img.shields.io/badge/PRs-Welcome-brightgreen.svg" alt="PRs Welcome"></a>
  </p>
</p>

---

## 📌 Overview

**AegisAI** is a Kubernetes-native security control plane engineered specifically for AI workloads—including autonomous agents, tool-enabled assistants, RAG pipelines, and inference services.

Traditional Kubernetes security controls operate in domain silos:
* **Kubernetes RBAC** only understands internal API objects (Pods, Secrets, Deployments).
* **NetworkPolicy** restricts IP and port reachability, but is blind to external API actions.
* **Cloud IAM** manages cloud services, but has no visibility into what an agent prompts or calls.
* **Runtime security** (eBPF) sees Linux syscalls, but lacks AI context.

When an autonomous AI agent running in a Pod needs to query GitHub, read Jira, query vector databases, and trigger cloud APIs, **no single control plane has end-to-end visibility**.

> **AegisAI bridges this gap.** It introduces a unified **Security Contract** for AI workloads, binding existing security primitives (RBAC, NetworkPolicy, Kyverno, MCP Gateways, and Tetragon) under a single declarative custom resource: `AIWorkloadPolicy`.

---

## 🏛️ High-Level Architecture

AegisAI reuses and orchestrates battle-tested security primitives rather than reinventing them:

```text
                           AIWorkloadPolicy (CRD)
                                     │
                                     ▼
                        ┌────────────────────────┐
                        │   AegisAI Controller   │
                        └────────────┬───────────┘
                                     │
         ┌───────────────────────────┼───────────────────────────┐
         ▼                           ▼                           ▼
  Kubernetes Layer              Tool/MCP Layer              Runtime Layer
  ├── RBAC Constraints         ├── MCP Gateway Sync        ├── Tetragon eBPF
  ├── Egress NetworkPolicies   ├── Action Filtering        └── Audit Logging
  └── Kyverno Admission Rules  └── Deny-by-Default             │
                                                                 ▼
                                                       Security Events & Alerts
                                                       (GitOps Drift Detection)
```

### Core Design Principles
* **Reuse, Don't Replace:** Works natively alongside Kyverno, NetworkPolicies, RBAC, and Tetragon.
* **Deny-by-Default:** Any tool, API, or network destination not explicitly approved in policy is blocked.
* **Zero Autonomous Escalation:** An AI agent cannot authorize itself. Authority changes require explicit engineer review via GitOps (Argo CD).
* **Explainable Decisions:** Every block or allow emits machine-readable and human-verifiable security decisions.

---

## 📄 The Security Contract (`AIWorkloadPolicy`)

A declarative specification declaring what an AI workload is authorized to execute across infrastructure and external tools:

```yaml
apiVersion: security.aegisai.io/v1alpha1
kind: AIWorkloadPolicy
metadata:
  name: ci-cd-agent-policy
  namespace: default
spec:
  targetNameSpace: "production"
  workloadName: "ci-cd-agent"
  operationalMode: "Enforce" # Detect | Enforce | Review | Remediate
  
  # Approved capabilities (Stage 2 schema preview)
  capabilities:
    tools:
      github:
        - "issues.read"
        - "pull_requests.create"
      jira:
        - "issues.read"
    kubernetes:
      - "pods.get"
      - "pods.list"
    network:
      egress:
        - "api.github.com"
        - "jira.company.internal"
```

---

## 🚀 Quickstart & Local Development

### Prerequisites
* [Docker](https://docs.docker.com/get-docker/) (v20.10+)
* [Go](https://go.dev/) (v1.24+)
* [Kind](https://kind.sigs.k8s.io/) (v0.20+)
* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Kubebuilder](https://book.kubebuilder.io/quick-start.html) (v4.0+)

### 1. Clone the Repository
```bash
git clone https://github.com/KhalidNazih/aegisai.git
cd aegisai/Kubebuilder
```

### 2. Create a Local Kind Cluster
```bash
kind create cluster --name aegis-dev
```

### 3. Install CRDs into the Cluster
```bash
make manifests
make install
```

### 4. Run the Operator Locally
```bash
make run
```

### 5. Apply a Sample AI Workload Policy (In another terminal)
```bash
kubectl apply -f config/samples/security_v1alpha1_aiworkloadpolicy.yaml
```

Inspect the status condition verified by the AegisAI controller:
```bash
kubectl describe aiworkloadpolicy aiworkloadpolicy-sample
```

---

## 🗺️ Roadmap & Milestones

- [x] **Stage 1: Local Kubernetes Operator Foundation**
  - Scaffolding via Kubebuilder & Kind local lab.
  - Initial `AIWorkloadPolicy` Custom Resource Definition.
  - Active controller reconciliation loop & Kubernetes Status Conditions.
- [ ] **Stage 2: Comprehensive Policy Schema**
  - Fine-grained capability modeling (Tools, Cloud IAM, K8s RBAC, Network egress).
  - OpenAPIV3 schema validation & operational modes (`Detect`, `Enforce`, `Review`).
- [ ] **Stage 3: Kubernetes Enforcement Adapters**
  - Dynamic NetworkPolicy generation for egress restrictions.
  - Kyverno admission policy generation & RBAC boundary synthesis.
- [ ] **Stage 4: Tool & MCP Authorization Gateway**
  - Integration with Model Context Protocol (MCP) tool calls.
  - Deny-by-default capability evaluation & structured decision events.
- [ ] **Stage 5: Runtime Security with Tetragon (eBPF)**
  - Correlating real Linux process & network telemetry with AI workload identity.
  - Anomalous behavior detection and automated containment.
- [ ] **Stage 6: GitOps Lifecycle Integration**
  - Argo CD continuous reconciliation & Git-based security reviews.
- [ ] **Stage 7: Local & Cloud Drift Detection**
- [ ] **Stage 8: Cloud Identity & Secrets (AWS IAM, OpenBao/Vault)**
- [ ] **Stage 9: Multi-Tenant Hardening**
- [ ] **Stage 10: Production Hardening (Webhooks, HA, Metrics)**

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
Feel free to check [issues page](https://github.com/KhalidNazih/aegisai/issues).

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📜 License

Distributed under the **Apache-2.0 License**. See `LICENSE` for more information.
