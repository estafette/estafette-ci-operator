package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	civ1 "github.com/estafette/estafette-ci-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	configMapOwnerKey = ".metadata.controller"
	apiGVStr          = civ1.GroupVersion.String()
)

// CredentialReconciler reconciles a Credential object
type CredentialReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
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

		configmap, err := r.constructConfigMapForCredential(&credential)
		if err != nil {
			log.Error(err, "unable to construct configmap from credential")
			// don't bother requeuing until we get a change to the spec
			return ctrl.Result{}, nil
		}
		if err := r.Create(ctx, configmap); err != nil {
			log.Error(err, "unable to create ConfigMap for Credential", "configmap", configmap)
			return ctrl.Result{}, err
		}

		log.V(1).Info("created ConfigMap for Credential run", "configmap", configmap)
	}

	return ctrl.Result{}, nil
}

func (r *CredentialReconciler) constructConfigMapForCredential(credential *civ1.Credential) (*corev1.ConfigMap, error) {
	name := fmt.Sprintf("%s-credential", credential.Name)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   credential.Namespace,
		},
		Data: make(map[string]string),
	}

	credentialsData := make(map[string]interface{})
	credentialConfig := make(map[string]interface{})
	credentialConfig["name"] = credential.ObjectMeta.Name
	credentialConfig["type"] = credential.Spec.Type
	credentialConfig["whitelistedTrustedImages"] = credential.Spec.WhitelistedTrustedImages
	credentialConfig["whitelistedPipelines"] = credential.Spec.WhitelistedPipelines
	for k, v := range credential.Spec.AdditionalProperties {
		credentialConfig[k] = v
	}

	credentialsData["credentials"] = []map[string]interface{}{credentialConfig}
	credentialsYaml, err := yaml.Marshal(credentialsData)
	if err != nil {
		return nil, err
	}
	cm.Data["credentials-config.yaml"] = string(credentialsYaml)

	for k, v := range credential.ObjectMeta.Labels {
		cm.ObjectMeta.Labels[k] = v
	}
	if err := SetControllerReferences(credential, cm, r.Scheme); err != nil {
		return nil, err
	}

	return cm, nil
}

func (r *CredentialReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(&corev1.ConfigMap{}, configMapOwnerKey, func(rawObj runtime.Object) []string {
		// grab the configmap object, extract the owner...
		configmap := rawObj.(*corev1.ConfigMap)
		owner := metav1.GetControllerOf(configmap)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Credential...
		if owner.APIVersion != apiGVStr || owner.Kind != "Credential" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&civ1.Credential{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
