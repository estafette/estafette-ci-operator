package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The external URL this cluster is accessible through.
	ExternalURL string `json:"externalURL"`

	// The internal URL to the estafette-ci-api service for build / release jobs to report to.
	InternalURL string `json:"internalURL"`






    // apiServer:
    //   baseURL: https://estafette.travix.com/
    //   serviceURL: https://estafette.tooling.internal.travix.io/

    // auth:
    //   iap:
    //     enable: true
    //     audience: /projects/1012027921753/global/backendServices/2110500691870907819
    //   apiKey: estafette.secret(F-GHrep4UAYEVH18.fhWaqbptXBp6Zc0jBFaisNwSP2TTqjzwKK6uD8_EomjU6JY_3t21uaRf7lzcUA==)

    // jobs:
    //   namespace: estafette-ci-jobs
    //   minCPUCores: 0.1
    //   maxCPUCores: 7.0
    //   cpuRequestRatio: 1.0
    //   minMemoryBytes: 67108864
    //   maxMemoryBytes: 21474836480
    //   memoryRequestRatio: 1.25

	// database:
    //   databaseName: estafette_ci_api
    //   host: cockroachdb-public.estafette.svc.cluster.local
    //   insecure: true
    //   certificateDir: /cockroach-certs
    //   port: 26257
    //   user: estafette_ci_api
    //   password: estafette.secret(PopOsWdYqBKsgc1f.6KmE3-UWIujjGHGutu6hIbtUdnCOswO2xA79gEhN3emoNs6pyktFH6n4rB4msA==)


	//   # registryMirror: {{.REGISTRY_MIRROR}}
	//   dindMtu: {{.DIND_MTU}}

}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
