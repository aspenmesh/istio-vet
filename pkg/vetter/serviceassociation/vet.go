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

// Package serviceassociation vets multiple service associations of pods in the
// mesh.
package serviceassociation

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
)

const (
	vetterId                               = "serviceassociation"
	multiple_service_association_note_type = "multiple-service-association"
	multiple_service_association_summary   = "Multiple service association - ${service_list}"
	multiple_service_association_msg       = "The services ${service_list} in namespace ${namespace}" +
		" are associated with the pod ${pod_name}. Consider updating the" +
		" service definitions ensuring the pod belongs to a single service."
)

type svcAssociation struct {
	info apiv1.Info
}

type endpointInfo struct {
	Namespace    string
	PodName      string
	ServiceNames []string
}

func createEndpointMap(e []corev1.Endpoints) map[string]endpointInfo {
	endpointMap := map[string]endpointInfo{}
	for _, ep := range e {
		for _, es := range ep.Subsets {
			for _, a := range es.Addresses {
				for _, p := range es.Ports {
					epMapKey := a.IP + ":" + fmt.Sprintf("%d", p.Port)
					if epInfo, ok := endpointMap[epMapKey]; !ok {
						endpointMap[epMapKey] = endpointInfo{
							Namespace:    ep.Namespace,
							PodName:      a.TargetRef.Name,
							ServiceNames: []string{ep.Name}}
					} else {
						svcs := append(epInfo.ServiceNames, ep.Name)
						epInfo.ServiceNames = svcs
						endpointMap[epMapKey] = epInfo
					}
				}
			}
		}
	}
	return endpointMap
}

func (m *svcAssociation) Vet(c kubernetes.Interface) ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}
	endpoints, err := util.ListEndpointsInMesh(c)
	if err != nil {
		if n := util.IstioInitializerDisabledNote(err.Error(), vetterId,
			multiple_service_association_note_type); n != nil {
			notes = append(notes, n)
			return notes, nil
		}
		return nil, err
	}

	epMap := createEndpointMap(endpoints)
	for _, v := range epMap {
		if len(v.ServiceNames) > 1 {
			notes = append(notes, &apiv1.Note{
				Type:    multiple_service_association_note_type,
				Summary: multiple_service_association_summary,
				Msg:     multiple_service_association_msg,
				Level:   apiv1.NoteLevel_ERROR,
				Attr: map[string]string{
					"pod_name":     v.PodName,
					"namespace":    v.Namespace,
					"service_list": strings.Join(v.ServiceNames, ", ")}})
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeId(notes[i])
	}

	return notes, nil
}

func (m *svcAssociation) Info() *apiv1.Info {
	return &m.info
}

// NewVetter returns "svcAssociation" which implements Vetter Interface
func NewVetter() *svcAssociation {
	return &svcAssociation{info: apiv1.Info{Id: vetterId, Version: "0.1.0"}}
}
