package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	civ1 "github.com/estafette/estafette-ci-operator/api/v1"
)

// CredentialReconciler reconciles a Credential object
type CredentialReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=ci.estafette.io,resources=credentials,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ci.estafette.io,resources=credentials/status,verbs=get;update;patch

func (r *CredentialReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("credential", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *CredentialReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&civ1.Credential{}).
		Complete(r)
}
