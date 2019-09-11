/*
Copyright 2019 Aspen Mesh Authors.

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

package invalidserviceforjwtpolicy

import (
	"github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"

	authv1alpha1 "github.com/aspenmesh/istio-client-go/pkg/client/listers/authentication/v1alpha1"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"
	"strings"
)

const (
	vetterID                  = "InvalidServiceForJWTPolicy"
	invalidServiceNoteType    = "invalid-service-for-jwt-authentication-policy"
	invalidServiceNoteSummary = "Target services must have valid service port names"
	invalidServiceNoteMsg     = "The authentication policy '${policy}' in namespace '${namespace}' has a target of" +
		" service '${service_target}', which does not contain a valid port name. Port names must be 'http', 'http2', 'https'," +
		" or must be prefixed with 'http-', 'http2-', or 'https-'."
	noServiceNoteType    = "target-service-not-found-for-jwt-authentication-policy"
	noServiceNoteSummary = "The authentication policy target service was not found in namespace '${namespace}'"
	noServiceNoteMsg     = "The authentication policy '${policy}' in namespace '${namespace}' references the service" +
		" '${service_target}', which does not exist in namespace '${namespace}'."
)

// Vetter implements Vetter interface
type Vetter struct {
	nsLister  v1.NamespaceLister
	svcLister v1.ServiceLister
	authPolicyListener authv1alpha1.PolicyLister
}

func (v *Vetter) Vet()([]*apiv1.Note, error) {
	ns, err := v.nsLister.List(labels.Everything())
	if err != nil {
		glog.Error("Failed to retrieve namespaces: ", err)
		return nil, err
	}

	var notes []*apiv1.Note
	for _, namespace := range ns {
		nsAuthPolicies, err := v.authPolicyListener.Policies(namespace.Name).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to retrieve Authentication Policies for namespace: %s : %s", namespace.Name, err)
			return nil, err
		}

		nsServices, err := v.svcLister.Services(namespace.Name).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to retrieve Services for namespace: %s : %s", namespace.Name, err)
			return nil, err
		}
		nsServiceLookup := createServiceLookup(nsServices)

		for _, policy := range nsAuthPolicies {
			policyNotes := createAuthPolicyNotes(policy, nsServiceLookup)
			notes = append(notes, policyNotes...)
		}
	}

	return notes, nil
}

func createServiceLookup(services []*corev1.Service) map[string]*corev1.Service {
	serviceLookup := make(map[string]*corev1.Service)
	for _, s := range services {
		serviceLookup[strings.ToLower(s.Name)] = s
	}
	return serviceLookup
}

func createAuthPolicyNotes(policy *v1alpha1.Policy, nsServiceLookup map[string]*corev1.Service) []*apiv1.Note {
	var notes []*apiv1.Note
	for _, o := range policy.Spec.GetOrigins() {
		if o.GetJwt() != nil {
			for _, t := range policy.Spec.GetTargets() {
				targetSvc := nsServiceLookup[strings.ToLower(t.Name)]
				if targetSvc == nil {
					n := apiv1.Note{
						Type:    noServiceNoteType,
						Summary: noServiceNoteSummary,
						Msg:     noServiceNoteMsg,
						Level:   apiv1.NoteLevel_ERROR,
						Attr:    map[string]string{
							"policy": policy.Name,
							"namespace": policy.Namespace,
							"service_target": t.Name,
						},
					}
					notes = append(notes, &n)
					continue
				}

				targetSvcIsValid := servicePortsContainAValidName(targetSvc)
				if !targetSvcIsValid {
					n := apiv1.Note{
						Type:    invalidServiceNoteType,
						Summary: invalidServiceNoteSummary,
						Msg:     invalidServiceNoteMsg,
						Level:   apiv1.NoteLevel_ERROR,
						Attr:    map[string]string{
							"policy": policy.Name,
							"namespace": policy.Namespace,
							"service_target": targetSvc.Name,
						},
					}
					notes = append(notes, &n)
				}
			}
		}
	}
	return notes
}

func servicePortsContainAValidName(targetSvc *corev1.Service) bool {
	for _, p := range targetSvc.Spec.Ports {
		portName := strings.ToLower(p.Name)
		if strings.HasPrefix(portName, "http-") ||
			strings.HasPrefix(portName, "http2-") ||
			strings.HasPrefix(portName, "https-") ||
			portName == "http" || portName == "http2" ||portName == "https" {
			return true
		}
	}
	return false
}

// Info returns information about the vetter
func (v *Vetter) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

// NewVetter returns "svcPortPrefix" which implements Vetter Interface
func NewVetter(factory vetter.ResourceListGetter) *Vetter {
	return &Vetter{
		nsLister:  factory.K8s().Core().V1().Namespaces().Lister(),
		svcLister: factory.K8s().Core().V1().Services().Lister(),
		authPolicyListener: factory.Istio().Authentication().V1alpha1().Policies().Lister(),
	}
}
