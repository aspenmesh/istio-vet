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
	meshDefaultOn = &authv1alpha1.MeshPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MeshPolicy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
		Spec: authv1alpha1.MeshPolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{},
				},
			},
		},
	}
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
					// &istioauthv1alpha1.PeerAuthenticationMethod{},
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
		Expect(err).To(BeNil())
		loaded.LoadMeshPolicy([]*authv1alpha1.MeshPolicy{meshDefaultOn})

		foo := Service{Namespace: "default", Name: "foo"}
		bar := Service{Namespace: "default", Name: "bar"}

		Expect(loaded.ByMesh()).To(Equal([]*authv1alpha1.MeshPolicy{meshDefaultOn}))
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
	XIt("should evaluate no-mtls-peer as false", func() {
		Expect(AuthPolicyIsMtls(apFooOff)).To(Equal(MTLSSetting_DISABLED))
	})
	XIt("should evaluate empty-mtls-peer as true", func() {
		Expect(AuthPolicyIsMtls(apFooPortsBarOn)).To(Equal(MTLSSetting_ENABLED))
	})
})

//
//
//
//
//
//
//
//
//
//

var _ = Describe("getModeFromPeers()", func() {
	Context("getModeFromPeers() takes a set of PeerAuthenticationMethods and returns a single mTls Mode", func() {
		peersPermissive := []*istioauthv1alpha1.PeerAuthenticationMethod{
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{
					Mtls: &istioauthv1alpha1.MutualTls{
						Mode: istioauthv1alpha1.MutualTls_PERMISSIVE,
					},
				},
			},
		}
		peersStrict := []*istioauthv1alpha1.PeerAuthenticationMethod{
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{
					Mtls: &istioauthv1alpha1.MutualTls{
						Mode: istioauthv1alpha1.MutualTls_STRICT,
					},
				},
			},
		}
		peersMixed := []*istioauthv1alpha1.PeerAuthenticationMethod{
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{
					Mtls: &istioauthv1alpha1.MutualTls{
						Mode: istioauthv1alpha1.MutualTls_PERMISSIVE,
					},
				},
			},
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{
					Mtls: &istioauthv1alpha1.MutualTls{},
				},
			},
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Jwt{},
			},
		}
		peersEnabledPlusJWT := []*istioauthv1alpha1.PeerAuthenticationMethod{
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{
					Mtls: &istioauthv1alpha1.MutualTls{},
				},
			},
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
			},
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Jwt{},
			},
		}
		peersDisabled := []*istioauthv1alpha1.PeerAuthenticationMethod{
			&istioauthv1alpha1.PeerAuthenticationMethod{},
			&istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Jwt{},
			},
		}
		XIt("returns MIXED when len() == 1 && Mode is set to permissive", func() {
			mtlsState := getModeFromPeers(peersPermissive)
			Expect(mtlsState).To(Equal(MTLSSetting_MIXED))
		})
		XIt("returns ENABLED when len() == 1 && the Mode is STRICT", func() {
			mtlsState := getModeFromPeers(peersStrict)
			Expect(mtlsState).To(Equal(MTLSSetting_ENABLED))
		})
		XIt("returns MIXED when PERMISSIVE is set and there are multiple options enabling mtls", func() {

			mtlsState := getModeFromPeers(peersMixed)
			Expect(mtlsState).To(Equal(MTLSSetting_MIXED))
		})
		XIt("returns ENABLED when there are multiple options enabling auth", func() {
			mtlsState := getModeFromPeers(peersEnabledPlusJWT)
			Expect(mtlsState).To(Equal(MTLSSetting_ENABLED))
		})
		XIt("returns DISABLED when there are no mtls auth methods present", func() {
			mtlsState := getModeFromPeers(peersDisabled)
			Expect(mtlsState).To(Equal(MTLSSetting_DISABLED))
		})
	})
})

var _ = Describe("paramIsMTls()", func() {
	Context("paramIsMTls()", func() {
		XIt("determines whether mtls is enabled for a Peer", func() {

			peer := istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
			}
			peer2 := istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Jwt{},
			}
			peer3 := istioauthv1alpha1.PeerAuthenticationMethod{
				Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{
					Mtls: &istioauthv1alpha1.MutualTls{},
				},
			}

			ok := paramIsMTls(&peer)
			Expect(ok).To(BeTrue())
			ok = paramIsMTls(&peer2)
			Expect(ok).To(BeFalse())
			ok = paramIsMTls(&peer3)
			Expect(ok).To(BeTrue())
		})
	})
})

// Context("testing functions that use policies", func() {
// 	// ----- Begin Set Up -----
// 	policy_On_NSBar_SvcDefault_YesPeers_NoTargets := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "default",
// 			Namespace: "namespaceBar",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{
// 				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
// 					&istioauthv1alpha1.PeerAuthenticationMethod{
// 						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	policy_On_NSBar_SvcDefault_YesPeers_NoTargets_PeerOpt := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "default",
// 			Namespace: "namespaceBar",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{
// 				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
// 					&istioauthv1alpha1.PeerAuthenticationMethod{
// 						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
// 					},
// 				},
// 				PeerIsOptional: true,
// 			},
// 		},
// 	}
// 	policy_Off_NSBar_SvcDefault_NoPeers_NoTargets := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "default",
// 			Namespace: "namespaceBar",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{},
// 		},
// 	}

// 	policy_Off_NSDefault_SvcDefault_NoPeers_NoTargets := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "default",
// 			Namespace: "default",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{},
// 		},
// 	}

// 	policy_On_NSDefault_SvcpolicyForNSDefaultTargetFoo_YesPeers_YesTargets := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "policyForNSDefaultTargetFoo",
// 			Namespace: "default",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{
// 				Targets: []*istioauthv1alpha1.TargetSelector{
// 					&istioauthv1alpha1.TargetSelector{
// 						Name: "foo",
// 					},
// 				},
// 				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
// 					&istioauthv1alpha1.PeerAuthenticationMethod{},
// 				},
// 			},
// 		},
// 	}

// 	nsNameDefault := "default"
// 	nsNameNSBar := "namespaceBar"
// 	meshPolMTlsState := MTLSSetting_MIXED
// 	policiesEmpty := []*authv1alpha1.Policy{}

// 	policies1 := []*authv1alpha1.Policy{
// 		policy_On_NSBar_SvcDefault_YesPeers_NoTargets,
// 		policy_On_NSDefault_SvcpolicyForNSDefaultTargetFoo_YesPeers_YesTargets,
// 	}
// 	policies2 := []*authv1alpha1.Policy{
// 		policy_Off_NSBar_SvcDefault_NoPeers_NoTargets,
// 		policy_Off_NSDefault_SvcDefault_NoPeers_NoTargets,
// 		policy_On_NSDefault_SvcpolicyForNSDefaultTargetFoo_YesPeers_YesTargets,
// 	}
// 	policies3 := []*authv1alpha1.Policy{
// 		policy_Off_NSBar_SvcDefault_NoPeers_NoTargets,
// 		policy_On_NSBar_SvcDefault_YesPeers_NoTargets,
// 	}
// 	policies4 := []*authv1alpha1.Policy{
// 		policy_On_NSBar_SvcDefault_YesPeers_NoTargets_PeerOpt,
// 		policy_Off_NSDefault_SvcDefault_NoPeers_NoTargets,
// 		policy_On_NSDefault_SvcpolicyForNSDefaultTargetFoo_YesPeers_YesTargets,
// 	}
// 	policies5 := []*authv1alpha1.Policy{
// 		policy_Off_NSBar_SvcDefault_NoPeers_NoTargets,
// 		policy_On_NSBar_SvcDefault_YesPeers_NoTargets,
// 	}
// 	policies6 := []*authv1alpha1.Policy{
// 		policy_On_NSBar_SvcDefault_YesPeers_NoTargets,
// 	}
// 	policies7 := []*authv1alpha1.Policy{
// 		policy_Off_NSBar_SvcDefault_NoPeers_NoTargets,
// 	}
// 	// ----- End Set Up -----

// 	Context(" AuthPolicyIsMtls()", func() {
// 		XIt("takes a set of empty policies for a resource and returns the fallback mtlsState", func() {
// 			// policiesEmpty has two equally specific policies
// 			mtlsState := AuthPolicyIsMtls(policiesEmpty)
// 			Expect(mtlsState).To(Equal(MTLSSetting_MIXED))
// 		})
// 		XIt("takes a set of equally specific policies for a resource and returns UNKNOWN", func() {
// 			// policies5 has two equally specific policies
// 			mtlsState := AuthPolicyIsMtls(policies5)
// 			Expect(mtlsState).To(Equal(MTLSSetting_UNKNOWN))
// 		})
// 		XIt("takes a set of one policy for a resource and returns the correct state", func() {
// 			// policies6 has one policy which is disabled
// 			mtlsState := AuthPolicyIsMtls(policies6)
// 			Expect(mtlsState).To(Equal(MTLSSetting_ENABLED))
// 		})
// 		XIt("takes a set of one policy for a resource and returns the correct state", func() {
// 			// policies7 has one policy which is disabled
// 			mtlsState := AuthPolicyIsMtls(policies7)
// 			Expect(mtlsState).To(Equal(MTLSSetting_DISABLED))
// 		})
// 	})
// })
// Context("determinePortPolObjMTLS()", func() {
// 	polOn_NSDefault_SvcFoo_Port8888 := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "foo-policy",
// 			Namespace: "default",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{
// 				Targets: []*istioauthv1alpha1.TargetSelector{
// 					&istioauthv1alpha1.TargetSelector{
// 						Name: "foo",
// 						Ports: []*istioauthv1alpha1.PortSelector{
// 							&istioauthv1alpha1.PortSelector{
// 								Port: &istioauthv1alpha1.PortSelector_Number{Number: 8888},
// 							},
// 						},
// 					},
// 				},
// 				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
// 					&istioauthv1alpha1.PeerAuthenticationMethod{
// 						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	polOff_NSDefault_SvcFoo_Port8888 := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "foo-policy",
// 			Namespace: "default",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{
// 				Targets: []*istioauthv1alpha1.TargetSelector{
// 					&istioauthv1alpha1.TargetSelector{
// 						Name: "foo",
// 						Ports: []*istioauthv1alpha1.PortSelector{
// 							&istioauthv1alpha1.PortSelector{
// 								Port: &istioauthv1alpha1.PortSelector_Number{Number: 8888},
// 							},
// 						},
// 					},
// 				},
// 				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{},
// 			},
// 		},
// 	}
// 	polOff_NSDefault_SvcFoo_Port8118 := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "foo-policy-8118",
// 			Namespace: "default",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{
// 				Targets: []*istioauthv1alpha1.TargetSelector{
// 					&istioauthv1alpha1.TargetSelector{
// 						Name: "foo",
// 						Ports: []*istioauthv1alpha1.PortSelector{
// 							&istioauthv1alpha1.PortSelector{
// 								Port: &istioauthv1alpha1.PortSelector_Number{Number: 8118},
// 							},
// 						},
// 					},
// 				},
// 				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{},
// 			},
// 		},
// 	}
// 	polOn_NSDefault_SvcFoo_Port8118 := &authv1alpha1.Policy{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "Policy",
// 			APIVersion: "authentication.istio.io/v1alpha1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "foo-policy-8118",
// 			Namespace: "default",
// 		},
// 		Spec: authv1alpha1.PolicySpec{
// 			Policy: istioauthv1alpha1.Policy{
// 				Targets: []*istioauthv1alpha1.TargetSelector{
// 					&istioauthv1alpha1.TargetSelector{
// 						Name: "foo",
// 						Ports: []*istioauthv1alpha1.PortSelector{
// 							&istioauthv1alpha1.PortSelector{
// 								Port: &istioauthv1alpha1.PortSelector_Number{Number: 8118},
// 							},
// 						},
// 					},
// 				},
// 				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
// 					&istioauthv1alpha1.PeerAuthenticationMethod{
// 						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	XIt("returns MIXED if there are equally specific policies", func() {
// 		nsName := "default"
// 		svcName := "foo"
// 		policies := []*authv1alpha1.Policy{
// 			polOff_NSDefault_SvcFoo_Port8888, polOn_NSDefault_SvcFoo_Port8888,
// 		}

// 		clusterAuthPols, errLocal := LoadAuthPolicies(policies)
// 		Expect(errLocal).To(BeNil())
// 		svcPortPols := determinePortPolObjMTLS(nsName, svcName, clusterAuthPols, MTLSSetting_MIXED)
// 		Expect(svcPortPols).To(Equal(MTLSSetting_MIXED))

// 	})
// 	XIt("returns MIXED if there are any policies for the same service and different ports which conflict", func() {
// 		nsName := "default"
// 		svcName := "foo"
// 		policies := []*authv1alpha1.Policy{
// 			polOn_NSDefault_SvcFoo_Port8888,
// 			polOff_NSDefault_SvcFoo_Port8118,
// 		}

// 		clusterAuthPols, errLocal := LoadAuthPolicies(policies)
// 		Expect(errLocal).To(BeNil())
// 		svcPortPols := determinePortPolObjMTLS(nsName, svcName, clusterAuthPols, MTLSSetting_MIXED)
// 		Expect(svcPortPols).To(Equal(MTLSSetting_MIXED))
// 	})
// 	XIt("returns ENABLED if there are any policies for the same service and same port which are all enabled", func() {
// 		nsName := "default"
// 		svcName := "foo"
// 		policies := []*authv1alpha1.Policy{
// 			polOn_NSDefault_SvcFoo_Port8888,
// 			polOn_NSDefault_SvcFoo_Port8118,
// 		}

// 		clusterAuthPols, errLocal := LoadAuthPolicies(policies)
// 		Expect(errLocal).To(BeNil())
// 		svcPortPols := determinePortPolObjMTLS(nsName, svcName, clusterAuthPols, MTLSSetting_MIXED)
// 		Expect(svcPortPols).To(Equal(MTLSSetting_ENABLED))
// 	})
// })
