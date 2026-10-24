/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	securityv1alpha1 "github.com/KhalidNazih/aegis.git/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AIWorkloadPolicyReconciler reconciles a AIWorkloadPolicy object
type AIWorkloadPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *AIWorkloadPolicyReconciler) builNetworkPolicy(policy *securityv1alpha1.AIWorkloadPolicy) *networkingv1.NetworkPolicy {
	dnsPort := intstr.FromInt(53)
	httpsPort := intstr.FromInt(443)
	tcp := corev1.ProtocolTCP
	udp := corev1.ProtocolUDP
	var egressRules []networkingv1.NetworkPolicyEgressRule
	if len(policy.Spec.NetworkEgress) > 0 {
		egressRules = []networkingv1.NetworkPolicyEgressRule{
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: &udp, Port: &dnsPort},
					{Protocol: &tcp, Port: &dnsPort},
				},
			},
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: &tcp, Port: &httpsPort},
				},
			},
		}
	}
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policy.Spec.WorkloadName + "-network-guardd",
			Namespace: policy.Spec.TargetNameSpace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": policy.Spec.WorkloadName,
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
			},
			Egress: egressRules,
		},
	}
}

// +kubebuilder:rbac:groups=security.aegis.io,resources=aiworkloadpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.aegis.io,resources=aiworkloadpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.aegis.io,resources=aiworkloadpolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *AIWorkloadPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var policy securityv1alpha1.AIWorkloadPolicy
	err := r.Get(ctx, req.NamespacedName, &policy) // read what we wrote in our yaml
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	fmt.Printf("AegisAI detected policy for AI workload %s in mode %s \n", policy.Spec.TargetNameSpace, policy.Spec.OperationalMode)

	//NetworkPolicy
	desiredNetPol := r.builNetworkPolicy(&policy) //holds the the name the namespace and the egress rule
	netPolKey := types.NamespacedName{
		Name:      desiredNetPol.Name,
		Namespace: desiredNetPol.Namespace,
	}
	var existingNetPol networkingv1.NetworkPolicy
	err = r.Get(ctx, netPolKey, &existingNetPol)
	if errors.IsNotFound(err) { //Network policy does not exist
		if err := r.Create(ctx, desiredNetPol); err != nil { //The API server is unreachable or timed out!
			return ctrl.Result{}, err
		}
		fmt.Printf("Aegis successfully enforced NetworkPolicy: %s in namespace %s\n", desiredNetPol.Name, desiredNetPol.Namespace)
	} else if err != nil { // Cluster crashed or connection broke during r.Get!.
		return ctrl.Result{}, err
	} else {
		existingNetPol.Spec.Egress = desiredNetPol.Spec.Egress //If it already exists, update its egress rules to match the latest desired state
		if err := r.Update(ctx, &existingNetPol); err != nil {
			return ctrl.Result{}, err
		}
	}

	//Update status Condition after reforcing netowrkpolicy
	meta.SetStatusCondition(&policy.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "NetworkEnforced",
		Message:            "NetworkPolicy successfully enforced for workload",
		ObservedGeneration: policy.Generation,
	})

	if err := r.Status().Update(ctx, &policy); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AIWorkloadPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.AIWorkloadPolicy{}).
		Owns(&networkingv1.NetworkPolicy{}). // add this also if sm1 deletes the networkpolicy
		Named("aiworkloadpolicy").
		Complete(r)
}
