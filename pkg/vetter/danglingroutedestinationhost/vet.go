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

// Package danglingroutedestinationhost vets if HTTP route destination host in
// any VirtualService resource points to services which don't exist in the
// cluster.
package danglingroutedestinationhost

import (
	"strings"

	v1alpha3 "github.com/aspenmesh/istio-client-go/pkg/apis/networking/v1alpha3"
	netv1alpha3 "github.com/aspenmesh/istio-client-go/pkg/client/listers/networking/v1alpha3"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/listers/core/v1"
)

const (
	vetterID                                = "DanglingRouteDestinationHost"
	danglingRouteDestinationHostNoteType    = "dangling-route-destination"
	danglingRouteDestinationHostNoteSummary = "Dangling route destination - ${vs_name}"
	danglingRouteDestinationHostNoteMsg     = "The VirtualService ${vs_name} in namespace ${namespace}" +
		" has route destination host(s) ${hostname_list} pointing to service(s)" +
		" which don't exist. Consider adding the services or removing the destination hosts" +
		" from the VirtualService resource."
)

// DanglingRouteDestinationHost implements Vetter interface
type DanglingRouteDestinationHost struct {
	nsLister  v1.NamespaceLister
	svcLister v1.ServiceLister
	vsLister  netv1alpha3.VirtualServiceLister
}

func createServiceMap(svcs []*corev1.Service) map[string]bool {
	serviceMap := map[string]bool{}
	for _, s := range svcs {
		key := s.Name + "." + s.Namespace + util.KubernetesDomainSuffix
		serviceMap[key] = true
	}
	return serviceMap
}

// createDanglingRouteHostNotes creates notes for VirtualService(s) which have
// dangling route hostname(s).
func createDanglingRouteHostNotes(svcs []*corev1.Service,
	vsList []*v1alpha3.VirtualService) []*apiv1.Note {
	var err error
	var host string
	notes := []*apiv1.Note{}
	svcMap := createServiceMap(svcs)
	for _, vs := range vsList {
		danglingHostnames := []string{}
		for _, routes := range vs.Spec.GetHttp() {
			for _, dw := range routes.GetRoute() {
				if d := dw.GetDestination(); d != nil {
					host = d.GetHost()
					if len(host) > 0 {
						host, err = util.ConvertHostnameToFQDN(host, vs.Namespace)
						if err == nil &&
							strings.HasSuffix(host, util.KubernetesDomainSuffix) {
							if _, ok := svcMap[host]; !ok {
								danglingHostnames = append(danglingHostnames, d.GetHost())
							}
						}
					}
				}
			}
		}
		if len(danglingHostnames) > 0 {
			notes = append(notes, &apiv1.Note{
				Type:    danglingRouteDestinationHostNoteType,
				Summary: danglingRouteDestinationHostNoteSummary,
				Msg:     danglingRouteDestinationHostNoteMsg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"vs_name":       vs.Name,
					"namespace":     vs.Namespace,
					"hostname_list": strings.Join(danglingHostnames, ","),
				},
			})
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}

	return notes
}

// Vet returns the list of generated notes
func (r *DanglingRouteDestinationHost) Vet() ([]*apiv1.Note, error) {
	svcs, err := util.ListServicesInMesh(r.nsLister, r.svcLister)
	if err != nil {
		return nil, err
	}

	vsList, err := util.ListVirtualServicesInMesh(r.nsLister, r.vsLister)
	if err != nil {
		return nil, err
	}

	notes := createDanglingRouteHostNotes(svcs, vsList)
	return notes, nil
}

// Info returns information about the vetter
func (r *DanglingRouteDestinationHost) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

// NewVetter returns "DanglingRouteDestinationHost" which implements Vetter Interface
func NewVetter(factory vetter.ResourceListGetter) *DanglingRouteDestinationHost {
	return &DanglingRouteDestinationHost{
		nsLister:  factory.K8s().Core().V1().Namespaces().Lister(),
		svcLister: factory.K8s().Core().V1().Services().Lister(),
		vsLister:  factory.Istio().Networking().V1alpha3().VirtualServices().Lister(),
	}
}
