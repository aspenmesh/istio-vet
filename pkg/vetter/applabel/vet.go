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

package applabel

import (
	"k8s.io/client-go/kubernetes"

	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
)

const (
	vetterId                    = "applabel"
	missing_app_label_note_type = "missing-app-label"
	missing_app_label_summary   = "Missing app label - ${pod_name}"
	missing_app_label_msg       = "The pod ${pod_name} in namespace ${namespace}" +
		" is missing \"app\" label. Consider adding the label \"app\" to the pod."
)

type applabel struct {
	info apiv1.Info
}

func (m *applabel) Vet(c kubernetes.Interface) ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}

	pods, err := util.ListPodsInMesh(c)
	if err != nil {
		if n := util.IstioInitializerDisabledNote(err.Error(), vetterId,
			missing_app_label_note_type); n != nil {
			notes = append(notes, n)
			return notes, nil
		}
		return nil, err
	}
	for _, p := range pods {
		if _, ok := p.Labels[util.IstioAppLabel]; !ok {
			notes = append(notes, &apiv1.Note{
				Type:    missing_app_label_note_type,
				Summary: missing_app_label_summary,
				Msg:     missing_app_label_msg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"pod_name":  p.Name,
					"namespace": p.Namespace}})
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeId(notes[i])
	}

	return notes, nil
}

func (m *applabel) Info() *apiv1.Info {
	return &m.info
}

func NewVetter() *applabel {
	return &applabel{info: apiv1.Info{Id: vetterId, Version: "0.1.0"}}
}
