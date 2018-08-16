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

package mtlsprobes

import (
	authv1alpha1api "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	mtlspolicyutil "github.com/aspenmesh/istio-vet/pkg/vetter/util/mtlspolicy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istiov1alpha1 "istio.io/api/authentication/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get an Endpoint Address that refers to a pod", func() {
	Context("With fake Endpoints", func() {

		var namespace string = "default"

		var Pod1 *corev1.Pod = &corev1.Pod{
			metav1.TypeMeta{},
			metav1.ObjectMeta{
				Name:         "Pod1",
				Namespace:    namespace,
				Initializers: &metav1.Initializers{}},
			corev1.PodSpec{},
			corev1.PodStatus{}}

		var Pod1EpAddress corev1.EndpointAddress = corev1.EndpointAddress{
			IP: "00.00.000.000",
			TargetRef: &corev1.ObjectReference{
				Kind:      "Pod",
				Namespace: namespace,
				Name:      "Pod1"}}

		var RandomPodEpAddress corev1.EndpointAddress = corev1.EndpointAddress{
			IP: "11.11.111.111",
			TargetRef: &corev1.ObjectReference{
				Kind:      "Pod",
				Namespace: namespace,
				Name:      "RandomPod"}}

		var nilEndpoint *corev1.Endpoints = nil

		// this endpoint only has pod1
		var Endpoint1 *corev1.Endpoints = &corev1.Endpoints{
			metav1.TypeMeta{},
			metav1.ObjectMeta{
				Namespace:    namespace,
				Initializers: &metav1.Initializers{}},
			[]corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{Pod1EpAddress}}}}

		// this endpoint only has random pod
		var Endpoint2 *corev1.Endpoints = &corev1.Endpoints{
			metav1.TypeMeta{},
			metav1.ObjectMeta{
				Namespace:    namespace,
				Initializers: &metav1.Initializers{}},
			[]corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{RandomPodEpAddress}}}}

		// this endpoint has pod1 and random pod
		var Endpoint3 *corev1.Endpoints = &corev1.Endpoints{
			metav1.TypeMeta{},
			metav1.ObjectMeta{
				Namespace:    namespace,
				Initializers: &metav1.Initializers{}},
			[]corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{RandomPodEpAddress, Pod1EpAddress}}}}

		// this endpoint has pod1 and random pod, but in separate subsets
		var Endpoint4 *corev1.Endpoints = &corev1.Endpoints{
			metav1.TypeMeta{},
			metav1.ObjectMeta{
				Namespace:    namespace,
				Initializers: &metav1.Initializers{}},
			[]corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{RandomPodEpAddress}},
				{
					Addresses: []corev1.EndpointAddress{Pod1EpAddress}}}}

		It("Returns nil when an empty endpoint list is passed", func() {
			endpointList := []*corev1.Endpoints{}
			pod := Pod1
			podEndpoint, err := getPodEndpoint(endpointList, pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(podEndpoint).To(Equal(nilEndpoint))
		})

		It("Returns an error when a nil pod is passed", func() {
			endpointList := []*corev1.Endpoints{Endpoint2}
			var pod *corev1.Pod = nil
			podEndpoint, err := getPodEndpoint(endpointList, pod)
			Expect(err).To(HaveOccurred())
			Expect(podEndpoint).To(Equal(nilEndpoint))
		})

		It("Returns nil when none of the endpoints refer to the pod", func() {
			endpointList := []*corev1.Endpoints{Endpoint2}
			pod := Pod1
			podEndpoint, err := getPodEndpoint(endpointList, pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(podEndpoint).To(Equal(nilEndpoint))
		})

		It("Returns the correct pod endpoint when there is a list of 2 endpoints, and one endpoint refers to the pod",
			func() {
				endpointList := []*corev1.Endpoints{Endpoint1, Endpoint2}
				pod := Pod1
				podEndpoint, err := getPodEndpoint(endpointList, pod)
				Expect(err).NotTo(HaveOccurred())
				Expect(podEndpoint).To(Equal(Endpoint1))
			})

		It("Returns the correct pod endpoint when multiple addresses are listed, and one endpoint refers to the pod",
			func() {
				endpointList := []*corev1.Endpoints{Endpoint3}
				pod := Pod1
				podEndpoint, err := getPodEndpoint(endpointList, pod)
				Expect(err).NotTo(HaveOccurred())
				Expect(podEndpoint).To(Equal(Endpoint3))
			})

		It("Returns the correct pod endpoint when the endpoint is in different subsets, and one endpoint refers to the pod",
			func() {
				endpointList := []*corev1.Endpoints{Endpoint4}
				pod := Pod1
				podEndpoint, err := getPodEndpoint(endpointList, pod)
				Expect(err).NotTo(HaveOccurred())
				Expect(podEndpoint).To(Equal(Endpoint4))
			})

		It("Returns an error when there is a list of 2 endpoints, and more than one endpoint refers to the pod", func() {
			endpointList := []*corev1.Endpoints{Endpoint1, Endpoint3}
			pod := Pod1
			podEndpoint, err := getPodEndpoint(endpointList, pod)
			Expect(err).To(HaveOccurred())
			Expect(podEndpoint).To(Equal(nilEndpoint))
		})
	})
})

// TODO(m-eaton): re-format the tests by creating a list of inputs and expected
// outputs to make it easier to read
var _ = Describe("Know when mTLS is correctly configured for a liveness/readiness probe",
	func() {
		Context("With fake auth policies", func() {
			var namespace string = "default"
			var globalMtls bool
			var policyList []*authv1alpha1api.Policy
			var authPolicies *mtlspolicyutil.AuthPolicies
			var generateNote bool
			var err error
			var probePort1 uint32 = 1234

			var nilEndpoint *corev1.Endpoints = nil

			var enableMtls *istiov1alpha1.PeerAuthenticationMethod = &istiov1alpha1.PeerAuthenticationMethod{
				Params: &istiov1alpha1.PeerAuthenticationMethod_Mtls{}}

			// this is a port-specific policy to disable mTLS (for probePort1)
			var Policy1 *authv1alpha1api.Policy = &authv1alpha1api.Policy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Policy",
					APIVersion: "authentication.istio.io/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:         "Policy1",
					Namespace:    namespace,
					Initializers: &metav1.Initializers{}},
				Spec: authv1alpha1api.PolicySpec{
					Policy: istiov1alpha1.Policy{
						Targets: []*istiov1alpha1.TargetSelector{
							&istiov1alpha1.TargetSelector{
								Name: "Foo",
								Ports: []*istiov1alpha1.PortSelector{
									&istiov1alpha1.PortSelector{
										Port: &istiov1alpha1.PortSelector_Number{probePort1}}}}},
						Peers: []*istiov1alpha1.PeerAuthenticationMethod{}}}}

			// this is a port-specific policy to enable mTLS (for probePort1)
			var Policy2 *authv1alpha1api.Policy = &authv1alpha1api.Policy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Policy",
					APIVersion: "authentication.istio.io/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:         "Policy2",
					Namespace:    namespace,
					Initializers: &metav1.Initializers{}},
				Spec: authv1alpha1api.PolicySpec{
					Policy: istiov1alpha1.Policy{
						Targets: []*istiov1alpha1.TargetSelector{
							&istiov1alpha1.TargetSelector{
								Name: "Foo",
								Ports: []*istiov1alpha1.PortSelector{
									&istiov1alpha1.PortSelector{
										Port: &istiov1alpha1.PortSelector_Number{probePort1}}}}},
						Peers: []*istiov1alpha1.PeerAuthenticationMethod{enableMtls}}}}

			// this is a name-specific policy to disable mTLS for Foo
			var Policy3 *authv1alpha1api.Policy = &authv1alpha1api.Policy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Policy",
					APIVersion: "authentication.istio.io/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:         "Policy3",
					Namespace:    namespace,
					Initializers: &metav1.Initializers{}},
				Spec: authv1alpha1api.PolicySpec{
					Policy: istiov1alpha1.Policy{
						Targets: []*istiov1alpha1.TargetSelector{
							&istiov1alpha1.TargetSelector{
								Name:  "Foo",
								Ports: []*istiov1alpha1.PortSelector{}}},
						Peers: []*istiov1alpha1.PeerAuthenticationMethod{}}}}

			// this is a name-specific policy to enable mTLS for Foo
			var Policy4 *authv1alpha1api.Policy = &authv1alpha1api.Policy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Policy",
					APIVersion: "authentication.istio.io/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:         "Policy4",
					Namespace:    namespace,
					Initializers: &metav1.Initializers{}},
				Spec: authv1alpha1api.PolicySpec{
					Policy: istiov1alpha1.Policy{
						Targets: []*istiov1alpha1.TargetSelector{
							&istiov1alpha1.TargetSelector{
								Name:  "Foo",
								Ports: []*istiov1alpha1.PortSelector{}}},
						Peers: []*istiov1alpha1.PeerAuthenticationMethod{enableMtls}}}}

			// this is a namespace-specific policy to disable mTLS for default
			var Policy5 *authv1alpha1api.Policy = &authv1alpha1api.Policy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Policy",
					APIVersion: "authentication.istio.io/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:         "default",
					Namespace:    namespace,
					Initializers: &metav1.Initializers{}},
				Spec: authv1alpha1api.PolicySpec{
					Policy: istiov1alpha1.Policy{
						Targets: []*istiov1alpha1.TargetSelector{},
						Peers:   []*istiov1alpha1.PeerAuthenticationMethod{}}}}

			// this is a namespace-specific policy to enable mTLS for default
			var Policy6 *authv1alpha1api.Policy = &authv1alpha1api.Policy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Policy",
					APIVersion: "authentication.istio.io/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:         "default",
					Namespace:    namespace,
					Initializers: &metav1.Initializers{}},
				Spec: authv1alpha1api.PolicySpec{
					Policy: istiov1alpha1.Policy{
						Targets: []*istiov1alpha1.TargetSelector{},
						Peers:   []*istiov1alpha1.PeerAuthenticationMethod{enableMtls}}}}

			var Endpoint1 *corev1.Endpoints = &corev1.Endpoints{
				metav1.TypeMeta{},
				metav1.ObjectMeta{
					Name:         "Foo",
					Namespace:    namespace,
					Initializers: &metav1.Initializers{}},
				[]corev1.EndpointSubset{}}

			It("Returns false when an empty auth policy list is passed and global mTLS is disabled",
				func() {
					policyList = []*authv1alpha1api.Policy{}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = false
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(false))
				})

			It("Returns false when endpoint is nil and global mTLS is disabled", func() {
				policyList = []*authv1alpha1api.Policy{Policy1}
				authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
				Expect(err).NotTo(HaveOccurred())
				globalMtls = false
				generateNote = isNoteRequiredForMtlsProbe(authPolicies, nilEndpoint, probePort1, globalMtls)
				Expect(generateNote).To(Equal(false))

			})

			It("Returns true when an empty auth policy list is passed and global mTLS is enabled",
				func() {
					// empty policy list
					policyList = []*authv1alpha1api.Policy{}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = true
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
				})

			It("Returns true when endpoint is nil and global mTLS is disabled", func() {
				policyList = []*authv1alpha1api.Policy{Policy1}
				authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
				Expect(err).NotTo(HaveOccurred())
				globalMtls = true
				generateNote = isNoteRequiredForMtlsProbe(authPolicies, nilEndpoint, probePort1, globalMtls)
				Expect(generateNote).To(Equal(true))

			})

			It("Returns false when a policy at the port level disables mTLS for the probe port",
				func() {
					// global mTLS enabled
					policyList = []*authv1alpha1api.Policy{Policy1}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = true
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(false))
					// global mTLS disabled
					globalMtls = false
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(false))
				})

			It("Returns true when a policy at the port level enables mTLS for the probe port",
				func() {
					// global mTLS enabled
					policyList = []*authv1alpha1api.Policy{Policy2}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = true
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
					// global mTLS disabled
					globalMtls = false
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
				})

			It("Returns false when a policy at the name level disables mTLS for the service",
				func() {
					// global mTLS enabled
					policyList = []*authv1alpha1api.Policy{Policy3}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = true
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(false))
					// global mTLS disabled
					globalMtls = false
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(false))
				})

			It("Returns true when a policy at the name level enables mTLS for the service",
				func() {
					// global mTLS enabled
					policyList = []*authv1alpha1api.Policy{Policy4}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = true
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
					// global mTLS disabled
					globalMtls = false
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
				})

			It("Returns false when a policy at the namespace level disables mTLS for the default namespace",
				func() {
					// global mTLS enabled
					policyList = []*authv1alpha1api.Policy{Policy5}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = true
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(false))
					// global mTLS disabled
					globalMtls = false
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(false))
				})

			It("Returns false when a policy at the namespace level enables mTLS for the default namespace",
				func() {
					// global mTLS enabled
					policyList = []*authv1alpha1api.Policy{Policy6}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = true
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
					// global mTLS disabled
					globalMtls = false
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
				})

			It("Returns true when a policy disables mTLS at the name level, but enables it for the liveness/readiness probe port",
				func() {
					// global mTLS enabled
					policyList = []*authv1alpha1api.Policy{Policy3, Policy2}
					authPolicies, err = mtlspolicyutil.LoadAuthPolicies(policyList)
					Expect(err).NotTo(HaveOccurred())
					globalMtls = true
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
					// global mTLS disabled
					globalMtls = false
					generateNote = isNoteRequiredForMtlsProbe(authPolicies, Endpoint1, probePort1, globalMtls)
					Expect(generateNote).To(Equal(true))
				})
		})
	})
