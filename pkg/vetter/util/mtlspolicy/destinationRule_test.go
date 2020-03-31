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

	istioNet "istio.io/api/networking/v1beta1"
	istioClientNet "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	drFooOn = &istioClientNet.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drFooOn",
			Namespace: "default",
		},
		Spec: istioNet.DestinationRule{
			Host: "foo.default.svc.cluster.local",
			TrafficPolicy: &istioNet.TrafficPolicy{
				Tls: &istioNet.TLSSettings{
					Mode: istioNet.TLSSettings_MUTUAL,
				},
			},
			Subsets: []*istioNet.Subset{},
		},
	}

	drBarOff = &istioClientNet.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drBarOff",
			Namespace: "default",
		},
		Spec: istioNet.DestinationRule{
			Host: "bar.default.svc.cluster.local",
			TrafficPolicy: &istioNet.TrafficPolicy{
				Tls: &istioNet.TLSSettings{
					Mode: istioNet.TLSSettings_DISABLE,
				},
			},
			Subsets: []*istioNet.Subset{},
		},
	}

	drFooPortOnlyOn8443 = &istioNet.TrafficPolicy_PortTrafficPolicy{
		Port: &istioNet.PortSelector{
			Number: 8443,
		},
		Tls: &istioNet.TLSSettings{
			Mode: istioNet.TLSSettings_MUTUAL,
		},
	}

	drFooPortOnlyOn = &istioClientNet.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drFooPortOnlyOn",
			Namespace: "default",
		},
		Spec: istioNet.DestinationRule{
			Host: "foo.default.svc.cluster.local",
			TrafficPolicy: &istioNet.TrafficPolicy{
				PortLevelSettings: []*istioNet.TrafficPolicy_PortTrafficPolicy{
					drFooPortOnlyOn8443,
				},
			},
			Subsets: []*istioNet.Subset{},
		},
	}

	drFooPortOnlyOff8443 = &istioNet.TrafficPolicy_PortTrafficPolicy{
		Port: &istioNet.PortSelector{
			Number: 8443,
		},
		Tls: &istioNet.TLSSettings{
			Mode: istioNet.TLSSettings_DISABLE,
		},
	}

	drFooPortOnlyOff = &istioClientNet.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drFooPortOnlyOff",
			Namespace: "default",
		},
		Spec: istioNet.DestinationRule{
			Host: "foo.default.svc.cluster.local",
			TrafficPolicy: &istioNet.TrafficPolicy{
				PortLevelSettings: []*istioNet.TrafficPolicy_PortTrafficPolicy{
					drFooPortOnlyOff8443,
				},
			},
			Subsets: []*istioNet.Subset{},
		},
	}

	drDefaultNsOn = &istioClientNet.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drDefaultNsFooPortOnlyOn",
			Namespace: "default",
		},
		Spec: istioNet.DestinationRule{
			Host: "*.default.svc.cluster.local",
			TrafficPolicy: &istioNet.TrafficPolicy{
				Tls: &istioNet.TLSSettings{
					Mode: istioNet.TLSSettings_DISABLE,
				},
			},
			Subsets: []*istioNet.Subset{},
		},
	}

	// This would turn on mTLS for all services in default namespace but only
	// on port 8443.  Weird, but allowed.
	drDefaultNsFooPortOnlyOn = &istioClientNet.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drDefaultNsFooPortOnlyOn",
			Namespace: "default",
		},
		Spec: istioNet.DestinationRule{
			Host: "*.default.svc.cluster.local",
			TrafficPolicy: &istioNet.TrafficPolicy{
				PortLevelSettings: []*istioNet.TrafficPolicy_PortTrafficPolicy{
					&istioNet.TrafficPolicy_PortTrafficPolicy{
						Port: &istioNet.PortSelector{
							Number: 8443,
						},
						Tls: &istioNet.TLSSettings{
							Mode: istioNet.TLSSettings_MUTUAL,
						},
					},
				},
			},
			Subsets: []*istioNet.Subset{},
		},
	}
)

var _ = Describe("LoadDestRules", func() {
	It("should load rules", func() {
		loaded, err := LoadDestRules([]*istioClientNet.DestinationRule{
			drFooOn,
			drBarOff,
			drFooPortOnlyOn,
			drDefaultNsOn,
		})
		Expect(err).To(Succeed())
		Expect(loaded.ByNamespace("default")).To(Equal([]*istioClientNet.DestinationRule{drDefaultNsOn}))
		Expect(loaded.ByName(Service{Name: "foo", Namespace: "default"})).To(Equal([]*istioClientNet.DestinationRule{drFooOn}))
		Expect(loaded.ByName(Service{Name: "bar", Namespace: "default"})).To(Equal([]*istioClientNet.DestinationRule{drBarOff}))
		Expect(loaded.ByPort(Service{Name: "foo", Namespace: "default"}, 8443)).To(Equal([]*PortDestRule{
			&PortDestRule{
				Rule:     drFooPortOnlyOn,
				PortRule: drFooPortOnlyOn8443,
			},
		}))
	})
})
