/*
Copyright 2017 Aspen Mesh Authors.

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

package danglingroutedestinationhost

import (
	v1alpha3 "github.com/aspenmesh/istio-client-go/pkg/apis/networking/v1alpha3"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Vet", func() {
	It("creates zero notes on empty lists", func() {
		notes := createDanglingRouteHostNotes(nil, nil)
		Expect(notes).To(HaveLen(0))
	})

	It("creates zero notes if all hosts exist as services", func() {
		svcs := []*corev1.Service{
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "team-foo",
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "team-bar",
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "baz",
					Namespace: "team-baz",
				},
			},
		}
		vsList := []*v1alpha3.VirtualService{
			&v1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "team-baz",
				},
				Spec: v1alpha3.VirtualServiceSpec{
					VirtualService: istiov1alpha3.VirtualService{
						Http: []*istiov1alpha3.HTTPRoute{
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "foo.team-foo.svc.cluster.local",
										},
									},
								},
							},
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "foo.com",
										},
									},
								},
							},
						},
					},
				},
			},
			&v1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "team-bar",
				},
				Spec: v1alpha3.VirtualServiceSpec{
					VirtualService: istiov1alpha3.VirtualService{
						Http: []*istiov1alpha3.HTTPRoute{
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "bar",
										},
									},
								},
							},
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "baz.team-baz.svc.cluster.local",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		notes := createDanglingRouteHostNotes(svcs, vsList)
		Expect(notes).To(HaveLen(0))
	})

	It("creates notes if services don't exist for hosts", func() {
		svcs := []*corev1.Service{
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "team-foo",
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "team-bar",
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "baz",
					Namespace: "team-baz",
				},
			},
		}
		vsList := []*v1alpha3.VirtualService{
			&v1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "team-baz",
				},
				// Ignores FQDN hostnames
				Spec: v1alpha3.VirtualServiceSpec{
					VirtualService: istiov1alpha3.VirtualService{
						Http: []*istiov1alpha3.HTTPRoute{
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "foo.team-foo",
										},
									},
								},
							},
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "foo.com",
										},
									},
								},
							},
						},
					},
				},
			},
			&v1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "team-bar",
				},
				// Generates notes if FQDN ending with .svc.cluster.local doesn't match
				// or shortnames after expansion don't match services in the registry
				Spec: v1alpha3.VirtualServiceSpec{
					VirtualService: istiov1alpha3.VirtualService{
						Http: []*istiov1alpha3.HTTPRoute{
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "bar.team-baz.svc.cluster.local",
										},
									},
								},
							},
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "baz",
										},
									},
								},
							},
						},
					},
				},
			},
			// Tests that dangling hosts in multiple VirtualService(s) generates
			// multiple vet notes
			&v1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bah",
					Namespace: "team-bah",
				},
				// Generates notes when no service exist
				Spec: v1alpha3.VirtualServiceSpec{
					VirtualService: istiov1alpha3.VirtualService{
						Http: []*istiov1alpha3.HTTPRoute{
							&istiov1alpha3.HTTPRoute{
								Route: []*istiov1alpha3.HTTPRouteDestination{
									&istiov1alpha3.HTTPRouteDestination{
										Destination: &istiov1alpha3.Destination{
											Host: "bah.team-bah.svc.cluster.local",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		expNotes := []*apiv1.Note{
			&apiv1.Note{
				Type:    danglingRouteDestinationHostNoteType,
				Summary: danglingRouteDestinationHostNoteSummary,
				Msg:     danglingRouteDestinationHostNoteMsg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"vs_name":       "bar",
					"namespace":     "team-bar",
					"hostname_list": "bar.team-baz.svc.cluster.local,baz",
				},
			},
			&apiv1.Note{
				Type:    danglingRouteDestinationHostNoteType,
				Summary: danglingRouteDestinationHostNoteSummary,
				Msg:     danglingRouteDestinationHostNoteMsg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"vs_name":       "bah",
					"namespace":     "team-bah",
					"hostname_list": "bah.team-bah.svc.cluster.local",
				},
			},
		}
		for i := range expNotes {
			expNotes[i].Id = util.ComputeID(expNotes[i])
		}
		notes := createDanglingRouteHostNotes(svcs, vsList)
		Expect(notes).To(Equal(expNotes))
	})
})
