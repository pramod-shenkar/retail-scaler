/*
Copyright 2025.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "retail-scalar/api/v1"
)

// OrgReconciler reconciles a Org object
type OrgReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.vajra.com,resources=orgs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.vajra.com,resources=orgs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.vajra.com,resources=orgs/finalizers,verbs=update

func (r *OrgReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// check current instance deployment holding

	// match with expected

	// if not found as expected take action
	return ctrl.Result{}, nil
}

func (r *OrgReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Org{}).
		Named("org").
		Complete(r)
}
