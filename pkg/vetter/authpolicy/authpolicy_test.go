package authpolicy

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sort"

	aspenv1a1 "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	istiov1alpha1 "istio.io/api/authentication/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func policyNameNoSvc(namespace, polName string) *aspenv1a1.Policy {
	return &aspenv1a1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      polName,
			Namespace: namespace,
		},
	}
}

func policyNameOneSvc(namespace, polName, service string) *aspenv1a1.Policy {
	return &aspenv1a1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      polName,
		},
		Spec: aspenv1a1.PolicySpec{
			Policy: istiov1alpha1.Policy{
				Targets: []*istiov1alpha1.TargetSelector{
					&istiov1alpha1.TargetSelector{
						Name: service,
					},
				},
			},
		},
	}
}
func policynameMultiSvcs(namespace, polName, svc1, svc2, svc3 string) *aspenv1a1.Policy {
	return &aspenv1a1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      polName,
		},
		Spec: aspenv1a1.PolicySpec{
			Policy: istiov1alpha1.Policy{
				Targets: []*istiov1alpha1.TargetSelector{
					&istiov1alpha1.TargetSelector{
						Name: svc1,
					},
					&istiov1alpha1.TargetSelector{
						Name: svc2,
					},
					&istiov1alpha1.TargetSelector{
						Name: svc3,
					},
				},
			},
		},
	}
}

func policynameMultiSvcsMultiPorts(namespace string, polName string, svc1 string, svc2 string, port1 uint32, port2 uint32) *aspenv1a1.Policy {
	return &aspenv1a1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      polName,
		},
		Spec: aspenv1a1.PolicySpec{
			Policy: istiov1alpha1.Policy{
				Targets: []*istiov1alpha1.TargetSelector{
					&istiov1alpha1.TargetSelector{
						Name: svc1,
						Ports: []*istiov1alpha1.PortSelector{
							&istiov1alpha1.PortSelector{
								Port: &istiov1alpha1.PortSelector_Number{Number: port1},
							},
						},
					},
					&istiov1alpha1.TargetSelector{
						Name: svc2,
						Ports: []*istiov1alpha1.PortSelector{
							&istiov1alpha1.PortSelector{
								Port: &istiov1alpha1.PortSelector_Number{Number: port2},
							},
						},
					},
				},
			},
		},
	}
}

func getEmptyAuthPolicies() []*aspenv1a1.Policy {
	return []*aspenv1a1.Policy{}
}

func sortNotes(notes []*apiv1.Note) []*apiv1.Note {
	sort.Slice(notes, func(i, j int) bool {
		if notes[i].Attr["namespace"] != notes[j].Attr["namespace"] {
			return notes[i].Attr["namespace"] < notes[j].Attr["namespace"]
		}
		if notes[i].Attr["policy_names"] != notes[j].Attr["policy_names"] {
			return notes[i].Attr["policy_names"] < notes[j].Attr["policy_names"]
		}
		if notes[i].Attr["target_service"] != notes[j].Attr["target_service"] {
			return notes[i].Attr["target_service"] < notes[j].Attr["target_service"]
		}
		return false
	})
	return notes
}

var _ = Describe("Authentication Policies", func() {

	Describe("Can evaluate the authentication policies for a service", func() {
		It("returns no notes for empty policies", func() {
			policies := getEmptyAuthPolicies()
			notes, err := notesForAuthPolicies(policies)
			Expect(err).To(BeNil())
			Expect(notes).To(HaveLen(0))
		})
		It("returns a note for same namespace, different policynames, no services", func() {
			pols := []*aspenv1a1.Policy{}
			a := policyNameNoSvc("namespace1", "name1")
			b := policyNameNoSvc("namespace1", "name2")
			pols = append(pols, a, b)
			notes, err := notesForAuthPolicies(pols)

			Expect(err).To(BeNil())
			Expect(notes).To(HaveLen(1))
			Expect(notes[0].Attr["namespace"]).To(Equal("namespace1"))
			Expect(notes[0].Attr["policy_names"]).To(Equal("name1, name2"))
		})
		It("returns no note for different namespace, same policynames, no services", func() {
			pols := []*aspenv1a1.Policy{}
			a := policyNameNoSvc("namespace1", "name2")
			b := policyNameNoSvc("namespace2", "name2") //added for noise
			pols = append(pols, a, b)
			notes, err := notesForAuthPolicies(pols)
			Expect(err).To(BeNil())
			Expect(notes).To(HaveLen(0))
		})
		It("returns no notes for same namespace, (diff policynames), different servicenames", func() {
			pols := []*aspenv1a1.Policy{}
			a := policyNameOneSvc("namespace1", "name1", "service1")
			b := policyNameOneSvc("namespace1", "name2", "service2")
			c := policyNameOneSvc("namespace1", "name3", "service3")
			pols = append(pols, a, b, c)
			notes, err := notesForAuthPolicies(pols)

			Expect(err).To(BeNil())
			Expect(notes).To(HaveLen(0))

		})
		It("returns a note for same namespace, (diff policynames), same servicename", func() {

			pols := []*aspenv1a1.Policy{}
			a := policyNameOneSvc("namespace1", "name1", "service1")
			b := policyNameOneSvc("namespace1", "name2", "service1")
			c := policyNameOneSvc("namespace1", "name3", "service3")
			d := policyNameNoSvc("namespace1", "name4") //added for noise
			pols = append(pols, a, b, c, d)
			notes, err := notesForAuthPolicies(pols)
			Expect(err).To(BeNil())
			Expect(notes).To(HaveLen(1))
			Expect(notes[0].Attr["policy_names"]).To(Equal("name1, name2"))
			Expect(notes[0].Attr["target_service"]).To(Equal("service1"))
		})
		It("returns the correct number of notes for policies with target_port", func() {
			pols := []*aspenv1a1.Policy{}
			a := policynameMultiSvcsMultiPorts("namespace1", "name1", "service1", "service2", 8000, 8001)
			b := policynameMultiSvcsMultiPorts("namespace1", "name2", "service1", "service3", 8000, 8003)
			c := policyNameOneSvc("namespace1", "name2", "service3") //added for noise
			pols = append(pols, a, b, c)
			notes, err := notesForAuthPolicies(pols)
			Expect(err).To(BeNil())
			Expect(notes).To(HaveLen(1))
			Expect(notes[0].Attr["target_port"]).To(Equal("8000"))
		})
		It("returns the correct notes with mixed policies", func() {
			pols := []*aspenv1a1.Policy{}
			a := policyNameOneSvc("namespace1", "name1", "service1")
			b := policynameMultiSvcs("namespace3", "name1", "service1", "service2", "service3")
			c := policyNameOneSvc("namespace1", "name3", "service3")
			d := policyNameNoSvc("namespace4", "name2")
			e := policyNameOneSvc("namespace2", "name2", "service2")
			f := policyNameNoSvc("namespace2", "name3")
			g := policynameMultiSvcs("namespace3", "name2", "service1", "service2", "service3")
			h := policyNameNoSvc("namespace3", "name3")
			i := policynameMultiSvcsMultiPorts("namespace5", "name1", "service1", "service2", 8002, 8001)
			j := policyNameOneSvc("namespace1", "name2", "service1")
			k := policyNameNoSvc("namespace4", "name1")
			l := policyNameOneSvc("namespace2", "name1", "service3")
			m := policynameMultiSvcsMultiPorts("namespace5", "name2", "service1", "service3", 8002, 8003)

			pols = append(pols, a, b, c, d, e, f, g, h, i, j, k, l, m)
			notes, err := notesForAuthPolicies(pols)
			notes = sortNotes(notes)

			Expect(err).To(BeNil())
			Expect(notes).To(HaveLen(6))

			Expect(notes[0].Type).To(Equal(authPolicyNoteTypeService))
			Expect(notes[0].Msg).To(Equal(authPolicyTargetSvcNameMsg))
			Expect(notes[0].Attr["namespace"]).To(Equal("namespace1"))
			Expect(notes[0].Attr["policy_names"]).To(Equal("name1, name2"))
			Expect(notes[0].Attr["target_service"]).To(Equal("service1"))

			Expect(notes[1].Type).To(Equal(authPolicyNoteTypeService))
			Expect(notes[1].Msg).To(Equal(authPolicyTargetSvcNameMsg))
			Expect(notes[1].Attr["namespace"]).To(Equal("namespace3"))
			Expect(notes[1].Attr["policy_names"]).To(Equal("name1, name2"))
			Expect(notes[1].Attr["target_service"]).To(Equal("service1"))

			Expect(notes[2].Type).To(Equal(authPolicyNoteTypeService))
			Expect(notes[2].Msg).To(Equal(authPolicyTargetSvcNameMsg))
			Expect(notes[2].Attr["namespace"]).To(Equal("namespace3"))
			Expect(notes[2].Attr["policy_names"]).To(Equal("name1, name2"))
			Expect(notes[2].Attr["target_service"]).To(Equal("service2"))

			Expect(notes[3].Type).To(Equal(authPolicyNoteTypeService))
			Expect(notes[3].Msg).To(Equal(authPolicyTargetSvcNameMsg))
			Expect(notes[3].Attr["namespace"]).To(Equal("namespace3"))
			Expect(notes[3].Attr["policy_names"]).To(Equal("name1, name2"))
			Expect(notes[3].Attr["target_service"]).To(Equal("service3"))

			Expect(notes[4].Type).To(Equal(authPolicyNoteTypeNamespace))
			Expect(notes[4].Msg).To(Equal(authPolicyNamespaceMsg))
			Expect(notes[4].Attr["namespace"]).To(Equal("namespace4"))
			Expect(notes[4].Attr["policy_names"]).To(Equal("name1, name2"))

			Expect(notes[5].Type).To(Equal(authPolicyNoteTypePorts))
			Expect(notes[5].Msg).To(Equal(authPolicySvcPortMsg))
			Expect(notes[5].Attr["namespace"]).To(Equal("namespace5"))
			Expect(notes[5].Attr["policy_names"]).To(Equal("name1, name2"))
			Expect(notes[5].Attr["target_port"]).To(Equal("8002"))
		})

	})

})
