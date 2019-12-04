package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	civ1 "github.com/estafette/estafette-ci-operator/api/v1"
)

// IntegrationReconciler reconciles a Integration object
type IntegrationReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=ci.estafette.io,resources=integrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ci.estafette.io,resources=integrations/status,verbs=get;update;patch

func (r *IntegrationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("integration", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *IntegrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&civ1.Integration{}).
		Complete(r)
}
