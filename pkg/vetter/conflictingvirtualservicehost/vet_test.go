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
		vsHostSummary := "Multiple VirtualServices define the same host (${host}) and gateway (${gateway})"
		vsHostMsg := "The VirtualServices ${vs_names} define the same host (${host}) and gateway (${gateway}). A VirtualService must have a unique combination of host and gateway. Consider updating the VirtualServices to have unique hostname and gateway."

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

		var Vs8 *v1alpha3.VirtualService = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Vs8",
				Namespace: namespace,
			},
			Spec: v1alpha3.VirtualServiceSpec{
				VirtualService: istiov1alpha3.VirtualService{
					Hosts: []string{"foo.com", "*.com"}}}}

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

		It("Generates a note when 2 short host names are identical and in the same namespace", func() {
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
					"host":     "host2.bar.svc.cluster.local",
					"gateway":  "mesh",
					"vs_names": "Vs1.bar, Vs2.bar"}}
			expectedNote.Id = util.ComputeID(expectedNote)
			Expect(vsNotes[0]).To(Equal(expectedNote))
		})

		It("Generates a note when the same hostname is defined in 2 different namespaces", func() {
			vsList := []*v1alpha3.VirtualService{Vs4, Vs6}
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
					"vs_names": "Vs4.bar, Vs6.foo"}}
			expectedNote.Id = util.ComputeID(expectedNote)
			Expect(vsNotes[0]).To(Equal(expectedNote))
		})

		It("Generates multiple notes with the correct number of VirtualService names when there are multiple conflicts found", func() {
			vsList := []*v1alpha3.VirtualService{Vs1, Vs3, Vs4, Vs5, Vs6, Vs7, Vs8}
			vsNotes, err := CreateVirtualServiceNotes(vsList)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsNotes).To(HaveLen(2))
			expectedNote1 := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"host":     "foo.com",
					"gateway":  "mesh",
					"vs_names": "Vs4.bar, Vs6.foo, Vs8.bar"}}
			expectedNote2 := &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"host":     "*.com",
					"gateway":  "mesh",
					"vs_names": "Vs7.bar, Vs8.bar"}}
			expectedNote1.Id = util.ComputeID(expectedNote1)
			expectedNote2.Id = util.ComputeID(expectedNote2)
			sort.Slice(vsNotes, func(i, j int) bool {
				return vsNotes[i].Attr["host"] > vsNotes[j].Attr["host"]
			})
			Expect(vsNotes[0]).To(Equal(expectedNote1))
			Expect(vsNotes[1]).To(Equal(expectedNote2))
		})
	})
})
