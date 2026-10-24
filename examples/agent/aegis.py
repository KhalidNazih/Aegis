"""
Aegis Python SDK
Kubernetes-native security control plane client for AI workloads and MCP agents.
"""

import os
import requests


class AegisSecurityError(Exception):
    """Raised when an AI tool execution is blocked by Aegis security policy."""
    pass


class AegisClient:
    """
    Client for the Aegis Tool Authorization Gateway.
    
    Usage in production:
        aegis = AegisClient(workload_name="ci-cd-agent", namespace="production")
        
        # In the agent's dynamic tool runner loop:
        aegis.authorize(dynamic_tool_name)
    """

    def __init__(
        self,
        workload_name: str,
        namespace: str = "production",
        gateway_url: str = None
    ):
        self.workload_name = workload_name
        self.namespace = namespace
        # Automatically detects Kubernetes Cluster DNS or falls back to localhost:
        self.gateway_url = gateway_url or os.getenv(
            "AEGIS_GATEWAY_URL",
            "http://localhost:9090/authorize"
        )

    def authorize(self, tool_name: str) -> bool:
        """
        Evaluates a tool call dynamically against the cluster's AIWorkloadPolicy.
        Returns True if approved. Raises AegisSecurityError if denied (HTTP 403).
        """
        payload = {
            "workloadName": self.workload_name,
            "targetNamespace": self.namespace,
            "requestedTool": tool_name
        }

        try:
            response = requests.post(self.gateway_url, json=payload, timeout=5)
        except requests.exceptions.ConnectionError:
            raise ConnectionError(
                f"[Aegis] Could not reach Aegis Gateway at '{self.gateway_url}'. "
                "Ensure the Aegis operator is running."
            )

        if response.status_code == 200:
            return True
        elif response.status_code == 403:
            data = response.json()
            reason = data.get("reason", "Action not approved by AIWorkloadPolicy")
            raise AegisSecurityError(f"[Aegis 403 Forbidden] {reason}")
        else:
            raise RuntimeError(
                f"[Aegis] Gateway returned unexpected HTTP status {response.status_code}: {response.text}"
            )

    def protect(self, func):
        """
        Python decorator to automatically guard any tool function with Aegis.
        
        @aegis.protect
        def my_tool(...):
            ...
        """
        def wrapper(*args, **kwargs):
            self.authorize(func.__name__)
            return func(*args, **kwargs)
        return wrapper
