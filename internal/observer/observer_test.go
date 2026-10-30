package observer

import (
	"strings"
	"testing"
	"time"

	securityv1alpha1 "github.com/KhalidNazih/aegis.git/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

const (
	testNamespace  = "production"
	testPolicyName = "aiworkloadpolicy-sample"
)

func TestIncidentRecordString(t *testing.T) {
	incident := IncidentRecord{
		WorkloadName:   "ci-cd-agent",
		Namespace:      testNamespace,
		PolicyName:     testPolicyName,
		ObservedAction: "Unauthorized execution of /bin/sh",
		Result:         "BLOCKED",
		Severity:       "HIGH",
		Timestamp:      time.Now(),
	}

	formatted := incident.String()

	expectedSubstrings := []string{
		"Workload: ci-cd-agent (Namespace: production)",
		"Policy:   aiworkloadpolicy-sample",
		"Observed: Unauthorized execution of /bin/sh",
		"Result:   BLOCKED",
		"Severity: HIGH",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(formatted, expected) {
			t.Errorf("Expected incident string to contain %q, but it did not.\nFormatted output:\n%s", expected, formatted)
		}
	}
}

func TestEmitIncidentWithRecorder(t *testing.T) {
	fakeRecorder := record.NewFakeRecorder(10)
	policy := &securityv1alpha1.AIWorkloadPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPolicyName,
			Namespace: testNamespace,
		},
	}

	incident := IncidentRecord{
		WorkloadName:   "ci-cd-agent",
		Namespace:      testNamespace,
		PolicyName:     testPolicyName,
		ObservedAction: "Unauthorized execution of /bin/sh",
		Result:         "BLOCKED",
		Severity:       "HIGH",
		Timestamp:      time.Now(),
	}

	EmitIncident(fakeRecorder, policy, incident)

	select {
	case event := <-fakeRecorder.Events:
		if !strings.Contains(event, "RuntimeAttackBlocked") {
			t.Errorf("Expected event reason RuntimeAttackBlocked, got: %s", event)
		}
		if !strings.Contains(event, "Unauthorized execution of /bin/sh") {
			t.Errorf("Expected event message to mention /bin/sh, got: %s", event)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for event on fake recorder")
	}
}
