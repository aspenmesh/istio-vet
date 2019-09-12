/*
Copyright 2019 Aspen Mesh Authors.

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

package invalidserviceforjwtpolicy

import (
	authv1alpha1api "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istiov1alpha1 "istio.io/api/authentication/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("Invalid Service For JWT Policy Vet Notes", func() {
	const (
		namespace = "default"
		vetterID                                = "InvalidServiceForJWTPolicy"
		invalidTargetServicePortNameNoteType    = "invalid-target-service-port-name"
		invalidTargetServicePortNameNoteSummary = "Target services must have valid service port names"
		invalidTargetServicePortNameNoteMsg     = "The authentication policy '${policy}' in namespace '${namespace}' has a target of" +
			" service '${service_target}', which does not contain a valid port name. Port names must be 'http', 'http2', 'https'," +
			" or must be prefixed with 'http-', 'http2-', or 'https-'."
		missingTargetServiceNoteType = "missing-target-service"
		missingTargetServiceSummary  = "The authentication policy target service was not found in namespace '${namespace}'"
		missingTargetServiceNoteMsg  = "The authentication policy '${policy}' in namespace '${namespace}' references the service" +
			" '${service_target}', which does not exist in namespace '${namespace}'."
	)

	httpbinJWTAuthPolicy := &authv1alpha1api.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:         "jwt-example",
			Namespace:    namespace,
			Initializers: &metav1.Initializers{}},
		Spec: authv1alpha1api.PolicySpec{
			Policy: istiov1alpha1.Policy{
				Origins: []*istiov1alpha1.OriginAuthenticationMethod{
					{
						Jwt: &istiov1alpha1.Jwt{
							Issuer: "testing@secure.istio.io",
							JwksUri: "https://raw.githubusercontent.com/istio/istio/release-1.2/security/tools/jwt/samples/jwks.json",
						},
					},
				},
				Targets: []*istiov1alpha1.TargetSelector{
					{
						Name:  "httpbin",
						Ports: []*istiov1alpha1.PortSelector{},
					},
				},
			},
		},
	}

	Context("when target service ports are named", func() {
		It("should not generate notes if any of the port names are 'http'", func() {
			nsServices := []*corev1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "httpbin",
						Namespace: namespace,
						Initializers: &metav1.Initializers{},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Protocol: "TCP",
								Name: "http",
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9376,
									StrVal: "9376",
								},
							},
						},
						Selector: map[string]string {
							"app": "httpbin",
						},
					},
				},
			}
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(0))
		})

		It("should not generate notes if any of the port names are 'http2'", func() {
			nsServices := []*corev1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "httpbin",
						Namespace: namespace,
						Initializers: &metav1.Initializers{},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Protocol: "TCP",
								Name: "http2",
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9376,
									StrVal: "9376",
								},
							},
						},
						Selector: map[string]string {
							"app": "httpbin",
						},
					},
				},
			}
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(0))
		})

		It("should not generate notes if any of the port names are 'https'", func() {
			nsServices := []*corev1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "httpbin",
						Namespace: namespace,
						Initializers: &metav1.Initializers{},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Protocol: "TCP",
								Name: "https",
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9376,
									StrVal: "9376",
								},
							},
						},
						Selector: map[string]string {
							"app": "httpbin",
						},
					},
				},
			}
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(0))
		})

		It("should not generate notes if any of the port names are prefixed with 'http-'", func() {
			nsServices := []*corev1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "httpbin",
						Namespace: namespace,
						Initializers: &metav1.Initializers{},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Protocol: "TCP",
								Name: "http-app",
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9376,
									StrVal: "9376",
								},
							},
						},
						Selector: map[string]string {
							"app": "httpbin",
						},
					},
				},
			}
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(0))
		})

		It("should not generate notes if any of the port names are prefixed with 'http2-'", func() {
			nsServices := []*corev1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "httpbin",
						Namespace: namespace,
						Initializers: &metav1.Initializers{},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Protocol: "TCP",
								Name: "http2-app",
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9376,
									StrVal: "9376",
								},
							},
						},
						Selector: map[string]string {
							"app": "httpbin",
						},
					},
				},
			}
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(0))
		})

		It("should not generate notes if any of the port names are prefixed with 'https-'", func() {
			nsServices := []*corev1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "httpbin",
						Namespace: namespace,
						Initializers: &metav1.Initializers{},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Protocol: "TCP",
								Name: "https-app",
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9376,
									StrVal: "9376",
								},
							},
						},
						Selector: map[string]string {
							"app": "httpbin",
						},
					},
				},
			}
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(0))
		})

		It("generates an 'invalid-target-service-port-name' note when none of the port names are valid", func() {
			expectedNote := &apiv1.Note{
				Type:    invalidTargetServicePortNameNoteType,
				Summary: invalidTargetServicePortNameNoteSummary,
				Msg:     invalidTargetServicePortNameNoteMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"policy": "jwt-example",
					"namespace": "default",
					"service_target": "httpbin",
				},
			}
			expectedNote.Id = util.ComputeID(expectedNote)
			nsServices := []*corev1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "httpbin",
						Namespace: namespace,
						Initializers: &metav1.Initializers{},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Protocol: "TCP",
								Name: "not-valid-name",
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9376,
									StrVal: "9376",
								},
							},
							{
								Protocol: "TCP",
								Name: "httpbutstillnotvalid",
								Port: 81,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9377,
									StrVal: "9377",
								},
							},
						},
						Selector: map[string]string {
							"app": "httpbin",
						},
					},
				},
			}
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(1))
			Expect(actualNotes[0]).To(Equal(expectedNote))
		})
	})

	Context("when target services ports are not named", func() {
		It("generates an 'invalid-target-service-port-name' note", func() {
			expectedNote := &apiv1.Note{
				Type:    invalidTargetServicePortNameNoteType,
				Summary: invalidTargetServicePortNameNoteSummary,
				Msg:     invalidTargetServicePortNameNoteMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"policy": "jwt-example",
					"namespace": "default",
					"service_target": "httpbin",
				},
			}
			expectedNote.Id = util.ComputeID(expectedNote)
			nsServices := []*corev1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "httpbin",
						Namespace: namespace,
						Initializers: &metav1.Initializers{},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Protocol: "TCP",
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9376,
									StrVal: "9376",
								},
							},
							{
								Protocol: "TCP",
								Port: 81,
								TargetPort: intstr.IntOrString{
									Type:   0,
									IntVal: 9377,
									StrVal: "9377",
								},
							},
						},
						Selector: map[string]string {
							"app": "httpbin",
						},
					},
				},
			}
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(1))
			Expect(actualNotes[0]).To(Equal(expectedNote))
		})
	})

	Context("when target service is not found in the current policy namespace", func() {
		It("generates an 'missing-target-service' note", func() {
			expectedNote := &apiv1.Note{
				Type:    missingTargetServiceNoteType,
				Summary: missingTargetServiceSummary,
				Msg:     missingTargetServiceNoteMsg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"policy": "jwt-example",
					"namespace": "default",
					"service_target": "httpbin",
				},
			}
			expectedNote.Id = util.ComputeID(expectedNote)
			var nsServices []*corev1.Service
			nsServiceLookup := createServiceLookup(nsServices)
			actualNotes := createAuthPolicyNotes(httpbinJWTAuthPolicy, nsServiceLookup)
			Expect(len(actualNotes)).To(Equal(1))
			Expect(actualNotes[0]).To(Equal(expectedNote))
		})
	})
})
