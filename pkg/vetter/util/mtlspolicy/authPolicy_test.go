/*
Copyright 2018 Aspen Mesh Authors.

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

package mtlspolicyutil

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	authv1alpha1 "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	istioauthv1alpha1 "istio.io/api/authentication/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	apDefaultOn = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "barNs",
		},
		Spec: authv1alpha1.PolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Targets: []*istioauthv1alpha1.TargetSelector{},
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{
						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
					},
				},
			},
		},
	}

	apFooOn = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apFooOn",
			Namespace: "default",
		},
		Spec: authv1alpha1.PolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Targets: []*istioauthv1alpha1.TargetSelector{
					&istioauthv1alpha1.TargetSelector{
						Name: "foo",
					},
				},
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{
						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
					},
				},
			},
		},
	}

	apFooOff = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apFooOff",
			Namespace: "default",
		},
		Spec: authv1alpha1.PolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Targets: []*istioauthv1alpha1.TargetSelector{
					&istioauthv1alpha1.TargetSelector{
						Name: "foo",
					},
				},
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{},
				},
			},
		},
	}

	apFooBarOn = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apFooBarOn",
			Namespace: "default",
		},
		Spec: authv1alpha1.PolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Targets: []*istioauthv1alpha1.TargetSelector{
					&istioauthv1alpha1.TargetSelector{
						Name: "foo",
					},
					&istioauthv1alpha1.TargetSelector{
						Name: "bar",
					},
				},
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{
						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
					},
				},
			},
		},
	}

	apFooPortsBarOn = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apFooPortsBarOn",
			Namespace: "default",
		},
		Spec: authv1alpha1.PolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Targets: []*istioauthv1alpha1.TargetSelector{
					&istioauthv1alpha1.TargetSelector{
						Name: "foo",
						Ports: []*istioauthv1alpha1.PortSelector{
							&istioauthv1alpha1.PortSelector{
								Port: &istioauthv1alpha1.PortSelector_Number{8443},
							},
						},
					},
					&istioauthv1alpha1.TargetSelector{
						Name: "bar",
					},
				},
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{
						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
					},
				},
			},
		},
	}
)

var _ = Describe("LoadAuthPolicies", func() {
	It("should load policies", func() {
		loaded, err := LoadAuthPolicies([]*authv1alpha1.Policy{
			apDefaultOn,
			apFooOn,
			apFooOff,
			apFooBarOn,
			apFooPortsBarOn,
		})
		Expect(err).To(Succeed())
		foo := Service{Namespace: "default", Name: "foo"}
		bar := Service{Namespace: "default", Name: "bar"}
		Expect(loaded.ByNamespace("barNs")).To(Equal([]*authv1alpha1.Policy{apDefaultOn}))
		Expect(loaded.ByNamespace("default")).To(Equal([]*authv1alpha1.Policy{}))

		Expect(loaded.ByName(foo)).To(Equal([]*authv1alpha1.Policy{
			apFooOn,
			apFooOff,
			apFooBarOn,
			// no apFooPortsBarOn because that is only for foo:8443, not foo
		}))
		Expect(loaded.ByName(bar)).To(Equal([]*authv1alpha1.Policy{
			apFooBarOn,
			// apFooPortsBarOn because that is for bar (not bar:8443)
			apFooPortsBarOn,
		}))

		Expect(loaded.ByPort(foo, 8443)).To(Equal([]*authv1alpha1.Policy{apFooPortsBarOn}))
		Expect(loaded.ByPort(foo, 1000)).To(Equal([]*authv1alpha1.Policy{}))
		Expect(loaded.ByPort(bar, 8443)).To(Equal([]*authv1alpha1.Policy{}))
	})
})

var _ = Describe("AuthPolicyIsMtls", func() {
	It("should evaluate no-mtls-peer as false", func() {
		Expect(AuthPolicyIsMtls(apFooOff)).To(BeFalse())
	})
	It("should evaluate empty-mtls-peer as true", func() {
		Expect(AuthPolicyIsMtls(apFooPortsBarOn)).To(BeTrue())
	})
})
