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

// Package applabel vets the labels defined for the pods in the mesh and
// generates notes if the label `app` is missing on any pod.
package applabel

import (
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	"k8s.io/client-go/listers/core/v1"
)

const (
	vetterID                = "AppLabel"
	missingAppLabelNoteType = "missing-app-label"
	missingAppLabelSummary  = "Missing app label - ${pod_name}"
	missingAppLabelMsg      = "The pod ${pod_name} in namespace ${namespace}" +
		" is missing \"app\" label. Consider adding the label \"app\" to the pod."
)

// AppLabel implements Vetter interface
type AppLabel struct {
	info      apiv1.Info
	nsLister  v1.NamespaceLister
	cmLister  v1.ConfigMapLister
	podLister v1.PodLister
}

// Vet returns the list of generated notes
func (m *AppLabel) Vet() ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}

	pods, err := util.ListPodsInMesh(m.nsLister, m.cmLister, m.podLister)
	if err != nil {
		if n := util.IstioInitializerDisabledNote(err.Error(), vetterID,
			missingAppLabelNoteType); n != nil {
			notes = append(notes, n)
			return notes, nil
		}
		return nil, err
	}
	for _, p := range pods {
		if _, ok := p.Labels[util.IstioAppLabel]; !ok {
			notes = append(notes, &apiv1.Note{
				Type:    missingAppLabelNoteType,
				Summary: missingAppLabelSummary,
				Msg:     missingAppLabelMsg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"pod_name":  p.Name,
					"namespace": p.Namespace}})
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}

	return notes, nil
}

func (m *AppLabel) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

// NewVetter returns "AppLabel" which implements Vetter Interface
func NewVetter(factory vetter.ResourceListGetter) *AppLabel {
	return &AppLabel{
		nsLister:  factory.K8s().Core().V1().Namespaces().Lister(),
		cmLister:  factory.K8s().Core().V1().ConfigMaps().Lister(),
		podLister: factory.K8s().Core().V1().Pods().Lister(),
	}
}
