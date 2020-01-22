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

	metanetv1alpha3 "github.com/aspenmesh/istio-client-go/pkg/apis/networking/v1alpha3"
	netv1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	drFooOn = &metanetv1alpha3.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drFooOn",
			Namespace: "default",
		},
		Spec: metanetv1alpha3.DestinationRuleSpec{
			DestinationRule: netv1alpha3.DestinationRule{
				Host: "foo.default.svc.cluster.local",
				TrafficPolicy: &netv1alpha3.TrafficPolicy{
					Tls: &netv1alpha3.TLSSettings{
						Mode: netv1alpha3.TLSSettings_MUTUAL,
					},
				},
				Subsets: []*netv1alpha3.Subset{},
			},
		},
	}

	drBarOff = &metanetv1alpha3.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drBarOff",
			Namespace: "default",
		},
		Spec: metanetv1alpha3.DestinationRuleSpec{
			DestinationRule: netv1alpha3.DestinationRule{
				Host: "bar.default.svc.cluster.local",
				TrafficPolicy: &netv1alpha3.TrafficPolicy{
					Tls: &netv1alpha3.TLSSettings{
						Mode: netv1alpha3.TLSSettings_DISABLE,
					},
				},
				Subsets: []*netv1alpha3.Subset{},
			},
		},
	}

	drFooPortOnlyOn8443 = &netv1alpha3.TrafficPolicy_PortTrafficPolicy{
		Port: &netv1alpha3.PortSelector{
			Number: 8443,
		},
		Tls: &netv1alpha3.TLSSettings{
			Mode: netv1alpha3.TLSSettings_MUTUAL,
		},
	}

	drFooPortOnlyOn = &metanetv1alpha3.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drFooPortOnlyOn",
			Namespace: "default",
		},
		Spec: metanetv1alpha3.DestinationRuleSpec{
			DestinationRule: netv1alpha3.DestinationRule{
				Host: "foo.default.svc.cluster.local",
				TrafficPolicy: &netv1alpha3.TrafficPolicy{
					PortLevelSettings: []*netv1alpha3.TrafficPolicy_PortTrafficPolicy{
						drFooPortOnlyOn8443,
					},
				},
				Subsets: []*netv1alpha3.Subset{},
			},
		},
	}

	drFooPortOnlyOff8443 = &netv1alpha3.TrafficPolicy_PortTrafficPolicy{
		Port: &netv1alpha3.PortSelector{
			Number: 8443,
		},
		Tls: &netv1alpha3.TLSSettings{
			Mode: netv1alpha3.TLSSettings_DISABLE,
		},
	}

	drFooPortOnlyOff = &metanetv1alpha3.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drFooPortOnlyOff",
			Namespace: "default",
		},
		Spec: metanetv1alpha3.DestinationRuleSpec{
			DestinationRule: netv1alpha3.DestinationRule{
				Host: "foo.default.svc.cluster.local",
				TrafficPolicy: &netv1alpha3.TrafficPolicy{
					PortLevelSettings: []*netv1alpha3.TrafficPolicy_PortTrafficPolicy{
						drFooPortOnlyOff8443,
					},
				},
				Subsets: []*netv1alpha3.Subset{},
			},
		},
	}

	drDefaultNsOn = &metanetv1alpha3.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drDefaultNsFooPortOnlyOn",
			Namespace: "default",
		},
		Spec: metanetv1alpha3.DestinationRuleSpec{
			DestinationRule: netv1alpha3.DestinationRule{
				Host: "*.default.svc.cluster.local",
				TrafficPolicy: &netv1alpha3.TrafficPolicy{
					Tls: &netv1alpha3.TLSSettings{
						Mode: netv1alpha3.TLSSettings_DISABLE,
					},
				},
				Subsets: []*netv1alpha3.Subset{},
			},
		},
	}

	// This would turn on mTLS for all services in default namespace but only
	// on port 8443.  Weird, but allowed.
	drDefaultNsFooPortOnlyOn = &metanetv1alpha3.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drDefaultNsFooPortOnlyOn",
			Namespace: "default",
		},
		Spec: metanetv1alpha3.DestinationRuleSpec{
			DestinationRule: netv1alpha3.DestinationRule{
				Host: "*.default.svc.cluster.local",
				TrafficPolicy: &netv1alpha3.TrafficPolicy{
					PortLevelSettings: []*netv1alpha3.TrafficPolicy_PortTrafficPolicy{
						&netv1alpha3.TrafficPolicy_PortTrafficPolicy{
							Port: &netv1alpha3.PortSelector{
								Number: 8443,
							},
							Tls: &netv1alpha3.TLSSettings{
								Mode: netv1alpha3.TLSSettings_MUTUAL,
							},
						},
					},
				},
				Subsets: []*netv1alpha3.Subset{},
			},
		},
	}
)

var _ = Describe("LoadDestRules", func() {
	It("should load rules", func() {
		loaded, err := LoadDestRules([]*metanetv1alpha3.DestinationRule{
			drFooOn,
			drBarOff,
			drFooPortOnlyOn,
			drDefaultNsOn,
		})
		Expect(err).To(Succeed())
		Expect(loaded.ByNamespace("default")).To(Equal([]*metanetv1alpha3.DestinationRule{drDefaultNsOn}))
		Expect(loaded.ByName(Service{Name: "foo", Namespace: "default"})).To(Equal([]*metanetv1alpha3.DestinationRule{drFooOn}))
		Expect(loaded.ByName(Service{Name: "bar", Namespace: "default"})).To(Equal([]*metanetv1alpha3.DestinationRule{drBarOff}))
		Expect(loaded.ByPort(Service{Name: "foo", Namespace: "default"}, 8443)).To(Equal([]*PortDestRule{
			&PortDestRule{
				Rule:     drFooPortOnlyOn,
				PortRule: drFooPortOnlyOn8443,
			},
		}))
	})
})
