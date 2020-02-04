package controllers

import (
	"context"
	"fmt"
	civ1 "github.com/estafette/estafette-ci-operator/api/v1"
	"github.com/go-logr/logr"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	configMapOwnerKey = ".metadata.controller"
	apiGVStr          = civ1.GroupVersion.String()
)

// CredentialReconciler reconciles a Credential object
type CredentialReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=ci.estafette.io,resources=credentials,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ci.estafette.io,resources=credentials/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;patch;delete

func (r *CredentialReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("credential", req.NamespacedName)
	log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))
	defer log.Info(fmt.Sprintf("Finish reconcile loop for %v", req.NamespacedName))

	var credential civ1.Credential
	if err := r.Get(ctx, req.NamespacedName, &credential); err != nil {
		log.Error(err, "unable to fetch Credential")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var childConfigMaps corev1.ConfigMapList
	if err := r.List(ctx, &childConfigMaps, client.InNamespace(req.Namespace), client.MatchingFields{configMapOwnerKey: "Credential"}); err != nil {
		log.Error(err, "unable to list child ConfigMaps")
		return ctrl.Result{}, err
	}
	if len(childConfigMaps.Items) > 0 {
		if len(childConfigMaps.Items) > 1 {
			log.Error(nil, "there are more than one configmap created for credentials")
			return ctrl.Result{}, nil
		}
		if IsControlledByThisCredential(credential, childConfigMaps.Items[0]) {
			log.V(1).Info("do nothing", "configmap", childConfigMaps.Items[0])
			return ctrl.Result{}, nil
		}
		configmap, err := r.appendCredentialToConfigMap(&credential, &childConfigMaps.Items[0])
		if err != nil {
			log.Error(err, "unable to append Credential to ConfigMap", "configmap", configmap)
			return ctrl.Result{}, err
		}
		if err := r.Update(ctx, configmap); err != nil {
			log.Error(err, "unable to update ConfigMap for Credential", "configmap", configmap)
			return ctrl.Result{}, err
		}

		log.V(1).Info("updated ConfigMap for Credential run", "configmap", configmap)
	} else {
		configmap, err := r.constructConfigMapForCredential(&credential)
		credential.Status.ConfigMap = configmap.ObjectMeta.Name
		if err := r.Status().Update(ctx, &credential); err != nil {
			log.Error(err, "unable to update credential status")
			return ctrl.Result{}, err
		}

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

func (r *CredentialReconciler) appendCredentialToConfigMap(credential *civ1.Credential, configmap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	credentialsYaml, ok := configmap.Data["credentials-config.yaml"]
	if !ok {
		return nil, fmt.Errorf("Could not find the 'credentials-config.yaml' inside credential configmap.")
	}
	var credentialsData map[string]interface{}
	err := yaml.Unmarshal([]byte(credentialsYaml), &credentialsData)
	if err != nil {
		return nil, err
	}

	credentials := credentialsData["credentials"].([]interface{})
	newCredentialsYaml, err := r.buildCredentialsYaml(credential, credentials)
	if err != nil {
		return nil, err
	}
	configmap.Data["credentials-config.yaml"] = string(newCredentialsYaml)
	if err := SetControllerReferences(credential, configmap, r.Scheme, false); err != nil {
		return nil, err
	}

	return configmap, nil
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

	credentials := make([]interface{}, 0)
	credentialsYaml, err := r.buildCredentialsYaml(credential, credentials)
	if err != nil {
		return nil, err
	}
	cm.Data["credentials-config.yaml"] = credentialsYaml

	for k, v := range credential.ObjectMeta.Labels {
		cm.ObjectMeta.Labels[k] = v
	}
	if err := SetControllerReferences(credential, cm, r.Scheme, true); err != nil {
		return nil, err
	}

	return cm, nil
}

func (r *CredentialReconciler) buildCredentialsYaml(credential *civ1.Credential, currentCredentials []interface{}) (string, error) {
	credentialsData := make(map[string]interface{})
	credentialConfig := make(map[string]interface{})
	credentialConfig["name"] = credential.ObjectMeta.Name
	credentialConfig["type"] = credential.Spec.Type
	credentialConfig["whitelistedTrustedImages"] = credential.Spec.WhitelistedTrustedImages
	credentialConfig["whitelistedPipelines"] = credential.Spec.WhitelistedPipelines
	for k, v := range credential.Spec.AdditionalProperties {
		credentialConfig[k] = v
	}

	credentials := append(currentCredentials, credentialConfig)
	credentialsData["credentials"] = credentials
	credentialsYaml, err := yaml.Marshal(credentialsData)
	if err != nil {
		return "", err
	}

	return string(credentialsYaml), nil
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
		return []string{owner.Kind}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&civ1.Credential{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
