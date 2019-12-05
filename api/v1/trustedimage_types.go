package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TrustedImageSpec defines the desired state of TrustedImage
// +k8s:openapi-gen=true
type TrustedImageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Type string `json:"type"`
	// +kubebuilder:validation:Optional
	WhitelistedPipelines string `json:"whitelistedPipelines,omitempty"`
	// +kubebuilder:validation:Required
	Path string `json:"path"`
	// +kubebuilder:validation:Optional
	RunPrivileged bool `json:"runPrivileged"`
	// +kubebuilder:validation:Optional
	RunDocker bool `json:"runDocker"`
	// +kubebuilder:validation:Optional
	AllowCommands bool `json:"allowCommands"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:UniqueItems
	InjectedCredentialTypes []string `json:"injectedCredentialTypes,omitempty"`
}

// TrustedImageStatus defines the observed state of TrustedImage
type TrustedImageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// TrustedImage is the Schema for the trustedimages API
type TrustedImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TrustedImageSpec   `json:"spec,omitempty"`
	Status TrustedImageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TrustedImageList contains a list of TrustedImage
type TrustedImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrustedImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TrustedImage{}, &TrustedImageList{})
}
