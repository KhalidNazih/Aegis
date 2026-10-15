<p align="center">
  <h1 align="center">AegisAI</h1>
  <p align="center">
    <strong>Kubernetes-Native Security Control Plane for AI Workloads, Autonomous Agents & Tool-Enabled LLMs</strong>
  </p>
  <p align="center">
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
    <a href="https://kubernetes.io"><img src="https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes&logoColor=white" alt="Kubernetes"></a>
    <a href="https://book.kubebuilder.io"><img src="https://img.shields.io/badge/Built%20with-Kubebuilder-000000?style=flat" alt="Kubebuilder"></a>
    <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
    <img src="https://img.shields.io/badge/Release-v0.1.0--alpha-orange" alt="Release">
    <a href="#contributing"><img src="https://img.shields.io/badge/PRs-Welcome-brightgreen.svg" alt="PRs Welcome"></a>
  </p>
</p>

---

## Overview

**AegisAI** is an open-source, Kubernetes-native security control plane engineered specifically for autonomous AI agents, tool-enabled assistants, RAG pipelines, and LLM inference workloads.

As AI workloads transition from static prompt-response scripts into autonomous agents running inside Kubernetes clusters, they cross infrastructure and application security boundaries:
* They make arbitrary tool calls (e.g., Model Context Protocol / MCP, internal APIs).
* They query external SaaS platforms (GitHub, Jira, Slack).
* They access sensitive cloud resources (S3, RDS, Secrets).
* They execute dynamic shell commands or generate code.

Standard Kubernetes primitives (**RBAC**, **NetworkPolicy**) only observe internal cluster calls or IP endpoints—they are blind to application-level AI tool operations. Cloud IAM and Secret managers have no context on agent autonomy.

> **AegisAI bridges this gap.** It introduces a unified, declarative **Security Contract** (`AIWorkloadPolicy`) that binds Kubernetes RBAC, egress NetworkPolicies, admission constraints, and tool authorization under a single, deny-by-default control plane.

---

## Core Capabilities

* **Declarative AI Security Contracts:** Define exactly what models, tools, cloud services, and network destinations an agent is authorized to access via a single Kubernetes CRD.
* **Deny-by-Default Tool Authorization:** Enforce strict capability boundaries for tool-enabled agents (MCP-compatible). If an action is not explicitly approved, it is denied.
* **Automated Egress & Network Containment:** Synthesizes and enforces strict network boundaries to prevent data exfiltration to unauthorized LLMs or API endpoints.
* **eBPF-Powered Runtime Telemetry:** Correlates low-level Linux process and network events back to specific AI workload identities.
* **GitOps-Centric Security Workflow:** Integrates seamlessly with GitOps tools like Argo CD. Policy adjustments require human engineer approval via standard pull requests—preventing unauthorized autonomous capability escalation.

---

## Architecture

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
  ├── RBAC Boundary Binding    ├── Tool Call Interception  ├── Tetragon eBPF
  ├── Egress NetworkPolicies   ├── Capability Evaluation   └── Audit Logging
  └── Kyverno Admission Rules  └── Deny-by-Default             │
                                                                 ▼
                                                       Security Events & Alerts
                                                       (GitOps Drift Detection)
```

### Core Design Principles
* **Reuse, Do Not Replace:** Works natively alongside Kyverno, NetworkPolicies, RBAC, and Tetragon.
* **Deny-by-Default:** Any tool, API, or network destination not explicitly approved in policy is blocked.
* **Zero Autonomous Escalation:** An AI agent cannot authorize itself. Authority changes require explicit engineer review via GitOps (Argo CD).
* **Explainable Decisions:** Every block or allow emits machine-readable and human-verifiable security decisions.

---

## Operational Modes

AegisAI supports four granular operational modes configured per policy:

| Mode | Purpose | Enforcement Action | Use Case |
| :--- | :--- | :--- | :--- |
| **`Detect`** | Discovery & Telemetry | Observe and log only | Auditing existing agents to discover needed permissions |
| **`Enforce`** | Active Protection | Blocks unauthorized tools & traffic | Production environments requiring least-privilege security |
| **`Review`** | Change Evaluation | Simulates policy evaluation | Testing new agent capabilities in staging |
| **`Remediate`** | Automated Drift Fix | Resets unauthorized cluster drift | Self-healing cluster security boundaries |

---

## The Security Contract: AIWorkloadPolicy

Here is an example of an `AIWorkloadPolicy` declaring approved capabilities for an autonomous CI/CD agent:

```yaml
apiVersion: security.aegisai.io/v1alpha1
kind: AIWorkloadPolicy
metadata:
  name: cicd-agent-policy
  namespace: aegis-system
spec:
  targetNameSpace: "production"
  workloadName: "ci-cd-agent"
  operationalMode: "Enforce"

  # Approved capabilities
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
        - "jira.internal.company"
```

---

## Getting Started

### Prerequisites
* Kubernetes cluster `v1.28+` (local Kind/Minikube or managed EKS/GKE/AKS)
* `kubectl` configured with cluster administrator access

---

### Option A: Deploy to a Kubernetes Cluster

You can deploy AegisAI directly into any Kubernetes cluster using standard manifests:

#### 1. Install the CRDs
```bash
kubectl apply -k https://github.com/KhalidNazih/aegisai.git/config/crd?ref=main
```

#### 2. Deploy the AegisAI Controller Manager
```bash
kubectl apply -k https://github.com/KhalidNazih/aegisai.git/config/default?ref=main
```

#### 3. Verify the Deployment
```bash
kubectl get pods -n aegis-system
```
You should see `aegisai-controller-manager` running and healthy.

#### 4. Apply Your First Policy
```bash
kubectl apply -f config/samples/security_v1alpha1_aiworkloadpolicy.yaml
```

Verify the policy status:
```bash
kubectl get aiworkloadpolicy -A
kubectl describe aiworkloadpolicy aiworkloadpolicy-sample
```

---

### Option B: Local Development & Contribution

If you are developing or contributing to AegisAI locally:

#### 1. Clone the Repository
```bash
git clone https://github.com/KhalidNazih/aegisai.git
cd aegisai/Kubebuilder
```

#### 2. Spin Up a Local Kind Cluster
```bash
kind create cluster --name aegis-dev
```

#### 3. Install CRDs & Run Locally
```bash
# Generate and install CRDs
make manifests
make install

# Run the controller against your local cluster context
make run
```

#### 4. Build and Push Container Images
```bash
make docker-build docker-push IMG=myregistry/aegisai-controller:latest
make deploy IMG=myregistry/aegisai-controller:latest
```

---

## Roadmap

- [x] **Core Operator Architecture:** Custom Resource Definition (`AIWorkloadPolicy`) and reconciliation engine.
- [x] **Status Condition Reporting:** Native Kubernetes condition framework (`Ready`, `Degraded`, `PolicyActive`).
- [ ] **Dynamic Policy Enforcement:** Automated generation of NetworkPolicies, RBAC boundaries, and Kyverno rules.
- [ ] **MCP & Tool Authorization Gateway:** Native interceptor for Model Context Protocol (MCP) tool execution.
- [ ] **Runtime Threat Telemetry:** eBPF-based event correlation and container containment via Tetragon.
- [ ] **GitOps & Drift Engine:** Automated detection of manual permission tampering and drift reporting via Argo CD.
- [ ] **Multi-Cloud Workload Identity:** Deep integration with AWS IAM Roles for Service Accounts (IRSA) and SPIFFE/SPIRE.

---

## Contributing

Contributions, issues, and feature requests are welcome.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/cool-feature`)
3. Commit your changes (`git commit -m 'feat: add cool feature'`)
4. Push to the branch (`git push origin feature/cool-feature`)
5. Open a Pull Request

---

## License

AegisAI is licensed under the [Apache 2.0 License](LICENSE).
