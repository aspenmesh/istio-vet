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
	"fmt"
	"strings"

	v1alpha3 "github.com/aspenmesh/istio-client-go/pkg/apis/networking/v1alpha3"
	netv1alpha3 "github.com/aspenmesh/istio-client-go/pkg/client/listers/networking/v1alpha3"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	"k8s.io/client-go/listers/core/v1"
)

const (
	defaultGateway = "mesh"
	vetterID       = "ConflictingVirtualServiceHost"
	vsHostNoteType = "host-in-multiple-vs"
	vsHostSummary  = "Multiple VirtualServices define the same host (${host}) and gateway (${gateway})"
	vsHostMsg      = "The VirtualServices ${vs_names}" +
		" define the same host (${host}) and gateway (${gateway}). A VirtualService must have a unique combination of host and gateway." +
		" Consider updating the VirtualServices to have unique hostname and gateway."
)

// VsHost implements Vetter interface
type VsHost struct {
	nsLister v1.NamespaceLister
	vsLister netv1alpha3.VirtualServiceLister
}
type hostAndGateway struct {
	gateway  string
	hostname string
}

type VirtualSvcByHostAndGateway map[hostAndGateway][]*v1alpha3.VirtualService

// CreateVirtualServiceNotes checks for multiple vs defining the same host and
// generates notes for these cases
func CreateVirtualServiceNotes(virtualServices []*v1alpha3.VirtualService) ([]*apiv1.Note, error) {
	vsByHostAndGateway := VirtualSvcByHostAndGateway{}
	for _, vs := range virtualServices {
		for _, host := range vs.Spec.GetHosts() {
			h, err := util.ConvertHostnameToFQDN(host, vs.Namespace)
			if err != nil {
				fmt.Printf("Unable to convert hostname: %s\n", err.Error())
				return nil, err
			}

			// One VS can have multiple hosts and gateways. Make 1 key per
			// combination.
			hg := hostAndGateway{hostname: h}
			if len(vs.Spec.GetGateways()) > 0 {
				for _, g := range vs.Spec.GetGateways() {
					hg.gateway = g
					populateVirtualServiceMap(hg, vs, vsByHostAndGateway)
				}
			} else {
				hg.gateway = defaultGateway
				populateVirtualServiceMap(hg, vs, vsByHostAndGateway)
			}
		}
	}

	// create vet notes
	notes := []*apiv1.Note{}
	for key, vsList := range vsByHostAndGateway {
		if len(vsList) > 1 {
			// there are multiple vs defining a host
			vsNames := []string{}
			for _, vs := range vsList {
				vsName := vs.Name + "." + vs.Namespace
				vsNames = append(vsNames, vsName)

			}
			notes = append(notes, &apiv1.Note{
				Type:    vsHostNoteType,
				Summary: vsHostSummary,
				Msg:     vsHostMsg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"host":     key.hostname,
					"gateway":  key.gateway,
					"vs_names": strings.Join(vsNames, ", ")}})
		}
	}
	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}
	return notes, nil
}

func populateVirtualServiceMap(hg hostAndGateway, vs *v1alpha3.VirtualService, vsByHostAndGateway VirtualSvcByHostAndGateway) {
	if _, ok := vsByHostAndGateway[hg]; !ok {
		vsByHostAndGateway[hg] = []*v1alpha3.VirtualService{vs}
	} else {
		vsByHostAndGateway[hg] = append(vsByHostAndGateway[hg], vs)
	}
}

// Vet returns the list of generated notes
func (v *VsHost) Vet() ([]*apiv1.Note, error) {
	virtualServices, err := util.ListVirtualServicesInMesh(v.nsLister, v.vsLister)
	if err != nil {
		fmt.Printf("Error occurred retrieving VirtualServices: %s\n", err.Error())
		return nil, err
	}
	notes, err := CreateVirtualServiceNotes(virtualServices)
	if err != nil {
		fmt.Printf("Error creating Conflicting VirtualService notes: %s\n", err.Error())
		return nil, err
	}
	return notes, nil
}

// Info returns information about the vetter
func (v *VsHost) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

// NewVetter returns "VsHost" which implements the Vetter Tnterface
func NewVetter(factory vetter.ResourceListGetter) *VsHost {
	return &VsHost{
		nsLister: factory.K8s().Core().V1().Namespaces().Lister(),
		vsLister: factory.Istio().Networking().V1alpha3().VirtualServices().Lister(),
	}
}
