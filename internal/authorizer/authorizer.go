package authorizer

import (
	"fmt"
	"slices"

	securityv1alpha1 "github.com/KhalidNazih/aegis.git/api/v1alpha1"
)

type AuthorizationRequest struct {
	WorkloadName    string `json:"workloadName"`
	TargetNamespace string `json:"targetNamespace"`
	RequestedTool   string `json:"requestedTool"`
}
type AuthorizationDecision struct {
	Decision     bool   `json:"decision"`
	Reason       string `json:"reason"`
	WorkloadName string `json:"workloadName"`
	ToolName     string `json:"toolName"`
}

func Authorize(req AuthorizationRequest, policy *securityv1alpha1.AIWorkloadPolicy) AuthorizationDecision {
	if policy == nil {
		return AuthorizationDecision{
			Decision:     false,
			Reason:       "Policy is nil or not found",
			WorkloadName: req.WorkloadName,
			ToolName:     req.RequestedTool}
	} else if policy.Spec.OperationalMode == "Detect" {

		return AuthorizationDecision{
			Decision:     true,
			Reason:       fmt.Sprintf("Allowed under Detect mode: observation recorded for tool %s", req.RequestedTool),
			WorkloadName: req.WorkloadName,
			ToolName:     req.RequestedTool}
	} else {
		if slices.Contains(policy.Spec.Tools, req.RequestedTool) {
			return AuthorizationDecision{
				Decision:     true,
				Reason:       fmt.Sprintf("Tool %s is explicitly approved by policy", req.RequestedTool),
				WorkloadName: req.WorkloadName,
				ToolName:     req.RequestedTool}
		}
		return AuthorizationDecision{
			Decision: false,
			Reason: fmt.Sprintf("Tool %s is DENIED: not approved in AIWorkloadPolicy for workload %s",
				req.RequestedTool, req.WorkloadName),

			WorkloadName: req.WorkloadName,
			ToolName:     req.RequestedTool,
		}
	}
}
