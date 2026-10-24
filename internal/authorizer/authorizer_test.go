package authorizer

import (
	"testing"

	securityv1alpha1 "github.com/KhalidNazih/aegis.git/api/v1alpha1"
)

func TestAuthorizeApprovedTool(t *testing.T) {
	policy := &securityv1alpha1.AIWorkloadPolicy{
		Spec: securityv1alpha1.AIWorkloadPolicySpec{
			TargetNameSpace: "production",
			WorkloadName:    "ci-cd-agent",
			OperationalMode: "Enforce",
			Tools: []string{
				"github.issues.read",
			},
		},
	}
	req := AuthorizationRequest{
		WorkloadName:    "ci-cd-agent",
		TargetNamespace: "production",
		RequestedTool:   "github.issues.read",
	}
	decision := Authorize(req, policy)
	if decision.Decision == false {
		t.Errorf("Expected true, got false! Reason: %s", decision.Reason)
	}

}
func TestAuthorizeBlockedTool(t *testing.T) {
	policy := &securityv1alpha1.AIWorkloadPolicy{
		Spec: securityv1alpha1.AIWorkloadPolicySpec{
			TargetNameSpace: "production",
			WorkloadName:    "ci-cd-agent",
			OperationalMode: "Enforce",
			Tools: []string{
				"github.issues.read",
			},
		},
	}
	req := AuthorizationRequest{
		WorkloadName:    "ci-cd-agent",
		TargetNamespace: "production",
		RequestedTool:   "github.repository.delete",
	}
	decision := Authorize(req, policy)
	if decision.Decision == true {
		t.Errorf("Expected false, got true! Reason: %s", decision.Reason)
	}

}
func TestAuthorizeDetectMode(t *testing.T) {
	policy := &securityv1alpha1.AIWorkloadPolicy{
		Spec: securityv1alpha1.AIWorkloadPolicySpec{
			TargetNameSpace: "production",
			WorkloadName:    "ci-cd-agent",
			OperationalMode: "Detect",
			Tools: []string{
				"github.issues.read",
			},
		},
	}
	req := AuthorizationRequest{
		WorkloadName:    "ci-cd-agent",
		TargetNamespace: "production",
		RequestedTool:   "github.repository.delete",
	}
	decision := Authorize(req, policy)
	if decision.Decision == false {
		t.Errorf("Expected true, got false! Reason: %s", decision.Reason)
	}

}
