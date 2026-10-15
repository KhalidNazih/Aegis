<p align="center">
  <img src=".github/assets/logo.png" width="160" alt="AegisAI Logo">
  <h1 align="center">AegisAI</h1>
  <p align="center">
    <strong>Kubernetes-Native Security Control Plane for AI Workloads, Autonomous Agents & Tool-Enabled LLMs</strong>
  </p>
  <p align="center">
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
    <a href="https://kubernetes.io"><img src="https://img.shields.io/badge/Kubernetes-1.28+-326CE5?style=flat&logo=kubernetes&logoColor=white" alt="Kubernetes"></a>
    <a href="https://book.kubebuilder.io"><img src="https://img.shields.io/badge/Built%20with-Kubebuilder-000000?style=flat" alt="Kubebuilder"></a>
    <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
    <a href="https://github.com/KhalidNazih/aegisai/releases"><img src="https://img.shields.io/badge/Release-v1.0.0-success" alt="Release"></a>
    <a href="#contributing"><img src="https://img.shields.io/badge/PRs-Welcome-brightgreen.svg" alt="PRs Welcome"></a>
  </p>
</p>

---

## Overview

**AegisAI** is a Kubernetes-native security control plane engineered specifically for autonomous AI agents, tool-enabled assistants, RAG pipelines, and LLM inference workloads.

When autonomous agents run inside Kubernetes, traditional security controls fall short:
* **Kubernetes RBAC** only regulates internal cluster objects, leaving external tool operations and model APIs unmonitored.
* **NetworkPolicy** restricts IP/CIDR blocks, but cannot evaluate or filter application-level tool actions.
* **Cloud IAM** grants static permissions without visibility into dynamic agent prompt executions or autonomous tool calls.

AegisAI bridges this gap by introducing a unified **Security Contract** (`AIWorkloadPolicy`). It orchestrates Kubernetes RBAC, egress network boundaries, admission constraints, Model Context Protocol (MCP) tool gateways, and eBPF runtime telemetry under a single, deny-by-default control plane.

---

## Core Features

* 🛡️ **Unified Security Contract:** Declare the entire operational boundary of an AI workload in a single Kubernetes manifest: approved MCP tools, cloud APIs, Kubernetes verbs, secrets, and egress domains.
* 🔒 **Tool & MCP Authorization Gateway:** Real-time interception and filtering of Model Context Protocol (MCP) tool calls. Non-approved actions (such as unauthorized repository deletions or administrative API calls) are blocked before reaching target services.
* 🌐 **Automated Egress Synthesis:** Automatically generates and reconciles Kubernetes `NetworkPolicy` objects based on declared egress domains, restricting compromised or hallucinating agents from exfiltrating data to unapproved endpoints.
* ⚡ **eBPF Runtime Threat Telemetry:** Leverages Tetragon eBPF sensors to monitor container process executions, socket connections, and namespace escapes, correlating low-level Linux telemetry directly to specific AI workload identities.
* 🔄 **Continuous Drift Detection:** Continuously compares the declared policy against actual cluster configuration. Unapproved changes made out-of-band are flagged, alerted, and automatically remediated.

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

### Design Principles
* **Reuse, Do Not Replace:** Integrates directly with Kyverno, NetworkPolicies, RBAC, and Tetragon.
* **Deny-by-Default:** Any tool, cloud capability, or egress destination not explicitly declared in policy is blocked.
* **Zero Autonomous Escalation:** An AI agent cannot grant itself permissions. All capability updates require human engineer approval via GitOps (Argo CD).
* **Explainable Decisions:** Every allow, block, or containment action produces structured audit logs and Prometheus metrics.

---

## Operational Modes

Each `AIWorkloadPolicy` operates in one of four granular modes:

| Mode | Purpose | Enforcement Action | Use Case |
| :--- | :--- | :--- | :--- |
| **`Detect`** | Discovery & Telemetry | Observe and log only | Auditing existing agents to discover baseline requirements |
| **`Enforce`** | Active Protection | Blocks unauthorized tools & traffic | Production environments requiring least-privilege security |
| **`Review`** | Change Evaluation | Simulates policy evaluation | Staging environments evaluating new capabilities |
| **`Remediate`** | Automated Drift Fix | Resets unauthorized cluster drift | Self-healing cluster security boundaries |

---

## The Security Contract: AIWorkloadPolicy

Below is an example of an `AIWorkloadPolicy` declaring approved capabilities for an autonomous CI/CD agent:

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
    cloud:
      aws:
        - "s3:GetObject"
    secrets:
      - "github-token"
      - "jira-api-key"
    network:
      egress:
        - "api.github.com"
        - "jira.internal.company"
```

---

## Installation & Deployment

### Production Cluster Deployment

#### 1. Install CRDs and Controller via Kustomize
```bash
kubectl apply -k https://github.com/KhalidNazih/aegisai.git/config/default?ref=main
```

#### 2. Verify Operator Health
```bash
kubectl get pods -n aegis-system
```

Expected output:
```text
NAME                                         READY   STATUS    RESTARTS   AGE
aegisai-controller-manager-7d8b598b9-x2k4l   2/2     Running   0          45s
```

#### 3. Apply a Policy
```bash
kubectl apply -f config/samples/security_v1alpha1_aiworkloadpolicy.yaml
```

#### 4. Inspect Policy Status
```bash
kubectl get aiworkloadpolicy -A
```

```text
NAMESPACE      NAME                 MODE      READY   AGE
production     cicd-agent-policy    Enforce   True    2m
```

Detailed inspection shows the condition history:
```bash
kubectl describe aiworkloadpolicy cicd-agent-policy -n production
```

---

### Local Development (Kind)

#### 1. Clone and Setup
```bash
git clone https://github.com/KhalidNazih/aegisai.git
cd aegisai/Kubebuilder
```

#### 2. Create Cluster and Run Operator
```bash
kind create cluster --name aegis-dev
make manifests
make install
make run
```

#### 3. Build Container Images
```bash
make docker-build docker-push IMG=myregistry/aegisai-controller:v1.0.0
make deploy IMG=myregistry/aegisai-controller:v1.0.0
```

---

## Observability & Metrics

AegisAI exports Prometheus metrics on port `:8080/metrics`:

| Metric Name | Type | Description |
| :--- | :--- | :--- |
| `aegisai_policy_reconcile_total` | Counter | Total reconciliation loops executed per policy |
| `aegisai_tool_call_blocked_total` | Counter | Total unauthorized tool actions blocked |
| `aegisai_network_violation_total` | Counter | Total blocked egress connection attempts |
| `aegisai_drift_detected_total` | Counter | Total configuration drift occurrences detected |

Grafana dashboard templates are located in the `dashboards/` directory.

---

## Contributing

Contributions, issues, and feature requests are welcome.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m 'feat: add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

---

## License

AegisAI is open-source software licensed under the [Apache 2.0 License](LICENSE).
