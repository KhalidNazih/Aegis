package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	securityv1alpha1 "github.com/KhalidNazih/aegis.git/api/v1alpha1"
	"k8s.io/client-go/tools/record"

	"github.com/KhalidNazih/aegis.git/internal/authorizer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Server struct {
	Client   client.Client
	Recorder record.EventRecorder
}

func (s *Server) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var authReq authorizer.AuthorizationRequest
	err := json.NewDecoder(r.Body).Decode(&authReq)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	var policyList securityv1alpha1.AIWorkloadPolicyList
	if err := s.Client.List(r.Context(), &policyList); err != nil {
		http.Error(w, "Failed to fetch policies", http.StatusInternalServerError)
		return
	}
	var targetPolicy *securityv1alpha1.AIWorkloadPolicy
	for i := range policyList.Items {
		p := &policyList.Items[i]
		if p.Spec.WorkloadName == authReq.WorkloadName && p.Spec.TargetNameSpace == authReq.TargetNamespace {
			targetPolicy = p
			break
		}
	}
	if targetPolicy == nil {
		http.Error(w, "No security policy found for this workload", http.StatusForbidden)
		return
	}
	decision := authorizer.Authorize(authReq, targetPolicy)
	w.Header().Set("Content-Type", "application/json")

	if !decision.Decision {
		if s.Recorder != nil {
			s.Recorder.Event(targetPolicy, "Warning", "UnauthorizedToolCall", decision.Reason)
		}
		w.WriteHeader(http.StatusForbidden)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/authorize", s.HandleAuthorize)
	return http.ListenAndServe(":9090", mux)
}
