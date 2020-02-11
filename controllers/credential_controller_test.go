package controllers

import (
	"context"
	"time"

	civ1 "github.com/estafette/estafette-ci-operator/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Credential Controller", func() {

	const timeout = time.Second * 10
	const interval = time.Second * 1
	const charset = "abcdefghijklmnopqrstuvwxyz"

	var credentialKeyName = "credential-" + randomStringWithCharset(10, charset)
	var privateCredentialKeyName = "credential-private-" + randomStringWithCharset(10, charset)
	var gkeCredentialKeyName = "credential-gke-" + randomStringWithCharset(10, charset)

	BeforeEach(func() {
		// failed test runs that don't clean up leave resources behind.
		keys := []string{credentialKeyName, privateCredentialKeyName}
		for _, value := range keys {

			cred := &civ1.Credential{
				ObjectMeta: metav1.ObjectMeta{
					Name:      value,
					Namespace: "default",
				},
			}

			_ = k8sClient.Delete(context.Background(), cred)
		}
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
		keys := []string{credentialKeyName, privateCredentialKeyName}
		for _, value := range keys {

			cred := &civ1.Credential{
				ObjectMeta: metav1.ObjectMeta{
					Name:      value,
					Namespace: "default",
				},
			}

			_ = k8sClient.Delete(context.Background(), cred)
		}
	})

	Context("One crededential", func() {
		It("Should handle credential correctly", func() {

			spec := civ1.CredentialSpec{
				Type:                 "container-registry",
				WhitelistedPipelines: "github.com/estafette/.+",
				AdditionalProperties: map[string]interface{}{
					"repository": "estafette",
					"private":    false,
					"username":   "estafettesvc",
					"password":   "supersecretpassword",
				},
			}

			key := types.NamespacedName{
				Name:      credentialKeyName,
				Namespace: "default",
			}

			configMapKey := types.NamespacedName{
				Name:      "estafette-external-credentials",
				Namespace: "default",
			}

			toCreate := &civ1.Credential{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec:   spec,
				Status: civ1.CredentialStatus{},
			}

			By("Creating the credential successfully")
			// act
			expectedResult := Expect(k8sClient.Create(context.Background(), toCreate))

			// assert
			expectedResult.Should(Succeed())

			time.Sleep(time.Second * 2)
			fetched := &civ1.Credential{}

			// act
			eventualResult := Eventually(func() string {
				_ = k8sClient.Get(context.Background(), key, fetched)
				return fetched.Status.ConfigMap
			}, timeout, interval)

			// assert
			eventualResult.Should(Equal(configMapKey.Name))

			By("Creating configmap successfully")
			fetchedConfigMap := &corev1.ConfigMap{}
			expectedConfig := map[string]interface{}{
				"name":                 credentialKeyName,
				"type":                 "container-registry",
				"whitelistedPipelines": "github.com/estafette/.+",
				"repository":           "estafette",
				"private":              false,
				"username":             "estafettesvc",
				"password":             "supersecretpassword",
			}
			expectedCredentialsData := map[string]interface{}{
				"credentials": []interface{}{expectedConfig},
			}
			expectedYaml, err := yaml.Marshal(expectedCredentialsData)
			Expect(err).To(BeNil())

			// act
			eventualResult = Eventually(func() string {
				_ = k8sClient.Get(context.Background(), configMapKey, fetchedConfigMap)
				return fetchedConfigMap.Data["credentials-config.yaml"]
			}, timeout, interval)

			// assert
			eventualResult.Should(Equal(string(expectedYaml)))

			By("Deleting the credential")
			// act
			eventualResult = Eventually(func() error {
				f := &civ1.Credential{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval)

			// assert
			eventualResult.Should(Succeed())

			// act (check if credential is really deleted)
			eventualResult = Eventually(func() error {
				f := &civ1.Credential{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval)

			// assert
			eventualResult.ShouldNot(Succeed())
		})
	})

	Context("Two credentials", func() {
		It("Should handle two credentials correctly", func() {

			specGkeCredential := civ1.CredentialSpec{
				Type:                     "kubernetes-engine",
				WhitelistedTrustedImages: "extensions/gke",
				AdditionalProperties: map[string]interface{}{
					"project": "estafette-project",
					"cluster": "estafette-cluster",
					"defaults": map[string]interface{}{
						"namespace": "estafette-ns",
					},
				},
			}

			gkeCredentialKey := types.NamespacedName{
				Name:      gkeCredentialKeyName,
				Namespace: "default",
			}

			gkeCredentialToCreate := &civ1.Credential{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gkeCredentialKey.Name,
					Namespace: gkeCredentialKey.Namespace,
				},
				Spec:   specGkeCredential,
				Status: civ1.CredentialStatus{},
			}

			privateCredentialSpec := civ1.CredentialSpec{
				Type: "container-registry",
				AdditionalProperties: map[string]interface{}{
					"repository": "gcr.io/estafette/",
					"private":    true,
					"username":   "estafettesecretsvc",
					"password":   "supersupersecretpassword",
				},
			}

			privateCredentialKey := types.NamespacedName{
				Name:      privateCredentialKeyName,
				Namespace: "default",
			}

			privateCredentialToCreate := &civ1.Credential{
				ObjectMeta: metav1.ObjectMeta{
					Name:      privateCredentialKey.Name,
					Namespace: privateCredentialKey.Namespace,
				},
				Spec:   privateCredentialSpec,
				Status: civ1.CredentialStatus{},
			}

			configMapKey := types.NamespacedName{
				Name:      "estafette-external-credentials",
				Namespace: "default",
			}

			By("Creating gke credential successfully")
			// act
			expectedResult := Expect(k8sClient.Create(context.Background(), gkeCredentialToCreate))

			// assert
			expectedResult.Should(Succeed())
			time.Sleep(time.Second * 2)

			fetchedGkeCredential := &civ1.Credential{}

			// act
			eventualResult := Eventually(func() string {
				_ = k8sClient.Get(context.Background(), gkeCredentialKey, fetchedGkeCredential)
				return fetchedGkeCredential.Status.ConfigMap
			}, timeout, interval)

			// assert
			eventualResult.Should(Equal(configMapKey.Name))

			By("Creating configmap successfully")
			fetchedConfigMap := &corev1.ConfigMap{}

			gkeCredentialExpectedConfig := map[string]interface{}{
				"name":                     gkeCredentialKeyName,
				"type":                     "kubernetes-engine",
				"whitelistedTrustedImages": "extensions/gke",
				"project":                  "estafette-project",
				"cluster":                  "estafette-cluster",
				"defaults": map[string]interface{}{
					"namespace": "estafette-ns",
				},
			}
			expectedCredentialsData := map[string]interface{}{
				"credentials": []interface{}{gkeCredentialExpectedConfig},
			}
			expectedGkeCredentialYaml, err := yaml.Marshal(expectedCredentialsData)
			Expect(err).To(BeNil())

			// act
			eventualResult = Eventually(func() string {
				_ = k8sClient.Get(context.Background(), configMapKey, fetchedConfigMap)
				return fetchedConfigMap.Data["credentials-config.yaml"]
			}, timeout, interval)

			// assert
			eventualResult.Should(Equal(string(expectedGkeCredentialYaml)))

			By("Creating private credential successfully")

			// act
			expectedResult = Expect(k8sClient.Create(context.Background(), privateCredentialToCreate))

			// assert
			expectedResult.Should(Succeed())
			time.Sleep(time.Second * 2)

			fetchedPrivateCredential := &civ1.Credential{}

			// act
			eventualResult = Eventually(func() string {
				_ = k8sClient.Get(context.Background(), privateCredentialKey, fetchedPrivateCredential)
				return fetchedPrivateCredential.Status.ConfigMap
			}, timeout, interval)

			// assert
			eventualResult.Should(Equal(configMapKey.Name))

			By("Updating the configmap successfully")
			privateCredentialExpectedConfig := map[string]interface{}{
				"name":       privateCredentialKeyName,
				"type":       "container-registry",
				"repository": "gcr.io/estafette/",
				"private":    true,
				"username":   "estafettesecretsvc",
				"password":   "supersupersecretpassword",
			}

			expectedCredentialsData["credentials"] = append(expectedCredentialsData["credentials"].([]interface{}), privateCredentialExpectedConfig)

			expectedPrivateCredentialYaml, err := yaml.Marshal(expectedCredentialsData)
			Expect(err).To(BeNil())

			// act
			eventualResult = Eventually(func() string {
				_ = k8sClient.Get(context.Background(), configMapKey, fetchedConfigMap)
				return fetchedConfigMap.Data["credentials-config.yaml"]
			}, timeout, interval)

			// assert
			eventualResult.Should(Equal(string(expectedPrivateCredentialYaml)))

			By("Deleting private credential")

			// act
			eventualResult = Eventually(func() error {
				f := &civ1.Credential{}
				_ = k8sClient.Get(context.Background(), privateCredentialKey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval)

			// assert
			eventualResult.Should(Succeed())

			// act
			eventualResult = Eventually(func() error {
				f := &civ1.Credential{}
				return k8sClient.Get(context.Background(), privateCredentialKey, f)
			}, timeout, interval)

			// assert
			eventualResult.ShouldNot(Succeed())
		})
	})

})
