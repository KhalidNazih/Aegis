package observer

import (
	"fmt"
	"time"

	securityv1alpha1 "github.com/KhalidNazih/aegis.git/api/v1alpha1"
	"k8s.io/client-go/tools/record"
)

type IncidentRecord struct {
	WorkloadName   string
	Namespace      string
	PolicyName     string
	ObservedAction string
	Result         string
	Severity       string
	Timestamp      time.Time
}

func (i *IncidentRecord) String() string {
	return fmt.Sprintf(`
--------------------------------------------------
    [AEGIS SECURITY INCIDENT RECORD]
    Workload: %s (Namespace: %s)
    Policy:   %s
    Observed: %s
    Result:   %s
    Severity: %s
--------------------------------------------------`,
		i.WorkloadName,
		i.Namespace,
		i.PolicyName,
		i.ObservedAction,
		i.Result,
		i.Severity,
	)
}
func EmitIncident(recorder record.EventRecorder, policy *securityv1alpha1.AIWorkloadPolicy, incident IncidentRecord) {
	fmt.Println(incident.String())
	if recorder != nil {
		recorder.Event(policy, "Warning", "RuntimeAttackBlocked", incident.ObservedAction)
	}
}
