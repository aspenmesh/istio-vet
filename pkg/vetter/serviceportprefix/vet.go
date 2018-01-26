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

// Package serviceportprefix vets the port names of the services in the mesh and
// generates notes if they are missing Istio recognized port protocol prefixes.
package serviceportprefix

import (
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	"k8s.io/client-go/listers/core/v1"
)

const (
	vetterID                  = "serviceportprefix"
	servicePortPrefixNoteType = "missing-service-port-prefix"
	servicePortPrefixSummary  = "Missing prefix in service - ${service_name}"
	servicePortPrefixMsg      = "The service ${service_name} in namespace ${namespace}" +
		" contains port names not prefixed with mesh supported protocols." +
		" Consider updating the service port name with one of the mesh recognized prefixes."
)

// SvcPortPrefix implements Vetter interface
type SvcPortPrefix struct {
	nsLister  v1.NamespaceLister
	cmLister  v1.ConfigMapLister
	svcLister v1.ServiceLister
}

// Vet returns the list of generated notes
func (m *SvcPortPrefix) Vet() ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}
	services, err := util.ListServicesInMesh(m.nsLister, m.cmLister, m.svcLister)
	if err != nil {
		if n := util.IstioInitializerDisabledNote(err.Error(), vetterID,
			servicePortPrefixNoteType); n != nil {
			notes = append(notes, n)
			return notes, nil
		}
		return nil, err
	}
	for _, s := range services {
		for _, p := range s.Spec.Ports {
			if p.Protocol != util.ServiceProtocolUDP &&
				util.ServicePortPrefixed(p.Name) == false {
				notes = append(notes, &apiv1.Note{
					Type:    servicePortPrefixNoteType,
					Summary: servicePortPrefixSummary,
					Msg:     servicePortPrefixMsg,
					Level:   apiv1.NoteLevel_WARNING,
					Attr: map[string]string{
						"service_name": s.Name,
						"namespace":    s.Namespace}})
			}
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}

	return notes, nil
}

// Info returns information about the vetter
func (m *SvcPortPrefix) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

// NewVetter returns "svcPortPrefix" which implements Vetter Interface
func NewVetter(factory vetter.ResourceListGetter) *SvcPortPrefix {
	return &SvcPortPrefix{
		nsLister:  factory.Core().V1().Namespaces().Lister(),
		cmLister:  factory.Core().V1().ConfigMaps().Lister(),
		svcLister: factory.Core().V1().Services().Lister(),
	}
}
