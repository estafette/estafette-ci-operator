package controllers

import (
	"fmt"
	civ1 "github.com/estafette/estafette-ci-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func SetControllerReferences(owner, controlled metav1.Object, scheme *runtime.Scheme, isController bool) error {
	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("%T is not a runtime.Object, cannot call SetControllerReference", owner)
	}

	ownerNs := owner.GetNamespace()
	if ownerNs != "" {
		objNs := controlled.GetNamespace()
		if objNs == "" {
			return fmt.Errorf("cluster-scoped resource must not have a namespace-scoped owner, owner's namespace %s", ownerNs)
		}
		if ownerNs != objNs {
			return fmt.Errorf("cross-namespace owner references are disallowed, owner's namespace %s, obj's namespace %s", owner.GetNamespace(), controlled.GetNamespace())
		}
	}

	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}

	// Create a new ref

	// blockOwnerDeletion := true
	// isController := true
	// ref := metav1.OwnerReference{
	// 	APIVersion:         gvk.GroupVersion().String(),
	// 	Kind:               gvk.Kind,
	// 	Name:               gvk.Kind,
	// 	UID:                owner.GetUID(),
	// 	BlockOwnerDeletion: &blockOwnerDeletion,
	// 	Controller:         &isController,
	// }

	ref := *metav1.NewControllerRef(owner, schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})
	ref.Controller = &isController
	ref.BlockOwnerDeletion = &isController

	existingRefs := controlled.GetOwnerReferences()
	fi := -1
	for i, r := range existingRefs {
		if referSameObject(ref, r) {
			fi = i
		}
	}
	if fi == -1 {
		existingRefs = append(existingRefs, ref)
	} else {
		existingRefs[fi] = ref
	}

	// Update owner references
	controlled.SetOwnerReferences(existingRefs)
	return nil
}

func IsControlledByThisCredential(credential civ1.Credential, configmap corev1.ConfigMap) bool {
	return true
}

func referSameObject(a, b metav1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV.Group == bGV.Group && a.Kind == b.Kind && a.Name == b.Name
}
