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

package conflictingvirtualservicehost

import (
	"sort"

	"github.com/aspenmesh/istio-client-go/pkg/apis/networking/v1alpha3"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Conflicting Virtual Service Host Vet Notes", func() {
	Context("With fake VirtualServices", func() {
		namespace := "bar"
		vsHostNoteType := "host-in-multiple-vs"
		vsHostSummary := "Multiple VirtualServices define the same host (${host}) and gateway (${gateway}) and conflict"
		vsHostMsg := "The VirtualServices ${vs_names} with routes ${routes}" +
			" define the same host (${host}) and gateway (${gateway}) and conflict. A VirtualService must have a unique combination of host and gateway or must not conflict." +
			" Consider updating the VirtualServices to have unique hostname and gateway or remove one of the conflicting rules."

		var Vs1 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs1",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"host1", "host2"}}}}

		var Vs2 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs2",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"host2"}}}}
		var Vs3 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs3",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"host3", "host4"}}}}

		var Vs4 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs4",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"foo.com"}}}}

		var Vs5 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs5",
				Namespace: "foo",
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"host1"}}}}

		var Vs6 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs6",
				Namespace: "foo",
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"foo.com"}}}}

		var Vs7 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs7",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"*.com"}}}}

		// var Vs8 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Name:      "Vs8",
		// 		Namespace: namespace,
		// 	},
		// Spec: v1alpha3.VirtualServiceSpec{
		// 	VirtualService: istiov1alpha3.VirtualService{
		// 		Hosts: []string{"foo.com", "*.com"}}}}

		var Vs9 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs9",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts:    []string{"host1"},
					Gateways: []string{"gateway1"}}}}
		var Vs10 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs10",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts:    []string{"host1", "host2"},
					Gateways: []string{"gateway2"}}}}
		var Vs11 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs11",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"host1", "host2"}}}}

		prefixRoute := istiov1alpha3.HTTPRoute{Name: "route1",
			Match: []*istiov1alpha3.HTTPMatchRequest{
				&istiov1alpha3.HTTPMatchRequest{
					Name: "",
					Uri:  &istiov1alpha3.StringMatch{MatchType: &istiov1alpha3.StringMatch_Prefix{Prefix: "/foo"}},
				},
			},
		}

		prefixRoute2Levels := istiov1alpha3.HTTPRoute{Name: "route1",
			Match: []*istiov1alpha3.HTTPMatchRequest{
				&istiov1alpha3.HTTPMatchRequest{
					Name: "",
					Uri:  &istiov1alpha3.StringMatch{MatchType: &istiov1alpha3.StringMatch_Prefix{Prefix: "/foo/bar"}},
				},
			},
		}

		exactRoute := istiov1alpha3.HTTPRoute{Name: "route1",
			Match: []*istiov1alpha3.HTTPMatchRequest{
				&istiov1alpha3.HTTPMatchRequest{
					Name: "",
					Uri:  &istiov1alpha3.StringMatch{MatchType: &istiov1alpha3.StringMatch_Exact{Exact: "/foo"}},
				},
			},
		}

		// exactRoute2Levels := istiov1alpha3.HTTPRoute{Name: "route1",
		// 	Match: []*istiov1alpha3.HTTPMatchRequest{
		// 		&istiov1alpha3.HTTPMatchRequest{
		// 			Name: "",
		// 			Uri:  &istiov1alpha3.StringMatch{MatchType: &istiov1alpha3.StringMatch_Exact{Exact: "/foo/bar"}},
		// 		},
		// 	},
		// }

		// exactRoute3Levels := istiov1alpha3.HTTPRoute{Name: "route1",
		// 	Match: []*istiov1alpha3.HTTPMatchRequest{
		// 		&istiov1alpha3.HTTPMatchRequest{
		// 			Name: "",
		// 			Uri:  &istiov1alpha3.StringMatch{MatchType: &istiov1alpha3.StringMatch_Exact{Exact: "/foo/bar/baz"}},
		// 		},
		// 	},
		// }

		regexRoute := istiov1alpha3.HTTPRoute{Name: "route1",
			Match: []*istiov1alpha3.HTTPMatchRequest{
				&istiov1alpha3.HTTPMatchRequest{
					Name: "",
					Uri:  &istiov1alpha3.StringMatch{MatchType: &istiov1alpha3.StringMatch_Regex{Regex: "/f*"}},
				},
			},
		}

		regexRoute2 := istiov1alpha3.HTTPRoute{Name: "route1",
			Match: []*istiov1alpha3.HTTPMatchRequest{
				&istiov1alpha3.HTTPMatchRequest{
					Name: "",
					Uri:  &istiov1alpha3.StringMatch{MatchType: &istiov1alpha3.StringMatch_Regex{Regex: "/b*"}},
				},
			},
		}

		It("Does not generate notes when passed an empty list of VirtualServices", func() {
			vsList := []*v1alpha3.VirtualService{}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(0))
		})

		It("Does not generate notes when all (short name) hosts are unique", func() {
			vsList := []*v1alpha3.VirtualService{Vs1, Vs3}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(0))
		})

		It("Does not generate notes when short host names are the same, but are in different namespaces", func() {
			vsList := []*v1alpha3.VirtualService{Vs1, Vs5}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(0))
		})

		It("Does not generate notes when host names are the same, but have different gateways", func() {
			vsList := []*v1alpha3.VirtualService{Vs9, Vs10}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(0))
		})

		It("Does not generate notes when a specific hostname and a similar hostname with a wildcard are defined", func() {
			vsList := []*v1alpha3.VirtualService{Vs6, Vs7}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(0))
		})

		It("Generates a note when 2 routes are identical and in the same namespace", func() {
			Vs1.Spec.Http = []*istiov1alpha3.HTTPRoute{&exactRoute}
			Vs2.Spec.Http = []*istiov1alpha3.HTTPRoute{&exactRoute}
			vsList := []*v1alpha3.VirtualService{Vs1, Vs2}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(1))

			expectedNote := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"vs_names": "Vs1.bar, Vs2.bar",
					"host":     "host2.bar.svc.cluster.local",
					"gateway":  "mesh",
					"routes":   "/foo exact /foo exact",
				}}
			expectedNote.Id = util.ComputeID(expectedNote)
			Expect(vsNotes[0]).To(Equal(expectedNote))
		})

		It("Generates a note when regex conflicts with another route", func() {
			vsList := []*v1alpha3.VirtualService{Vs4, Vs6}
			Vs4.Spec.Http = []*istiov1alpha3.HTTPRoute{&regexRoute}
			Vs6.Spec.Http = []*istiov1alpha3.HTTPRoute{&prefixRoute2Levels}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(1))
			expectedNote := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"host":     "foo.com",
					"gateway":  "mesh",
					"routes":   "/f* regex /foo/bar prefix",
					"vs_names": "Vs4.bar, Vs6.foo"}}
			expectedNote.Id = util.ComputeID(expectedNote)
			Expect(vsNotes[0]).To(Equal(expectedNote))
		})

		It("Does not generate notes when there is more than one regex", func() {
			vsList := []*v1alpha3.VirtualService{Vs4, Vs6}
			Vs4.Spec.Http = []*istiov1alpha3.HTTPRoute{&regexRoute, &regexRoute2}
			Vs6.Spec.Http = []*istiov1alpha3.HTTPRoute{&prefixRoute2Levels}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(0))
		})

		It("Generates multiple notes with the correct number of VirtualService names when there are multiple conflicts found", func() {
			Vs1.Spec.Http = []*istiov1alpha3.HTTPRoute{&exactRoute, &prefixRoute}
			Vs2.Spec.Http = []*istiov1alpha3.HTTPRoute{&prefixRoute2Levels, &exactRoute}
			Vs11.Spec.Http = []*istiov1alpha3.HTTPRoute{&prefixRoute}
			vsList := []*v1alpha3.VirtualService{Vs1, Vs2, Vs11}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(8))
			expectedNote1 := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"routes":   "/foo exact /foo exact",
					"vs_names": "Vs1.bar, Vs2.bar",
					"host":     "host2.bar.svc.cluster.local",
					"gateway":  "mesh",
				}}
			expectedNote2 := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"gateway":  "mesh",
					"routes":   "/foo prefix /foo prefix",
					"vs_names": "Vs1.bar, Vs11.bar",
					"host":     "host1.bar.svc.cluster.local host1.bar.svc.cluster.local host2.bar.svc.cluster.local host2.bar.svc.cluster.local",
				}}
			expectedNote3 := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"vs_names": "Vs1.bar, Vs2.bar",
					"host":     "host2.bar.svc.cluster.local host2.bar.svc.cluster.local",
					"gateway":  "mesh",
					"routes":   "/foo exact /foo exact",
				}}
			expectedNote4 := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"host":     "host2.bar.svc.cluster.local host2.bar.svc.cluster.local",
					"gateway":  "mesh",
					"routes":   "/foo prefix /foo exact",
					"vs_names": "Vs1.bar, Vs2.bar",
				}}
			expectedNote5 := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"gateway":  "mesh",
					"routes":   "/foo exact /foo prefix",
					"vs_names": "Vs2.bar, Vs11.bar",
					"host":     "host2.bar.svc.cluster.local host2.bar.svc.cluster.local",
				}}
			expectedNote1.Id = util.ComputeID(expectedNote1)
			expectedNote2.Id = util.ComputeID(expectedNote2)
			expectedNote3.Id = util.ComputeID(expectedNote3)
			expectedNote4.Id = util.ComputeID(expectedNote4)
			expectedNote5.Id = util.ComputeID(expectedNote5)
			sort.Slice(vsNotes, func(i, j int) bool {
				return vsNotes[i].Attr["host"] > vsNotes[j].Attr["host"]
			})
			expecteds := []*apiv1.Note{expectedNote1, expectedNote2, expectedNote3, expectedNote4, expectedNote5}
			for _, note := range vsNotes {
				found := false
				for _, expected := range expecteds {
					if note == expected {
						found = true
					}
					if !found {
						Expect(false)
					}
				}
			}
			Expect(true)
		})
	})
})
