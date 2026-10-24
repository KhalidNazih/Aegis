package authorizer

import (
	"testing"

	securityv1alpha1 "github.com/KhalidNazih/aegis.git/api/v1alpha1"
)

const (
	testNamespace = "production"
	testWorkload  = "ci-cd-agent"
	testToolRead  = "github.issues.read"
)

func TestAuthorizeApprovedTool(t *testing.T) {
	policy := &securityv1alpha1.AIWorkloadPolicy{
		Spec: securityv1alpha1.AIWorkloadPolicySpec{
			TargetNameSpace: testNamespace,
			WorkloadName:    testWorkload,
			OperationalMode: "Enforce",
			Tools: []string{
				testToolRead,
			},
		},
	}
	req := AuthorizationRequest{
		WorkloadName:    testWorkload,
		TargetNamespace: testNamespace,
		RequestedTool:   testToolRead,
	}
	decision := Authorize(req, policy)
	if decision.Decision == false {
		t.Errorf("Expected true, got false! Reason: %s", decision.Reason)
	}

}
func TestAuthorizeBlockedTool(t *testing.T) {
	policy := &securityv1alpha1.AIWorkloadPolicy{
		Spec: securityv1alpha1.AIWorkloadPolicySpec{
			TargetNameSpace: testNamespace,
			WorkloadName:    testWorkload,
			OperationalMode: "Enforce",
			Tools: []string{
				testToolRead,
			},
		},
	}
	req := AuthorizationRequest{
		WorkloadName:    testWorkload,
		TargetNamespace: testNamespace,
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
			TargetNameSpace: testNamespace,
			WorkloadName:    testWorkload,
			OperationalMode: "Detect",
			Tools: []string{
				testToolRead,
			},
		},
	}
	req := AuthorizationRequest{
		WorkloadName:    testWorkload,
		TargetNamespace: testNamespace,
		RequestedTool:   "github.repository.delete",
	}
	decision := Authorize(req, policy)
	if decision.Decision == false {
		t.Errorf("Expected true, got false! Reason: %s", decision.Reason)
	}

}
