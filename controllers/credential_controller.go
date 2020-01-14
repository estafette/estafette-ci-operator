package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	civ1 "github.com/estafette/estafette-ci-operator/api/v1"
)

var (
	configMapOwnerKey = ".metadata.controller"
	apiGVStr          = civ1.GroupVersion.String()
)

// CredentialReconciler reconciles a Credential object
type CredentialReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=ci.estafette.io,resources=credentials,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ci.estafette.io,resources=credentials/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get

func (r *CredentialReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("credential", req.NamespacedName)

	var credential civ1.Credential
	if err := r.Get(ctx, req.NamespacedName, &credential); err != nil {
		log.Error(err, "unable to fetch Credential")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var childConfigMaps corev1.ConfigMapList
	if err := r.List(ctx, &childConfigMaps, client.InNamespace(req.Namespace), client.MatchingFields{configMapOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child ConfigMaps")
		return ctrl.Result{}, err
	}
	if len(childConfigMaps.Items) <= 0 {
		// TODO: create new config map here
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *CredentialReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&civ1.Credential{}).
		Complete(r)
}
