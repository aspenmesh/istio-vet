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

// Package mtlsprobes inspects if Istio mTLS and liveness probes are enabled for
// any Pods in the mesh.
package mtlsprobes

import (
	"errors"

	authv1alpha1 "github.com/aspenmesh/istio-client-go/pkg/client/listers/authentication/v1alpha1"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	mtlspolicyutil "github.com/aspenmesh/istio-vet/pkg/vetter/util/mtlspolicy"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/listers/core/v1"
)

const (
	vetterID                 = "MtlsProbes"
	mtlsProbesNoteType       = "mtls-probes-incompatible"
	mtlsLivenessProbeSummary = "mTLS and liveness probe incompatible - ${pod_name}"
	mtlsLivenessProbeMsg     = "The pod ${pod_name} in namespace ${namespace} uses" +
		" liveness probe which is incompatible with mTLS. Consider disabling the" +
		" liveness probe or mTLS."
	mtlsReadinessProbeSummary = "mTLS and readiness probe incompatible - ${pod_name}"
	mtlsReadinessProbeMsg     = "The pod ${pod_name} in namespace ${namespace} uses" +
		" readiness probe which is incompatible with mTLS. Consider disabling the" +
		" readiness probe or mTLS."
)

// MtlsProbes implements Vetter interface
type MtlsProbes struct {
	podLister v1.PodLister
	nsLister  v1.NamespaceLister
	cmLister  v1.ConfigMapLister
	epLister  v1.EndpointsLister
	apLister  authv1alpha1.PolicyLister
	mpLister  authv1alpha1.MeshPolicyLister
}

// getPodEndpoint returns an Endpoint for a pod (if the endpoint exists)
func getPodEndpoint(endpointList []*corev1.Endpoints, pod *corev1.Pod) (*corev1.Endpoints, error) {
	if pod == nil {
		err := errors.New("pod cannot be nil")
		return nil, err
	}
	podEndpoints := []*corev1.Endpoints{}
	for _, ep := range endpointList {
		if ep.Namespace == pod.Namespace {
			for _, es := range ep.Subsets {
				for _, ea := range es.Addresses {
					if &ea != nil && &ea.TargetRef != nil {
						if ea.TargetRef.Kind == "Pod" && pod.Name == ea.TargetRef.Name {
							podEndpoints = append(podEndpoints, ep)
						}
					}
				}
			}
		}
	}
	if len(podEndpoints) == 1 {
		return podEndpoints[0], nil
	} else if len(podEndpoints) == 0 {
		return nil, nil
	} else {
		err := errors.New("multiple endpoints refer to the same pod")
		return nil, err
	}
}

// isNoteRequiredForMtlsProbe returns true if a note needs to be generated
// based off of auth policies related to the liveness/readiness probe
func isNoteRequiredForMtlsProbe(authPolicies *mtlspolicyutil.AuthPolicies, endpoint *corev1.Endpoints,
	probePort uint32, globalMtls bool) bool {
	// if the endpoint is nil, just return the status of globalMtls
	if endpoint == nil {
		return globalMtls
	}
	// create service
	var svc mtlspolicyutil.Service = mtlspolicyutil.Service{
		Name:      endpoint.Name,
		Namespace: endpoint.Namespace}
	mtls, _, err := authPolicies.TLSByPort(svc, probePort)
	if err != nil && err.Error() == "Use Mesh Policy" {
		// TLSByPort was refactored to check for a mesh policy as part of the AuthPolicy struct. If mTls for the mesh policy is determined separately, this is the catch.
		return globalMtls
	} else if err != nil {
		// TODO(BLaurenB): actually, an error here would mean that we couldn't determine the mtls state (likely conflicting policies). We should exit the function and throw an error or return false instead of allowing the vetter to write a note.
		// (m-eaton ?) no policies were found for port, name or namespace, return status of globalMtls
		return globalMtls
	} else {
		// policy was found, return the mTLS status of the policy
		return mtls
	}
}

// // isNoteRequiredForMtlsProbe returns true if a note needs to be generated
// // based off of auth policies related to the liveness/readiness probe
// func isNoteRequiredForMtlsProbe(authPolicies *mtlspolicyutil.AuthPolicies, endpoint *corev1.Endpoints,
// 	probePort uint32, globalMtls bool) bool {
// 	// if the endpoint is nil, just return the status of globalMtls
// 	if endpoint == nil {
// 		return globalMtls
// 	}
// 	// create service
// 	var svc mtlspolicyutil.Service = mtlspolicyutil.Service{
// 		Name:      endpoint.Name,
// 		Namespace: endpoint.Namespace}
// 	mtls, _, err := authPolicies.TLSByPort(svc, probePort)
// 	if err != nil {
// 		// no policies were found for port, name or namespace, return status of globalMtls
// 		return globalMtls
// 	} else {
// 		// policy was found, return the mTLS status of the policy
// 		return mtls
// 	}
// }

// Vet returns the list of generated notes
func (m *MtlsProbes) Vet() ([]*apiv1.Note, error) {
	var notes []*apiv1.Note
	pods, err := util.ListPodsInMesh(m.nsLister, m.cmLister, m.podLister)
	if err != nil {
		if n := util.IstioInitializerDisabledNote(err.Error(), vetterID,
			mtlsProbesNoteType); n != nil {
			notes = append(notes, n)
			return notes, nil
		}
		return nil, err
	}
	// get auth policies
	policyList, err := m.apLister.List(labels.Everything())
	if err != nil {
		glog.Errorln("Unable to retreive auth policies")
		return nil, err
	}
	authPolicies, err := mtlspolicyutil.LoadAuthPolicies(policyList)
	if err != nil {
		glog.Errorln("Unable to load auth policies")
		return nil, err
	}
	// get global mTLS policy
	meshPolicyList, err := m.mpLister.List(labels.Everything())
	if err != nil {
		glog.Errorf("Failed to retrieve MeshPolicies: %s", err)
		return nil, err
	}
	globalMtls, err := mtlspolicyutil.IsGlobalMtlsEnabled(meshPolicyList)
	if err != nil {
		glog.Errorln("Unable to determine status of global mTLS")
		return nil, err
	}
	// get list of endpoints
	endpointsList, err := util.ListEndpointsInMesh(m.nsLister, m.cmLister, m.epLister)
	if err != nil {
		glog.Errorln("unable to retrieve list of endpoints in the mesh")
		return nil, err
	}

	for _, p := range pods {
		if util.SidecarInjected(p) == true {
			cList := p.Spec.Containers
			for _, c := range cList {
				if (c.LivenessProbe != nil && c.LivenessProbe.Exec == nil) ||
					(c.ReadinessProbe != nil && c.ReadinessProbe.Exec == nil) {
					// get port for the probe
					var probePort intstr.IntOrString
					if c.LivenessProbe.Handler.HTTPGet != nil {
						probePort = c.LivenessProbe.Handler.HTTPGet.Port
					} else if c.LivenessProbe.Handler.TCPSocket != nil {
						probePort = c.LivenessProbe.Handler.TCPSocket.Port
					} else if c.ReadinessProbe.Handler.HTTPGet != nil {
						probePort = c.ReadinessProbe.Handler.HTTPGet.Port
					} else {
						probePort = c.ReadinessProbe.Handler.TCPSocket.Port
					}
					var intstrptr *intstr.IntOrString = &probePort
					probePortNum := uint32(intstrptr.IntValue())
					if probePortNum == 0 {
						// TODO(m-eaton): handle port names by finding the corresponding port
						// number
						glog.Errorln("Probe port is a name, skipping to next pod")
						continue
					} else if probePortNum > 65536 {
						glog.Errorln("Probe port number is out of range, skipping to next pod")
						continue
					} else {
						// get endpoint for the pod
						podEndpoint, err := getPodEndpoint(endpointsList, p)
						if err != nil {
							glog.Errorln("Error getting pod endpoint, skipping to next pod")
							continue
						}
						// check to see if mTLS needs to be disabled for the probe
						if generateNote := isNoteRequiredForMtlsProbe(authPolicies, podEndpoint, probePortNum, globalMtls); generateNote {
							if c.LivenessProbe != nil {
								notes = append(notes, &apiv1.Note{
									Type:    mtlsProbesNoteType,
									Summary: mtlsLivenessProbeSummary,
									Msg:     mtlsLivenessProbeMsg,
									Level:   apiv1.NoteLevel_ERROR,
									Attr: map[string]string{
										"pod_name":  p.Name,
										"namespace": p.Namespace}})
							} else if c.ReadinessProbe != nil {
								notes = append(notes, &apiv1.Note{
									Type:    mtlsProbesNoteType,
									Summary: mtlsReadinessProbeSummary,
									Msg:     mtlsReadinessProbeMsg,
									Level:   apiv1.NoteLevel_ERROR,
									Attr: map[string]string{
										"pod_name":  p.Name,
										"namespace": p.Namespace}})
							}
						}
					}
				}
			}
		}
	}
	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}
	return notes, nil
}

// Info returns information about the vetter
func (m *MtlsProbes) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

// NewVetter returns "mtlsProbes" which implements Vetter Interface
func NewVetter(factory vetter.ResourceListGetter) *MtlsProbes {
	return &MtlsProbes{
		podLister: factory.K8s().Core().V1().Pods().Lister(),
		cmLister:  factory.K8s().Core().V1().ConfigMaps().Lister(),
		nsLister:  factory.K8s().Core().V1().Namespaces().Lister(),
		epLister:  factory.K8s().Core().V1().Endpoints().Lister(),
		apLister:  factory.Istio().Authentication().V1alpha1().Policies().Lister(),
		mpLister:  factory.Istio().Authentication().V1alpha1().MeshPolicies().Lister(),
	}
}
