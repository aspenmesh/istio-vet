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
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	authv1alpha1 "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	istioauthv1alpha1 "istio.io/api/authentication/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	meshDefaultOn = []*authv1alpha1.MeshPolicy{
		&authv1alpha1.MeshPolicy{
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
						&istioauthv1alpha1.PeerAuthenticationMethod{
							Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
						},
					},
				},
			},
		},
	}

	nsbarNs_On = &authv1alpha1.Policy{
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
	nsbarNs_Off = &authv1alpha1.Policy{
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
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{},
			},
		},
	}
	nsDefault_apFoo_On = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nsDefault_apFoo_On",
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
	nsDefault_apFoo_Off = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nsDefault_apFoo_Off",
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
	nsDefault_apFoo_apBar_On = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nsDefault_apFoo_apBar_On",
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
	nsDefault_apFooPorts_apBar_On = &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nsDefault_apFooPorts_apBar_On",
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
	peersPermissive = []*istioauthv1alpha1.PeerAuthenticationMethod{
		&istioauthv1alpha1.PeerAuthenticationMethod{
			Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{
				Mtls: &istioauthv1alpha1.MutualTls{
					Mode: istioauthv1alpha1.MutualTls_PERMISSIVE,
				},
			},
		},
	}
	peersStrict = []*istioauthv1alpha1.PeerAuthenticationMethod{
		&istioauthv1alpha1.PeerAuthenticationMethod{
			Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{
				Mtls: &istioauthv1alpha1.MutualTls{
					Mode: istioauthv1alpha1.MutualTls_STRICT,
				},
			},
		},
	}
	peersMixed = []*istioauthv1alpha1.PeerAuthenticationMethod{
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
	peersEnabledPlusJWT = []*istioauthv1alpha1.PeerAuthenticationMethod{
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
	peersDisabled = []*istioauthv1alpha1.PeerAuthenticationMethod{
		&istioauthv1alpha1.PeerAuthenticationMethod{},
		&istioauthv1alpha1.PeerAuthenticationMethod{
			Params: &istioauthv1alpha1.PeerAuthenticationMethod_Jwt{},
		},
	}
	peersEmpty = []*istioauthv1alpha1.PeerAuthenticationMethod{}
	noTargets  = []*istioauthv1alpha1.TargetSelector{}
)

func diyPolicy(nsName, polName string, peers []*istioauthv1alpha1.PeerAuthenticationMethod, targets []*istioauthv1alpha1.TargetSelector) *authv1alpha1.Policy {
	return &authv1alpha1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      polName,
			Namespace: nsName,
		},
		Spec: authv1alpha1.PolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Peers:   peers,
				Targets: targets,
			},
		},
	}
}

func targetWithPort(tName string, pNum uint32) []*istioauthv1alpha1.TargetSelector {
	return []*istioauthv1alpha1.TargetSelector{
		&istioauthv1alpha1.TargetSelector{
			Name: tName,
			Ports: []*istioauthv1alpha1.PortSelector{
				&istioauthv1alpha1.PortSelector{
					Port: &istioauthv1alpha1.PortSelector_Number{pNum},
				},
			},
		},
	}
}
func targetNoPort(tName string) []*istioauthv1alpha1.TargetSelector {
	return []*istioauthv1alpha1.TargetSelector{
		&istioauthv1alpha1.TargetSelector{
			Name: tName,
		},
	}
}

var _ = Describe("LoadAuthPolicies and LoadMeshPolicy", func() {
	It("should load policies", func() {
		loaded, err := LoadAuthPolicies([]*authv1alpha1.Policy{
			nsbarNs_On,
			nsDefault_apFoo_On,
			nsDefault_apFoo_Off,
			nsDefault_apFoo_apBar_On,
			nsDefault_apFooPorts_apBar_On,
		}, meshDefaultOn)
		Expect(err).To(BeNil())

		foo := Service{Namespace: "default", Name: "foo"}
		bar := Service{Namespace: "default", Name: "bar"}

		Expect(loaded.ByMesh()).To(Equal(meshDefaultOn))
		Expect(loaded.ByNamespace("barNs")).To(Equal([]*authv1alpha1.Policy{nsbarNs_On}))
		Expect(loaded.ByNamespace("default")).To(Equal([]*authv1alpha1.Policy{}))

		Expect(loaded.ByName(foo)).To(Equal([]*authv1alpha1.Policy{
			nsDefault_apFoo_On,
			nsDefault_apFoo_Off,
			nsDefault_apFoo_apBar_On,
			// no nsDefault_apFooPorts_apBar_On because that is only for foo:8443, not foo
		}))
		Expect(loaded.ByName(bar)).To(Equal([]*authv1alpha1.Policy{
			nsDefault_apFoo_apBar_On,
			// nsDefault_apFooPorts_apBar_On because that is for bar (not bar:8443)
			nsDefault_apFooPorts_apBar_On,
		}))

		Expect(loaded.ByPort(foo, 8443)).To(Equal([]*authv1alpha1.Policy{nsDefault_apFooPorts_apBar_On}))
		Expect(loaded.ByPort(foo, 1000)).To(Equal([]*authv1alpha1.Policy{}))
		Expect(loaded.ByPort(bar, 8443)).To(Equal([]*authv1alpha1.Policy{}))
	})
})

var _ = Describe("AuthPolicyIsMtls", func() {
	It("should evaluate no-mtls-peer as false", func() {
		Expect(AuthPolicyIsMtls(nsDefault_apFoo_Off)).To(Equal(MTLSSetting_DISABLED))
	})
	It("should evaluate empty-mtls-peer as true", func() {
		Expect(AuthPolicyIsMtls(nsDefault_apFooPorts_apBar_On)).To(Equal(MTLSSetting_ENABLED))
	})
})

var _ = Describe("TLS Details", func() {
	Context("TLSDetailsByNamespace()", func() {

		It("returns enabled when there is an enabled policy", func() {
			loadedOn, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				nsbarNs_On,
			}, meshDefaultOn)

			Expect(err).To(BeNil())
			s := Service{Namespace: "barNs", Name: ""}
			mtlsStateOn, _, err1 := loadedOn.TLSDetailsByNamespace(s)
			Expect(err1).To(BeNil())
			Expect(mtlsStateOn).To(Equal(MTLSSetting_ENABLED))
		})
		It("returns enabled when there is no policy for a namespace, but the mesh policy exists and is enabled", func() {
			loadedNone, err := LoadAuthPolicies([]*authv1alpha1.Policy{}, meshDefaultOn)
			Expect(err).To(BeNil())
			s := Service{Namespace: "barNs", Name: ""}
			mtlsStateNone, _, err2 := loadedNone.TLSDetailsByNamespace(s)
			Expect(err2).To(BeNil())
			Expect(mtlsStateNone).To(Equal(MTLSSetting_ENABLED))
		})
	})

	Context("TLSDetailsByName()", func() {
		It("returns enabled when there is an enabled policy", func() {
			loadedOn, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				nsDefault_apFooPorts_apBar_On,
			}, nil)
			Expect(err).To(BeNil())
			s := Service{Namespace: "default", Name: "bar"}
			mtlsStateOn, _, err1 := loadedOn.TLSDetailsByName(s)

			Expect(err1).To(BeNil())
			Expect(mtlsStateOn).To(Equal(MTLSSetting_ENABLED))
		})
		It("returns disabled when there is an enabled Port policy and disabled service policy", func() {
			loadedOn, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				nsDefault_apFoo_Off,
				nsDefault_apFooPorts_apBar_On,
			}, nil)
			Expect(err).To(BeNil())
			s := Service{Namespace: "default", Name: "foo"}
			mtlsStateOn, _, err1 := loadedOn.TLSDetailsByName(s)

			Expect(err1).To(BeNil())
			Expect(mtlsStateOn).To(Equal(MTLSSetting_DISABLED))
		})
		It("returns enabled when there is no policy for a service, but the namespace policy exists and is enabled", func() {
			loadedNone, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				nsbarNs_On,
			}, nil)
			Expect(err).To(BeNil())
			s := Service{Namespace: "barNs", Name: ""}
			mtlsStateNone, _, err2 := loadedNone.TLSDetailsByName(s)

			Expect(err2).To(BeNil())
			Expect(mtlsStateNone).To(Equal(MTLSSetting_ENABLED))
		})
	})
	Context("TLSDetailsByPort()", func() {

		It("returns enabled when there is an enabled policy", func() {
			loadedOn, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				nsDefault_apFooPorts_apBar_On,
			}, nil)
			Expect(err).To(BeNil())
			s := Service{Namespace: "default", Name: "foo"}
			mtlsStateOn, _, err1 := loadedOn.TLSDetailsByPort(s, 8443)

			Expect(err1).To(BeNil())
			Expect(mtlsStateOn).To(Equal(MTLSSetting_ENABLED))
		})

		It("returns enabled when there is no policy for a target service Port, but a service policy exists and is enabled", func() {
			loadedNone, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				nsDefault_apFoo_apBar_On,
			}, nil)
			Expect(err).To(BeNil())
			s := Service{Namespace: "default", Name: "foo"}
			mtlsStateNone, _, err2 := loadedNone.TLSDetailsByPort(s, 8443)

			Expect(err2).To(BeNil())
			Expect(mtlsStateNone).To(Equal(MTLSSetting_ENABLED))
		})
	})
})

var _ = Describe("ForEachPolByPort()", func() {
	Context("tallies the right mtlsState when an AuthPolicies struct is checked for port policies", func() {
		var enabled, disabled, mixed, unknown int
		BeforeEach(func() {
			enabled, disabled, mixed, unknown = 0, 0, 0, 0
		})

		cb := func(policies []*authv1alpha1.Policy) {
			if len(policies) == 1 {
				mtlsState := AuthPolicyIsMtls(policies[0])
				switch state := mtlsState; {
				case state == MTLSSetting_MIXED:
					mixed++
				case state == MTLSSetting_ENABLED:
					enabled++
				case state == MTLSSetting_DISABLED:
					disabled++
				}
			} else {
				unknown++
			}
		}
		// make some test cases that have multiple policies that hit the cb so that I can count the tallies. Maybe one big one and then have expects that match all the tallies. FRPBP(s) takes a service.

		It("when passed valid policies with target ports", func() {
			loadedOn, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				diyPolicy("default", "default", peersDisabled, noTargets),
				diyPolicy("default", "pol-1", peersStrict, targetWithPort("foo", uint32(8123))),
				diyPolicy("default", "pol-2", peersPermissive, targetWithPort("foo", uint32(8456))),
				diyPolicy("default", "pol-3", peersDisabled, targetWithPort("foo", uint32(8789))),
				diyPolicy("default", "pol-4", peersEnabledPlusJWT, targetWithPort("foo", uint32(8000))),
				diyPolicy("default", "pol-5", peersDisabled, targetWithPort("foo", uint32(8999))),
				diyPolicy("default", "pol-6", peersDisabled, targetWithPort("foo", uint32(8999))),
			}, nil)
			Expect(err).To(BeNil())
			s := Service{Namespace: "default", Name: "foo"}
			loadedOn.ForEachPolByPort(s, cb)

			Expect(enabled).To(Equal(2))
			Expect(disabled).To(Equal(1))
			Expect(mixed).To(Equal(1))
			Expect(unknown).To(Equal(1))

		})
		It("when passed valid policies with different target ports", func() {
			loadedOn, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				diyPolicy("default", "default", peersDisabled, noTargets),
				diyPolicy("default", "pol-2", peersDisabled, targetWithPort("bar", uint32(8888))),
				diyPolicy("default", "pol-3", peersPermissive, targetWithPort("foo", uint32(8765))),
			}, nil)
			Expect(err).To(BeNil())
			s := Service{Namespace: "default", Name: "foo"}
			loadedOn.ForEachPolByPort(s, cb)

			Expect(disabled).To(Equal(0))
			Expect(mixed).To(Equal(1))

		})
		It("when passed valid policies with NO target port", func() {
			loadedOn, err := LoadAuthPolicies([]*authv1alpha1.Policy{
				diyPolicy("default", "default", peersDisabled, noTargets),
				diyPolicy("ns-2", "default", peersPermissive, targetNoPort("svc-1")),
			}, nil)
			Expect(err).To(BeNil())
			s := Service{Namespace: "default", Name: "foo"}
			loadedOn.ForEachPolByPort(s, cb)

			Expect(enabled).To(Equal(0))
			Expect(disabled).To(Equal(0))
			Expect(enabled).To(Equal(0))
			Expect(enabled).To(Equal(0))

		})
	})
})
var _ = Describe("getModeFromPeers()", func() {
	Context("getModeFromPeers() takes a set of PeerAuthenticationMethods and returns a single mTls Mode", func() {

		It("returns MIXED when len() == 1 && Mode is set to permissive", func() {
			mtlsState := getModeFromPeers(peersPermissive)
			Expect(mtlsState).To(Equal(MTLSSetting_MIXED))
		})
		It("returns ENABLED when len() == 1 && the Mode is STRICT", func() {
			mtlsState := getModeFromPeers(peersStrict)
			Expect(mtlsState).To(Equal(MTLSSetting_ENABLED))
		})
		It("returns MIXED when PERMISSIVE is set and there are multiple options enabling mtls", func() {

			mtlsState := getModeFromPeers(peersMixed)
			Expect(mtlsState).To(Equal(MTLSSetting_MIXED))
		})
		It("returns ENABLED when there are multiple options enabling auth", func() {
			mtlsState := getModeFromPeers(peersEnabledPlusJWT)
			Expect(mtlsState).To(Equal(MTLSSetting_ENABLED))
		})
		It("returns DISABLED when there are no mtls auth methods present", func() {
			mtlsState := getModeFromPeers(peersDisabled)
			Expect(mtlsState).To(Equal(MTLSSetting_DISABLED))
		})
	})
})
var _ = Describe("evaluateMTlsForPeer()", func() {
	Context("takes a set of peers and the peerIsOptional setting, and returns an mTls setting", func() {

		It("when peerIsOptional is true, it returns MIXED", func() {
			pio := true
			mtlsState := evaluateMTlsForPeer(peersEnabledPlusJWT, pio)
			Expect(mtlsState).To(Equal(MTLSSetting_MIXED))
		})
		It("when there are no peers, pio is ignored and mtls is DISABLED", func() {
			pio := true
			mtlsState := evaluateMTlsForPeer(peersEmpty, pio)
			Expect(mtlsState).To(Equal(MTLSSetting_DISABLED))
		})
	})
})

var _ = Describe("paramIsMTls()", func() {
	Context("paramIsMTls()", func() {
		It("determines whether mtls is enabled for a Peer", func() {

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

type expectedValue struct {
	input       []*authv1alpha1.MeshPolicy
	mtlsEnabled bool
	err         error
}

var (
	defaultOn = &authv1alpha1.MeshPolicy{
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
					&istioauthv1alpha1.PeerAuthenticationMethod{
						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
					},
				},
			},
		},
	}

	defaultOff = &authv1alpha1.MeshPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MeshPolicy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
		Spec: authv1alpha1.MeshPolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{},
			},
		},
	}

	nonDefaultNamedOn = &authv1alpha1.MeshPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MeshPolicy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "named",
		},
		Spec: authv1alpha1.MeshPolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{
						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
					},
				},
			},
		},
	}
)

var _ = Describe("IsGlobalMtlsEnabled and MeshPolicyIsMtls", func() {
	var expectedValues = []expectedValue{
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{},
			mtlsEnabled: false,
			err:         nil,
		},
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{defaultOn, defaultOff},
			mtlsEnabled: false,
			err:         errors.New(""),
		},
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{defaultOn},
			mtlsEnabled: true,
			err:         nil,
		},
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{defaultOff},
			mtlsEnabled: false,
			err:         nil,
		},
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{nonDefaultNamedOn},
			mtlsEnabled: false,
			err:         errors.New(""),
		},
	}
	It("Should match the expected values", func() {
		for i := 0; i < len(expectedValues); i++ {
			mtlsEnabled, err := IsGlobalMtlsEnabled(expectedValues[i].input)
			if expectedValues[i].err != nil {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(mtlsEnabled).To(Equal(expectedValues[i].mtlsEnabled))
		}
	})

})
